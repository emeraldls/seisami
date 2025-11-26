import { motion, AnimatePresence } from "framer-motion";
import {
  ChevronLeft,
  Home,
  Settings,
  FolderKanban,
  ClosedCaption,
  Cloud,
  LogOut,
  Search,
} from "lucide-react";
import { cn } from "~/lib/utils";
import { Link, useLocation } from "react-router-dom";
import { useSidebar } from "~/contexts/sidebar-context";
import { Button } from "./ui/button";
import { BoardSelector } from "./board-selector";
import { useState } from "react";
import { useDesktopAuthStore } from "~/stores/auth-store";
import { BoardMembersPanel } from "./board-members-panel";

import { CloudLoginDialog } from "./cloud-login-dialog";
import { VersionUpdate } from "./version-update";
import { useCommandPalette } from "~/contexts/command-palette-context";
import { LogoutConfirmationDialog } from "./logout-confirmation-dialog";
import { Logo } from "./logo";
import { useCollaborationStore } from "~/stores/collab-store";
import { DesktopAuthService } from "~/lib/desktop-auth-service";
import { toast } from "sonner";

interface NavItem {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  active?: boolean;
  badge?: string;
}

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
  const {
    isAuthenticated,
    logout,
    setError,
    clearError,
    setLoading,
    error: storeError,
  } = useDesktopAuthStore();
  const { open: openCommandPalette } = useCommandPalette();
  const [isAuthenticating, setIsAuthenticating] = useState(false);

  const [isLogoutDialogOpen, setIsLogoutDialogOpen] = useState(false);

  const { teardown } = useCollaborationStore();

  const handleLogoutClick = () => {
    setIsLogoutDialogOpen(true);
  };

  const handleLogoutConfirm = (clearLocalData: boolean) => {
    logout(clearLocalData);
    teardown();
    setIsLogoutDialogOpen(false);

    window.location.reload();
  };

  const handleDesktopFlow = async () => {
    clearError();
    setLoading(true);
    setIsAuthenticating(true);

    try {
      await DesktopAuthService.startAuthFlow();
    } catch (error) {
      let message = "Failed to start desktop login flow";
      if (error instanceof Error) {
        message = error.message;
      }

      setError(message);
      toast.error(message);
    } finally {
      setLoading(false);
      setIsAuthenticating(false);
    }
  };

  return (
    <>
      <AnimatePresence>
        <motion.aside
          animate={{ width: collapsed ? 72 : 256 }}
          transition={{ type: "spring", stiffness: 300, damping: 30 }}
          className={cn(
            "bg-sidebar fixed border-r border-sidebar-border flex flex-col h-screen z-50",
            collapsed ? "w-18" : "w-64"
          )}
          style={{ width: collapsed ? 72 : 256 }}
        >
          <div className="p-3 flex items-center justify-between gap-2">
            <Link
              to="/"
              className={cn(
                "flex items-center gap-3 rounded-xl px-1 py-1 transition-all",
                collapsed ? "justify-center" : "hover:bg-sidebar-accent/30"
              )}
            >
              <Logo
                className={cn(
                  "h-10 w-10 transition-transform",
                  collapsed ? "scale-90" : "scale-100"
                )}
              />
              {!collapsed && (
                <div className="flex flex-col leading-tight">
                  <span className="text-base font-semibold text-sidebar-foreground">
                    Seisami
                  </span>
                  <span className="text-[10px] text-sidebar-foreground/60">
                    Capture • Sync • Share
                  </span>
                </div>
              )}
            </Link>
            <Button
              aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
              onClick={toggleCollapsed}
              variant="ghost"
              size="icon"
              className={cn(
                "h-9 w-9 rounded-full border border-sidebar-border/60 bg-sidebar/40",
                "text-sidebar-foreground shadow-sm transition-colors",
                "hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground"
              )}
            >
              <motion.span
                initial={false}
                animate={{ rotate: collapsed ? 180 : 0 }}
                transition={{ type: "spring", stiffness: 300, damping: 30 }}
                className="inline-flex items-center justify-center"
              >
                <ChevronLeft className="h-4 w-4" />
              </motion.span>
            </Button>
          </div>

          {!collapsed && (
            <div className="px-3 pb-2">
              <BoardSelector className="w-full" />
            </div>
          )}

          {!collapsed && (
            <div className="px-3 pb-2">
              <Button
                onClick={openCommandPalette}
                variant="outline"
                className="w-full justify-start gap-2 text-muted-foreground"
                size="sm"
              >
                <Search className="h-4 w-4" />
                <span className="flex-1 text-left">Search...</span>
                <kbd className="hidden sm:inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium">
                  ⌘K
                </kbd>
              </Button>
            </div>
          )}

          {collapsed && (
            <div className="px-1 pb-2">
              <Button
                onClick={openCommandPalette}
                variant="ghost"
                className="w-full justify-center"
                size="sm"
              >
                <Search className="h-5 w-5" />
              </Button>
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
            <VersionUpdate collapsed={collapsed} />
            {isAuthenticated ? (
              <div className="space-y-2">
                {!collapsed && <BoardMembersPanel key={"board-panel"} />}
              </div>
            ) : (
              <Button
                onClick={handleDesktopFlow}
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
                    {!collapsed && (
                      <span className="text-sm">{item.label}</span>
                    )}
                  </Link>
                </div>
              );
            })}
            {isAuthenticated && (
              <div className="space-y-2">
                <Button
                  onClick={handleLogoutClick}
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
      <LogoutConfirmationDialog
        open={isLogoutDialogOpen}
        onOpenChange={setIsLogoutDialogOpen}
        onConfirm={handleLogoutConfirm}
      />
    </>
  );
};
