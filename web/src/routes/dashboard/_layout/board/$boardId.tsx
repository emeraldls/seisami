import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { ApiClient, type BoardMetadata } from "@/lib/api-client";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Download,
  Loader2,
  FolderKanban,
  Columns3,
  CreditCard,
  ClosedCaption,
  AlertCircle,
} from "lucide-react";

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
    <div className="flex items-center justify-center min-h-[calc(100vh-200px)]">
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-2xl">
            <FolderKanban className="h-6 w-6" />
            Board Shared With You
          </CardTitle>
          <CardDescription>
            Open this board in the Seisami desktop app
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading && (
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

          {error && !isLoading && (
            <div className="flex flex-col items-center justify-center py-12 space-y-4">
              <div className="rounded-full bg-destructive/10 p-3">
                <AlertCircle className="h-8 w-8 text-destructive" />
              </div>
              <div className="text-center space-y-1">
                <p className="text-sm font-medium">Failed to load board</p>
                <p className="text-xs text-muted-foreground">{error}</p>
              </div>
              <Button onClick={fetchBoardMetadata} variant="outline" size="sm">
                Try Again
              </Button>
            </div>
          )}

          {metadata && !isLoading && (
            <div className="space-y-6">
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

              <div className="space-y-2">
                <p className="text-sm font-medium text-muted-foreground">
                  This board contains:
                </p>
                <div className="grid grid-cols-3 gap-3">
                  <div className="flex flex-col items-center gap-2 p-4 rounded-lg border bg-card">
                    <div className="rounded-md bg-blue-500/10 p-2">
                      <Columns3 className="h-5 w-5 text-blue-600" />
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-blue-600">
                        {metadata.columns_count}
                      </p>
                      <p className="text-xs text-muted-foreground">Columns</p>
                    </div>
                  </div>

                  <div className="flex flex-col items-center gap-2 p-4 rounded-lg border bg-card">
                    <div className="rounded-md bg-green-500/10 p-2">
                      <CreditCard className="h-5 w-5 text-green-600" />
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-green-600">
                        {metadata.cards_count}
                      </p>
                      <p className="text-xs text-muted-foreground">Cards</p>
                    </div>
                  </div>

                  <div className="flex flex-col items-center gap-2 p-4 rounded-lg border bg-card">
                    <div className="rounded-md bg-purple-500/10 p-2">
                      <ClosedCaption className="h-5 w-5 text-purple-600" />
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-purple-600">
                        {metadata.transcriptions_count}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        Transcriptions
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <Button
                onClick={openInDesktopApp}
                className="w-full"
                size="lg"
                disabled={isOpening}
              >
                {isOpening ? (
                  <>
                    <Loader2 className="h-5 w-5 animate-spin" />
                    <span>Opening Desktop App...</span>
                  </>
                ) : (
                  <>
                    <Download className="h-5 w-5" />
                    <span>Open in Desktop App</span>
                  </>
                )}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
