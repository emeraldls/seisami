import { motion, AnimatePresence } from "framer-motion";
import {
  ChevronLeft,
  Home,
  Settings,
  FolderKanban,
  ClosedCaption,
  Cloud,
  Download,
  LogOut,
} from "lucide-react";
import { cn } from "~/lib/utils";
import { Link, useLocation } from "react-router-dom";
import { useSidebar } from "~/contexts/sidebar-context";
import { Button } from "./ui/button";
import { BoardSelector } from "./board-selector";
import { useEffect, useMemo, useState } from "react";
import { useDesktopAuthStore } from "~/stores/auth-store";
import { BoardMembersPanel } from "./board-members-panel";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import { InstallUpdate } from "../../wailsjs/go/main/App";
import { types } from "../../wailsjs/go/models";
import { CloudLoginDialog } from "./cloud-login-dialog";

interface NavItem {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  active?: boolean;
  badge?: string;
}

type DownloadProgressState = {
  percent: number | null;
  downloadedBytes: number;
  totalBytes: number;
};

export const Sidebar = () => {
  const navItems: NavItem[] = [
    { id: "/", label: "Home", icon: Home, active: true },
    { id: "/transcriptions", label: "Transcripts", icon: ClosedCaption },
    { id: "/boards", label: "Manage Boards", icon: FolderKanban },
  ];

  const bottomItems: NavItem[] = [
    { id: "settings", label: "Settings", icon: Settings },
  ];

  const { pathname } = useLocation();
  const { collapsed, toggleCollapsed } = useSidebar();
  const { isAuthenticated, logout, setError } =
    useDesktopAuthStore();
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [isDownloadingUpdate, setIsDownloadingUpdate] = useState(false);
  const [availableUpdate, setAvailableUpdate] =
    useState<types.AppVersion | null>(null);
  const [downloadProgress, setDownloadProgress] =
    useState<DownloadProgressState | null>(null);
  const [isLoginDialogOpen, setIsLoginDialogOpen] = useState(false);

  useEffect(() => {
    const unsubscribeUpdateAvailable = EventsOn(
      "update:available",
      (data: any) => {
        if (!data) {
          return;
        }

        setAvailableUpdate(types.AppVersion.createFrom(data));
      }
    );

    const unsubscribeDownloadStarted = EventsOn(
      "update:download_started",
      (data: any) => {
        const totalBytes =
          typeof data?.totalBytes === "number" ? data.totalBytes : -1;
        setIsDownloadingUpdate(true);
        setDownloadProgress({
          percent: totalBytes > 0 ? 0 : null,
          downloadedBytes: 0,
          totalBytes,
        });
      }
    );

    const unsubscribeDownloadProgress = EventsOn(
      "update:download_progress",
      (data: any) => {
        const downloadedBytes =
          typeof data?.downloadedBytes === "number" ? data.downloadedBytes : 0;
        const totalBytes =
          typeof data?.totalBytes === "number" ? data.totalBytes : -1;
        const percentValue =
          typeof data?.percent === "number" ? data.percent : null;

        setDownloadProgress({
          percent:
            totalBytes > 0 && percentValue !== null ? percentValue : null,
          downloadedBytes,
          totalBytes,
        });
      }
    );

    const unsubscribeDownloadComplete = EventsOn(
      "update:download_complete",
      (data: any) => {
        setDownloadProgress((prev) =>
          prev
            ? {
                percent: prev.totalBytes > 0 ? 100 : prev.percent,
                downloadedBytes:
                  typeof data?.downloadedBytes === "number"
                    ? data.downloadedBytes
                    : prev.downloadedBytes,
                totalBytes:
                  typeof data?.totalBytes === "number" && data.totalBytes >= 0
                    ? data.totalBytes
                    : prev.totalBytes,
              }
            : prev
        );
      }
    );

    const unsubscribeDownloadError = EventsOn(
      "update:download_error",
      (message: string) => {
        console.error("Update download error:", message);
        setDownloadProgress(null);
        setIsDownloadingUpdate(false);
      }
    );

    return () => {
      unsubscribeUpdateAvailable();
      unsubscribeDownloadStarted();
      unsubscribeDownloadProgress();
      unsubscribeDownloadComplete();
      unsubscribeDownloadError();
    };
  }, []);

  const handleCloudLogin = () => {
    setError(null);
    setIsLoginDialogOpen(true);
  };

  const handleLogout = () => {
    logout();
  };

  const handleDownloadUpdate = async () => {
    if (!availableUpdate) {
      return;
    }

    setIsDownloadingUpdate(true);
    setDownloadProgress({ percent: 0, downloadedBytes: 0, totalBytes: -1 });
    try {
      await InstallUpdate(availableUpdate);
      setAvailableUpdate(null);
    } catch (error) {
      console.error("Failed to download update:", error);
    } finally {
      setIsDownloadingUpdate(false);
      setDownloadProgress(null);
    }
  };

  const downloadLabel = useMemo(() => {
    if (!availableUpdate) {
      return "";
    }

    if (!isDownloadingUpdate) {
      return `Download ${availableUpdate.version}`;
    }

    if (!downloadProgress) {
      return "Downloading update...";
    }

    const downloadedMB = downloadProgress.downloadedBytes / 1024 / 1024;

    if (downloadProgress.totalBytes > 0) {
      const totalMB = downloadProgress.totalBytes / 1024 / 1024;
      return `Downloading ${downloadedMB.toFixed(2)} / ${totalMB.toFixed(
        2
      )} MB`;
    }

    return `Downloading ${downloadedMB.toFixed(2)} MB`;
  }, [availableUpdate, downloadProgress, isDownloadingUpdate]);

  const downloadFillPercent = useMemo(() => {
    if (!downloadProgress) {
      return null;
    }

    if (downloadProgress.percent !== null) {
      return Math.min(100, Math.max(0, downloadProgress.percent));
    }

    if (downloadProgress.totalBytes > 0) {
      const ratio =
        downloadProgress.downloadedBytes /
        Math.max(downloadProgress.totalBytes, 1);
      return Math.min(100, Math.max(0, ratio * 100));
    }

    return null;
  }, [downloadProgress]);

  return (
    <>
      <AnimatePresence>
        <motion.aside
        animate={{ width: collapsed ? 72 : 256 }}
        transition={{ type: "spring", stiffness: 300, damping: 30 }}
        className={cn(
          "bg-sidebar fixed border-r border-sidebar-border flex flex-col h-screen z-20",
          collapsed ? "w-18" : "w-64"
        )}
        style={{ width: collapsed ? 72 : 256 }}
      >
        <div className="p-3 flex items-center justify-between">
          {/* <img
            src={Logo}
            alt="logo"
            className={cn(
              "transition-all duration-300",
              collapsed ? "w-8 -left-0" : "w-fit relative -left-7"
            )}
            style={{ maxWidth: collapsed ? 32 : undefined }}
          /> */}
          Logo
          <Button
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
            onClick={toggleCollapsed}
            variant={"ghost"}
          >
            <motion.span
              initial={false}
              animate={{ rotate: collapsed ? 180 : 0 }}
              transition={{ type: "spring", stiffness: 300, damping: 30 }}
              className="inline-block"
            >
              <ChevronLeft />
            </motion.span>
          </Button>
        </div>

        {!collapsed && (
          <div className="px-3 pb-2">
            <BoardSelector className="w-full" />
          </div>
        )}

        <nav className={cn("flex-1 px-3", collapsed && "px-1")}>
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = pathname === item.id;

            return (
              <Link
                to={item.id}
                key={item.id}
                className={cn(
                  "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg mb-1",
                  "transition-all duration-200",
                  isActive
                    ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium"
                    : "text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground",
                  collapsed && "justify-center px-0"
                )}
              >
                <Icon className="h-5 w-5" />
                {!collapsed && <span className="text-sm">{item.label}</span>}
              </Link>
            );
          })}
        </nav>

        <div
          className={cn(
            "px-3 pb-4 border-t border-sidebar-border pt-2 flex flex-col gap-2",
            collapsed && "px-1"
          )}
        >
          {availableUpdate && (
            <Button
              onClick={handleDownloadUpdate}
              disabled={isDownloadingUpdate}
              className={cn(
                "relative w-full overflow-hidden transition-colors",
                isDownloadingUpdate ? "bg-blue-500/10" : undefined
              )}
              size="sm"
            >
              {isDownloadingUpdate && downloadFillPercent !== null && (
                <span
                  className="pointer-events-none absolute inset-y-0 left-0 bg-blue-500/70 transition-all duration-300 ease-out"
                  style={{ width: `${downloadFillPercent}%` }}
                  aria-hidden
                />
              )}
              {isDownloadingUpdate && downloadFillPercent === null && (
                <span
                  className="pointer-events-none absolute inset-0 bg-gradient-to-r from-transparent via-blue-400/40 to-transparent animate-pulse"
                  aria-hidden
                />
              )}
              <Download
                className={cn(
                  "relative z-10 h-4 w-4",
                  isDownloadingUpdate ? "text-black" : undefined
                )}
              />
              {!collapsed && (
                <span
                  className={cn(
                    "relative z-10 text-xs",
                    isDownloadingUpdate ? "text-black" : undefined
                  )}
                >
                  {downloadLabel}
                </span>
              )}
            </Button>
          )}
          {isAuthenticated ? (
            <div className="space-y-2">
              {!collapsed && <BoardMembersPanel key={"board-panel"} />}
            </div>
          ) : (
            <Button
              onClick={handleCloudLogin}
              disabled={isAuthenticating}
              className="w-full"
              size="sm"
            >
              <Cloud className="h-4 w-4" />
              {!collapsed && (
                <span className="text-xs" key={"auth"}>
                  {isAuthenticating ? "Authenticating..." : "Cloud Features"}
                </span>
              )}
            </Button>
          )}

          {bottomItems.map((item) => {
            const Icon = item.icon;

            return (
              <div key={item.id}>
                <Link
                  to={"/" + item.id}
                  className={cn(
                    "w-full flex items-center gap-3 px-2 py-2.5 rounded-lg mb-1",
                    "transition-all duration-200",
                    "text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground",
                    collapsed && "justify-center px-0"
                  )}
                >
                  <Icon className="h-5 w-5" />
                  {!collapsed && <span className="text-sm">{item.label}</span>}
                </Link>
              </div>
            );
          })}
          {isAuthenticated && (
            <div className="space-y-2">
              <Button
                onClick={handleLogout}
                variant="ghost"
                size="sm"
                className="w-full justify-start text-red-600"
              >
                <LogOut className="h-4 w-4" />
                {!collapsed && <span className="">Logout</span>}
              </Button>
            </div>
          )}
        </div>
        </motion.aside>
      </AnimatePresence>
      <CloudLoginDialog
        open={isLoginDialogOpen}
        onOpenChange={(open) => {
          if (!open) {
            setIsAuthenticating(false);
          }
          setIsLoginDialogOpen(open);
        }}
        onAuthStart={() => {
          setIsAuthenticating(true);
        }}
        onAuthEnd={() => {
          setIsAuthenticating(false);
        }}
      />
    </>
  );
};
