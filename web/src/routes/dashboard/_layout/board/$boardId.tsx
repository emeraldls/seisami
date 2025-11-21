import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { ApiClient, type BoardMetadata } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import {
  Loader2,
  FolderKanban,
  Columns3,
  CreditCard,
  ClosedCaption,
  AlertCircle,
  ArrowRight,
  Share2
} from "lucide-react";
import { DitherBackground } from "@/components/ui/dither-background";

export const Route = createFileRoute("/dashboard/_layout/board/$boardId")({
  component: BoardSharePage,
});

function BoardSharePage() {
  const { boardId } = Route.useParams();
  const [isLoading, setIsLoading] = useState(true);
  const [metadata, setMetadata] = useState<BoardMetadata | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isOpening, setIsOpening] = useState(false);

  useEffect(() => {
    fetchBoardMetadata();
  }, [boardId]);

  const fetchBoardMetadata = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await ApiClient.getBoardMetadata(boardId);
      setMetadata(response.data);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to fetch board metadata";
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  const openInDesktopApp = () => {
    if (!metadata) return;

    setIsOpening(true);

    const deepLink = `seisami://board/import?board_id=${metadata.id}`;

    window.location.href = deepLink;

    setTimeout(() => {
      setIsOpening(false);
    }, 2000);
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
    <div className="flex items-center justify-center min-h-[calc(100vh-200px)] relative overflow-hidden">
      <DitherBackground opacity={0.1} />
      
      <div className="w-full max-w-3xl relative z-10">
        <div className="mb-8 text-center">
          <div className="inline-flex items-center justify-center w-12 h-12 mb-6 bg-black dark:bg-white text-white dark:text-black">
            <Share2 size={24} />
          </div>
          <h1 className="text-3xl font-bold tracking-tighter uppercase mb-2">Shared Board</h1>
          <p className="text-gray-500 dark:text-gray-400 font-mono text-sm">Import this board into your Seisami workspace</p>
        </div>

        <div className="bg-white dark:bg-black border border-black dark:border-white p-8 sm:p-12 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] dark:shadow-[8px_8px_0px_0px_rgba(255,255,255,1)]">
          {isLoading && (
            <div className="flex flex-col items-center justify-center py-12 space-y-4">
              <Loader2 className="h-12 w-12 animate-spin text-black dark:text-white" />
              <div className="text-center space-y-1">
                <p className="text-sm font-mono uppercase tracking-wider">Loading board metadata...</p>
              </div>
            </div>
          )}

          {error && !isLoading && (
            <div className="flex flex-col items-center justify-center py-12 space-y-6">
              <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-full">
                <AlertCircle className="h-8 w-8 text-red-600 dark:text-red-400" />
              </div>
              <div className="text-center space-y-2">
                <p className="text-lg font-bold uppercase">Failed to load board</p>
                <p className="text-sm font-mono text-red-600 dark:text-red-400">{error}</p>
              </div>
              <Button 
                onClick={fetchBoardMetadata} 
                variant="outline" 
                className="rounded-none border-black dark:border-white hover:bg-gray-50 dark:hover:bg-gray-900 font-mono uppercase tracking-widest"
              >
                Try Again
              </Button>
            </div>
          )}

          {metadata && !isLoading && (
            <div className="space-y-12">
              <div className="flex items-start gap-6 pb-8 border-b border-black/10 dark:border-white/10">
                <div className="bg-black dark:bg-white text-white dark:text-black p-4 hidden sm:block">
                  <FolderKanban className="h-8 w-8" />
                </div>
                <div className="space-y-2">
                  <h3 className="text-3xl sm:text-4xl font-bold tracking-tighter uppercase">{metadata.name}</h3>
                  <div className="flex items-center gap-2 text-xs font-mono text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    <span>Created {formatDate(metadata.created_at)}</span>
                    <span>•</span>
                    <span>ID: {metadata.id.substring(0, 8)}</span>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4 sm:gap-8">
                <div className="flex flex-col items-center gap-2 p-6 border border-black/10 dark:border-white/10 bg-gray-50 dark:bg-white/5">
                  <Columns3 className="h-6 w-6 text-gray-400 mb-2" />
                  <p className="text-3xl font-bold">{metadata.columns_count}</p>
                  <p className="text-xs font-mono uppercase tracking-wider text-gray-500">Columns</p>
                </div>

                <div className="flex flex-col items-center gap-2 p-6 border border-black/10 dark:border-white/10 bg-gray-50 dark:bg-white/5">
                  <CreditCard className="h-6 w-6 text-gray-400 mb-2" />
                  <p className="text-3xl font-bold">{metadata.cards_count}</p>
                  <p className="text-xs font-mono uppercase tracking-wider text-gray-500">Cards</p>
                </div>

                <div className="flex flex-col items-center gap-2 p-6 border border-black/10 dark:border-white/10 bg-gray-50 dark:bg-white/5">
                  <ClosedCaption className="h-6 w-6 text-gray-400 mb-2" />
                  <p className="text-3xl font-bold">{metadata.transcriptions_count}</p>
                  <p className="text-xs font-mono uppercase tracking-wider text-gray-500">Voice Notes</p>
                </div>
              </div>

              <Button
                onClick={openInDesktopApp}
                className="w-full h-16 text-lg rounded-none bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 font-mono uppercase tracking-widest"
                disabled={isOpening}
              >
                {isOpening ? (
                  <>
                    <Loader2 className="h-5 w-5 animate-spin mr-3" />
                    Opening App...
                  </>
                ) : (
                  <>
                    Open in Desktop App <ArrowRight className="ml-3 h-5 w-5" />
                  </>
                )}
              </Button>
              
              <p className="text-center text-xs font-mono text-gray-400 dark:text-gray-600">
                Requires Seisami Desktop App installed on your device.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
