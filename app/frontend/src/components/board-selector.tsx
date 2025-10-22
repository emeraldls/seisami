import { useState, useEffect } from "react";
import {
  Check,
  ChevronDown,
  Plus,
  Edit3,
  Trash2,
  MoreHorizontal,
} from "lucide-react";
import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "~/components/ui/alert-dialog";
import { useBoardStore, NormalizedBoard } from "~/stores/board-store";

interface BoardSelectorProps {
  className?: string;
}

export function BoardSelector({ className }: BoardSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [newBoardName, setNewBoardName] = useState("");
  const [editingBoard, setEditingBoard] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [boardToDelete, setBoardToDelete] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [warningDialogOpen, setWarningDialogOpen] = useState(false);

  const {
    boards,
    currentBoard,
    fetchBoards,
    selectBoard,
    createBoard,
    updateBoard,
    deleteBoard,
    isLoading,
  } = useBoardStore();

  useEffect(() => {
    fetchBoards();
  }, [fetchBoards]);

  const handleSelectBoard = async (board: NormalizedBoard) => {
    await selectBoard(board.id);
    setIsOpen(false);
  };

  const handleCreateBoard = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newBoardName.trim()) return;

    setIsCreating(true);
    const board = await createBoard(newBoardName.trim());
    if (board) {
      setNewBoardName("");
      setIsOpen(false);
    }
    setIsCreating(false);
  };

  const handleEditBoard = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingBoard || !editingBoard.name.trim()) return;

    const success = await updateBoard(
      editingBoard.id,
      editingBoard.name.trim()
    );
    if (success) {
      setEditingBoard(null);
    }
  };

  const handleDeleteBoard = (boardId: string, boardName: string) => {
    if (boards.length <= 1) {
      setWarningDialogOpen(true);
      return;
    }

    setBoardToDelete({ id: boardId, name: boardName });
    setDeleteDialogOpen(true);
  };

  const confirmDeleteBoard = async () => {
    if (boardToDelete) {
      await deleteBoard(boardToDelete.id);
      setDeleteDialogOpen(false);
      setBoardToDelete(null);
    }
  };

  return (
    <>
      <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            className={`justify-between h-auto p-3 ${className}`}
            disabled={isLoading}
          >
            <div className="flex flex-col items-start">
              <span className="text-xs text-muted-foreground">
                Current Board
              </span>
              <span className="font-medium text-sm">
                {currentBoard?.name || "Select Board"}
              </span>
            </div>
            <ChevronDown className="h-4 w-4 ml-2" />
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent className="w-80" align="start">
          {boards.length > 0 && (
            <>
              {boards.map((board) => (
                <div key={board.id} className="flex items-center group">
                  {editingBoard?.id === board.id ? (
                    <form
                      onSubmit={handleEditBoard}
                      className="flex-1 flex items-center p-2 space-x-2"
                    >
                      <input
                        type="text"
                        value={editingBoard.name}
                        onChange={(e) =>
                          setEditingBoard({
                            ...editingBoard,
                            name: e.target.value,
                          })
                        }
                        className="flex-1 px-2 py-1 text-sm border rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        autoFocus
                        onBlur={() => setEditingBoard(null)}
                      />
                      <Button type="submit" size="sm" className="h-6 px-2">
                        Save
                      </Button>
                    </form>
                  ) : (
                    <>
                      <DropdownMenuItem
                        onClick={() => handleSelectBoard(board)}
                        className="flex-1 flex items-center justify-between py-2"
                      >
                        <div className="flex flex-col">
                          <span className="font-medium">{board.name}</span>
                          <span className="text-xs text-muted-foreground">
                            {new Date(board.created_at).toLocaleDateString()}
                          </span>
                        </div>
                        {currentBoard?.id === board.id && (
                          <Check className="h-4 w-4 text-green-600" />
                        )}
                      </DropdownMenuItem>

                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0 opacity-0 group-hover:opacity-100"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={(e) => {
                              e.stopPropagation();
                              setEditingBoard({
                                id: board.id,
                                name: board.name,
                              });
                            }}
                          >
                            <Edit3 className="h-4 w-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteBoard(board.id, board.name);
                            }}
                            className="text-red-600 focus:text-red-600"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </>
                  )}
                </div>
              ))}
              <DropdownMenuSeparator />
            </>
          )}

          {/* Quick Create Form */}
          <div className="p-2">
            <form
              onSubmit={handleCreateBoard}
              className="flex flex-col space-y-2"
            >
              <input
                type="text"
                placeholder="New board name..."
                value={newBoardName}
                onChange={(e) => setNewBoardName(e.target.value)}
                className="px-3 py-2 text-sm border rounded-md focus:outline-none focus:ring-1"
                disabled={isCreating}
              />
              <Button
                type="submit"
                size="sm"
                disabled={!newBoardName.trim() || isCreating}
                variant={"outline"}
              >
                <Plus className="h-3 w-3 mr-1" />
                {isCreating ? "Creating..." : "Create Board"}
              </Button>
            </form>
          </div>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Board</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{boardToDelete?.name}"? This
              action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDeleteBoard}
              className="bg-red-600 hover:bg-red-700"
            >
              Delete Board
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Warning Dialog for Single Board */}
      <AlertDialog open={warningDialogOpen} onOpenChange={setWarningDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Cannot Delete Board</AlertDialogTitle>
            <AlertDialogDescription>
              You cannot delete your only board. Create another board first
              before deleting this one.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={() => setWarningDialogOpen(false)}>
              OK
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
