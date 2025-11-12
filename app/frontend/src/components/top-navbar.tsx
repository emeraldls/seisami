import { NotificationsPopover } from "./notifications-popover";
import { ConnectedUsers } from "./connected-users";
import { useBoardStore } from "~/stores/board-store";
import { useSidebar } from "~/contexts/sidebar-context";

export const TopNavbar = () => {
  const { currentBoard } = useBoardStore();
  const { collapsed } = useSidebar();

  console.log("Current Board", currentBoard);

  return (
    <div
      className="fixed top-0 right-0 z-40 h-14 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 transition-all duration-300"
      style={{ left: collapsed ? 72 : 256 }}
    >
      <div className="flex h-full items-center justify-between px-6">
        <div className="flex-1" />

        <div className="flex items-center gap-4">
          {currentBoard && <ConnectedUsers boardId={currentBoard.id} />}
          <NotificationsPopover />
        </div>
      </div>
    </div>
  );
};
