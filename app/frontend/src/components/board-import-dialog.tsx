import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Button } from "./ui/button";
import {
  Download,
  Loader2,
  FolderKanban,
  Columns3,
  CreditCard,
  ClosedCaption,
  CheckCircle2,
  AlertCircle,
} from "lucide-react";
import { useBoardStore } from "~/stores/board-store";
import { toast } from "sonner";
import { ApiClient, BoardMetadata } from "~/lib/api-client";
import { ImportNewBoard } from "../../wailsjs/go/main/App";

interface BoardImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  boardId?: string;
}

export const BoardImportDialog = ({
  open,
  onOpenChange,
  boardId,
}: BoardImportDialogProps) => {
  const [isLoadingMetadata, setIsLoadingMetadata] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [metadata, setMetadata] = useState<BoardMetadata | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { fetchBoards } = useBoardStore();

  useEffect(() => {
    if (open && boardId) {
      fetchBoardMetadata(boardId);
    }
  }, [open, boardId]);

  const fetchBoardMetadata = async (id: string) => {
    setIsLoadingMetadata(true);
    setError(null);
    setMetadata(null);

    try {
      const response = await ApiClient.getBoardMetadata(id);
      setMetadata(response.data);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch board metadata";
      setError(message);
      toast.error(message);
    } finally {
      setIsLoadingMetadata(false);
    }
  };

  const handleImport = async () => {
    if (!metadata) return;

    setIsImporting(true);
    try {
      // Call the Go backend to import the board
      await ImportNewBoard(metadata.id);

      toast.success("Board imported successfully!");

      // Refresh boards list
      await fetchBoards();

      // Close dialog
      onOpenChange(false);

      // Reset state
      setMetadata(null);
      setError(null);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to import board";
      setError(message);
      toast.error(message);
    } finally {
      setIsImporting(false);
    }
  };

  const handleClose = () => {
    if (!isImporting) {
      onOpenChange(false);
      // Reset state when closing
      setTimeout(() => {
        setMetadata(null);
        setError(null);
        setIsLoadingMetadata(false);
      }, 200);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      month: "long",
      day: "numeric",
      year: "numeric",
    });
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Download className="h-5 w-5" />
            Import Board
          </DialogTitle>
          <DialogDescription>
            {isLoadingMetadata
              ? "Fetching board information..."
              : "Review the board details before importing"}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Loading State */}
          {isLoadingMetadata && (
            <div className="flex flex-col items-center justify-center py-12 space-y-4">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
              <div className="text-center space-y-1">
                <p className="text-sm font-medium">Loading board metadata...</p>
                <p className="text-xs text-muted-foreground">
                  This will only take a moment
                </p>
              </div>
            </div>
          )}

          {/* Error State */}
          {error && !isLoadingMetadata && (
            <div className="flex flex-col items-center justify-center py-12 space-y-4">
              <div className="rounded-full bg-destructive/10 p-3">
                <AlertCircle className="h-8 w-8 text-destructive" />
              </div>
              <div className="text-center space-y-1">
                <p className="text-sm font-medium">Failed to load board</p>
                <p className="text-xs text-muted-foreground">{error}</p>
              </div>
              <Button
                onClick={() => boardId && fetchBoardMetadata(boardId)}
                variant="outline"
                size="sm"
              >
                Try Again
              </Button>
            </div>
          )}

          {/* Metadata Display */}
          {metadata && !isLoadingMetadata && (
            <div className="space-y-4">
              {/* Board Info */}
              <div className="rounded-lg border p-4 space-y-3 bg-muted/30">
                <div className="flex items-start gap-3">
                  <div className="rounded-lg bg-primary/10 p-2.5">
                    <FolderKanban className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1 space-y-1">
                    <h3 className="font-semibold text-lg">{metadata.name}</h3>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground pt-1">
                      <span>Created {formatDate(metadata.created_at)}</span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Import Summary */}
              <div className="space-y-2">
                <p className="text-sm font-medium text-muted-foreground">
                  This board will import:
                </p>
                <div className="space-y-2">
                  <div className="flex items-center gap-3 p-3 rounded-lg border bg-card">
                    <div className="rounded-md bg-blue-500/10 p-2">
                      <Columns3 className="h-4 w-4 text-blue-600" />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-medium">Columns</p>
                      <p className="text-xs text-muted-foreground">
                        Board structure and workflow
                      </p>
                    </div>
                    <span className="text-sm font-semibold text-blue-600">
                      {metadata.columns_count}
                    </span>
                  </div>

                  <div className="flex items-center gap-3 p-3 rounded-lg border bg-card">
                    <div className="rounded-md bg-green-500/10 p-2">
                      <CreditCard className="h-4 w-4 text-green-600" />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-medium">Cards</p>
                      <p className="text-xs text-muted-foreground">
                        Tasks and action items
                      </p>
                    </div>
                    <span className="text-sm font-semibold text-green-600">
                      {metadata.cards_count}
                    </span>
                  </div>

                  <div className="flex items-center gap-3 p-3 rounded-lg border bg-card">
                    <div className="rounded-md bg-purple-500/10 p-2">
                      <ClosedCaption className="h-4 w-4 text-purple-600" />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-medium">Transcriptions</p>
                      <p className="text-xs text-muted-foreground">
                        Meeting notes and recordings
                      </p>
                    </div>
                    <span className="text-sm font-semibold text-purple-600">
                      {metadata.transcriptions_count}
                    </span>
                  </div>
                </div>
              </div>

              {/* Import Button */}
              <div className="flex gap-2 pt-2">
                <Button
                  onClick={handleClose}
                  variant="outline"
                  className="flex-1"
                  disabled={isImporting}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleImport}
                  className="flex-1"
                  disabled={isImporting}
                >
                  {isImporting ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      <span>Importing...</span>
                    </>
                  ) : (
                    <>
                      <CheckCircle2 className="h-4 w-4" />
                      <span>Import Board</span>
                    </>
                  )}
                </Button>
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
