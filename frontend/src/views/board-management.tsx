import { useState, useEffect } from "react";
import { Plus, Edit3, Trash2, Calendar } from "lucide-react";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
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
    <div className="h-full flex flex-col bg-gray-50/50">
      <div className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="max-w-6xl mx-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Boards</h1>
          <p className="text-sm text-gray-500 mt-1">
            Manage and organize your project boards
          </p>
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        <div className="max-w-6xl mx-auto p-6">
          {isLoading && boards.length === 0 ? (
            <div className="flex items-center justify-center py-20">
              <div className="text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
                <p className="text-gray-500">Loading boards...</p>
              </div>
            </div>
          ) : boards.length === 0 ? (
            <div className="flex items-center justify-center py-20">
              <div className="text-center max-w-md">
                <div className="w-24 h-24 mx-auto bg-gray-100 rounded-lg flex items-center justify-center mb-6">
                  <Plus className="h-8 w-8 text-gray-400" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-2">
                  No boards yet
                </h3>
                <p className="text-gray-500 mb-8">
                  Create your first board to start organizing your tasks and
                  projects
                </p>

                <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-sm">
                  <h4 className="text-lg font-medium text-gray-900 mb-4">
                    Create your first board
                  </h4>
                  <form onSubmit={handleCreateBoard} className="space-y-4">
                    <Input
                      type="text"
                      placeholder="Enter board name..."
                      value={newBoardName}
                      onChange={(e) => setNewBoardName(e.target.value)}
                      disabled={isCreating}
                      className="w-full"
                    />
                    <Button
                      type="submit"
                      disabled={!newBoardName.trim() || isCreating}
                      className="w-full gap-2"
                    >
                      <Plus className="h-4 w-4" />
                      {isCreating ? "Creating..." : "Create Board"}
                    </Button>
                  </form>
                </div>
              </div>
            </div>
          ) : (
            <div className="space-y-6">
              <div className="bg-white border border-gray-200 rounded-sm p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">
                  Create New Board
                </h2>
                <form onSubmit={handleCreateBoard} className="flex gap-3">
                  <Input
                    type="text"
                    placeholder="Enter board name..."
                    value={newBoardName}
                    onChange={(e) => setNewBoardName(e.target.value)}
                    disabled={isCreating}
                    className="flex-1 outline-none focus:outline-none focus:ring-0"
                  />
                  <Button
                    type="submit"
                    disabled={!newBoardName.trim() || isCreating}
                    className="gap-2 px-6"
                  >
                    <Plus className="h-4 w-4" />
                    {isCreating ? "Creating..." : "Create"}
                  </Button>
                </form>
              </div>

              <div>
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-gray-900">
                    Your Boards ({boards.length})
                  </h2>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {boards.map((board) => (
                    <div
                      key={board.id}
                      className={`group bg-white border rounded-sm p-4 shadow-sm hover:shadow-md transition-all cursor-pointer ${
                        currentBoard?.id === board.id
                          ? "border-black bg-blue-50/50 ring-1 ring-black"
                          : "border-gray-200 hover:border-gray-300"
                      }`}
                    >
                      <div className="flex items-start justify-between mb-3">
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
                              className="font-medium text-gray-900"
                              autoFocus
                              onBlur={() => setEditingBoard(null)}
                            />
                          </form>
                        ) : (
                          <div
                            className="flex-1 min-w-0"
                            onClick={() => handleSelectBoard(board)}
                          >
                            <h3 className="font-medium text-gray-900 truncate group-hover:text-blue-600 transition-colors">
                              {board.name}
                            </h3>
                            {currentBoard?.id === board.id && (
                              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-700 mt-1">
                                Current
                              </span>
                            )}
                          </div>
                        )}

                        {/* Actions */}
                        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              setEditingBoard({
                                id: board.id,
                                name: board.name,
                              });
                            }}
                            className="h-8 w-8 p-0 text-gray-400 hover:text-gray-600"
                          >
                            <Edit3 className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteBoard(board.id, board.name);
                            }}
                            className="h-8 w-8 p-0 text-gray-400 hover:text-red-600"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>

                      <div
                        className="space-y-2"
                        onClick={() => handleSelectBoard(board)}
                      >
                        <div className="flex items-center text-sm text-gray-500">
                          <Calendar className="h-4 w-4 mr-2" />
                          Created{" "}
                          {new Date(board.created_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

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
