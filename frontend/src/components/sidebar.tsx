import { motion, AnimatePresence } from "framer-motion";
import {
  ChevronLeft,
  Home,
  Settings,
  FolderKanban,
  ClosedCaption,
} from "lucide-react";
import { cn } from "~/lib/utils";
import { Link, useLocation } from "react-router-dom";
import { useSidebar } from "~/contexts/sidebar-context";
import { Button } from "./ui/button";
import { BoardSelector } from "./board-selector";

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

        {/* Board Selector */}
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
            "px-3 pb-4 border-t border-sidebar-border pt-4 flex flex-col gap-2",
            collapsed && "px-1"
          )}
        >
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
