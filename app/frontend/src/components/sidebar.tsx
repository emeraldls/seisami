import { motion, AnimatePresence } from "framer-motion";
import {
  ChevronLeft,
  Home,
  Settings,
  FolderKanban,
  ClosedCaption,
  Cloud,
} from "lucide-react";
import { cn } from "~/lib/utils";
import { Link, useLocation } from "react-router-dom";
import { useSidebar } from "~/contexts/sidebar-context";
import { Button } from "./ui/button";
import { BoardSelector } from "./board-selector";
import { useState } from "react";
import { useDesktopAuthStore } from "~/stores/auth-store";
import { DesktopAuthService } from "~/lib/desktop-auth-service";

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
  const { isAuthenticated, email, setToken, logout, setLoading, setError } =
    useDesktopAuthStore();
  const [isAuthenticating, setIsAuthenticating] = useState(false);

  const handleCloudLogin = async () => {
    setIsAuthenticating(true);
    setLoading(true);
    try {
      await DesktopAuthService.startAuthFlow();
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Authentication failed";
      setError(message);
      console.error("Cloud login error:", error);
    } finally {
      setIsAuthenticating(false);
      setLoading(false);
    }
  };

  const handleLogout = () => {
    logout();
  };

  return (
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
          {isAuthenticated ? (
            <div className="space-y-2">
              {!collapsed && (
                <div className="px-3 py-2 rounded-lg bg-sidebar-accent/30">
                  <p className="text-xs font-medium text-sidebar-foreground/70">
                    Cloud Enabled
                  </p>
                  <p className="text-xs text-sidebar-foreground/50 truncate">
                    {email}
                  </p>
                </div>
              )}
              <Button
                onClick={handleLogout}
                variant="ghost"
                size="sm"
                className="w-full justify-center"
              >
                <Cloud className="h-4 w-4" />
                {!collapsed && <span className="text-xs">Logout</span>}
              </Button>
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
                <span className="text-xs">
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
                    "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg mb-1",
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
        </div>
      </motion.aside>
    </AnimatePresence>
  );
};
