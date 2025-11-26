import { NotificationsPopover } from "./notifications-popover";
import { ConnectedUsers } from "./connected-users";
import { useBoardStore } from "~/stores/board-store";
import { useSidebar } from "~/contexts/sidebar-context";
import { useDesktopAuthStore } from "~/stores/auth-store";

export const TopNavbar = () => {
  const { currentBoard } = useBoardStore();
  const { collapsed } = useSidebar();
  const { isAuthenticated } = useDesktopAuthStore();

  return (
    <div
      className="fixed top-0 right-0 z-40 h-14 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 transition-all duration-300"
      style={{ left: collapsed ? 72 : 256 }}
    >
      <div className="flex h-full items-center justify-between px-6">
        <div className={`flex-1 ${collapsed && "ml-5"}`}>
          <h2 className="text-xl">
            Hold{" "}
            <kbd className="px-2 bg-black/70 text-white rounded text-sm font-mono">
              Fn
            </kbd>{" "}
            to dictate anywhere!
          </h2>
        </div>

        <div className="flex items-center gap-4">
          {currentBoard && <ConnectedUsers boardId={currentBoard.id} />}
          {isAuthenticated && <NotificationsPopover />}
        </div>
      </div>
    </div>
  );
};
