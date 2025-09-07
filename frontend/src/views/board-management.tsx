import { useState, useEffect } from "react";
import { Plus, Edit3, Trash2, Calendar, Clock, SquarePen } from "lucide-react";
import { Button } from "~/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
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
import { useNavigate } from "react-router-dom";

export default function BoardManagement() {
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
    createBoard,
    updateBoard,
    deleteBoard,
    selectBoard,
    isLoading,
  } = useBoardStore();

  const navigate = useNavigate();

  useEffect(() => {
    fetchBoards();
  }, [fetchBoards]);

  const handleCreateBoard = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newBoardName.trim()) return;

    setIsCreating(true);
    const board = await createBoard(newBoardName.trim());
    if (board) {
      setNewBoardName("");
    }
    setIsCreating(false);
  };

  const handleUpdateBoard = async (e: React.FormEvent) => {
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
    console.log("deleting board...");
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

  const handleSelectBoard = async (board: NormalizedBoard) => {
    await selectBoard(board.id);
    navigate("/");
  };

  return (
    <div className="flex-1 flex flex-col">
      <header className="bg-white border-b px-8 py-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">Board Management</h1>
            <p className="text-sm text-neutral-500 mt-1">
              Manage your boards and workspaces
            </p>
          </div>
        </div>
      </header>

      <div className="flex-1 p-8 overflow-auto">
        <div className="mb-8">
          <Card className="max-w-md mx-auto rounded-none shadow-none">
            <CardHeader>
              <CardTitle className="text-lg">Create New Board</CardTitle>
              <CardDescription>
                Add a new board to organize your tasks and ideas
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleCreateBoard} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="boardName">Board Name</Label>
                  <Input
                    id="boardName"
                    type="text"
                    placeholder="e.g., Personal Tasks, Work Projects..."
                    value={newBoardName}
                    onChange={(e) => setNewBoardName(e.target.value)}
                    disabled={isCreating}
                  />
                </div>
                <Button
                  type="submit"
                  disabled={!newBoardName.trim() || isCreating}
                  className="w-full"
                >
                  <Plus className="h-4 w-4 mr-2" />
                  {isCreating ? "Creating..." : "Create Board"}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>

        <div>
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold">
              Your Boards ({boards.length})
            </h2>
          </div>

          {isLoading && boards.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
            </div>
          ) : boards.length === 0 ? (
            <div className="text-center py-12">
              <div className="mx-auto h-16 w-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
                <Plus className="h-8 w-8 text-gray-400" />
              </div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                No boards yet
              </h3>
              <p className="text-gray-500 mb-4">
                Create your first board to get started
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {boards.map((board) => (
                <Card
                  key={board.id}
                  className={`cursor-pointer px-0 py-6 w-full rounded-none shadow-none transition-all  ${
                    currentBoard?.id === board.id ? "border" : ""
                  }`}
                >
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      {editingBoard?.id === board.id ? (
                        <form
                          onSubmit={handleUpdateBoard}
                          className="flex-1 mr-2"
                        >
                          <Input
                            value={editingBoard.name}
                            onChange={(e) =>
                              setEditingBoard({
                                ...editingBoard,
                                name: e.target.value,
                              })
                            }
                            className="text-lg font-semibold"
                            autoFocus
                            onBlur={() => setEditingBoard(null)}
                          />
                        </form>
                      ) : (
                        <CardTitle
                          className="text-base cursor-pointer"
                          onClick={() => handleSelectBoard(board)}
                        >
                          <span className="truncate line-clamp-1">
                            {board.name}
                          </span>
                          {currentBoard?.id === board.id && (
                            <span className="ml-2 text-xs bg-indigo-100 text-indigo-700 px-2 py-1 rounded-full">
                              Current
                            </span>
                          )}
                        </CardTitle>
                      )}

                      <div className="flex space-x-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            setEditingBoard({ id: board.id, name: board.name });
                          }}
                          className="h-8 w-8 p-0"
                        >
                          <SquarePen className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteBoard(board.id, board.name);
                          }}
                          className="h-8 w-8 p-0 text-red-600 hover:text-red-700"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  </CardHeader>

                  <CardContent
                    className="pt-0"
                    onClick={() => handleSelectBoard(board)}
                  >
                    <div className="flex items-center space-x-4 text-sm text-gray-500">
                      <div className="flex items-center">
                        <Calendar className="h-4 w-4 mr-1" />
                        {new Date(board.created_at).toLocaleDateString()}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </div>

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

      {/* Warning Dialog for Last Board */}
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
              Got it
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
