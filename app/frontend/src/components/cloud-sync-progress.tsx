import { useEffect, useState } from "react";
import { Cloud, CheckCircle2, AlertCircle, Loader } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { dataSyncService, type SyncProgress } from "~/lib/data-sync-service";

export const CloudSyncProgress = () => {
  const [progress, setProgress] = useState<SyncProgress | null>(null);

  useEffect(() => {
    const unsubscribe = dataSyncService.onProgress((newProgress) => {
      setProgress(newProgress);
    });

    return unsubscribe;
  }, []);

  if (!progress || progress.status === "idle") {
    return null;
  }

  const isError = progress.status === "error";
  const isComplete = progress.status === "completed";
  const isUploading =
    progress.status === "uploading" || progress.status === "preparing";

  return (
    <Card className="fixed bottom-4 right-4 w-96 shadow-lg z-50">
      <CardHeader className="pb-3">
        <div className="flex items-center gap-2">
          {isError ? (
            <AlertCircle className="h-5 w-5 text-red-500" />
          ) : isComplete ? (
            <CheckCircle2 className="h-5 w-5 text-green-500" />
          ) : (
            <Loader className="h-5 w-5 text-blue-500 animate-spin" />
          )}
          <div>
            <CardTitle className="text-base">
              {isError
                ? "Sync Failed"
                : isComplete
                ? "Sync Complete"
                : "Syncing to Cloud"}
            </CardTitle>
            {progress.message && (
              <CardDescription className="text-xs mt-1">
                {progress.message}
              </CardDescription>
            )}
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-3">
        {!isError && progress.percentComplete !== undefined && (
          <>
            <div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${progress.percentComplete}%` }}
                />
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                {progress.processedItems} of {progress.totalItems} items
                {progress.percentComplete !== undefined
                  ? ` (${progress.percentComplete}%)`
                  : ""}
              </p>
            </div>
          </>
        )}

        {isComplete && (
          <div className="text-sm space-y-1">
            {progress.duplicateBoards !== undefined &&
              progress.duplicateBoards > 0 && (
                <p className="text-yellow-600">
                  ⚠️ {progress.duplicateBoards} duplicate board(s) skipped
                </p>
              )}
            {progress.duplicateColumns !== undefined &&
              progress.duplicateColumns > 0 && (
                <p className="text-yellow-600">
                  ⚠️ {progress.duplicateColumns} duplicate column(s) skipped
                </p>
              )}
            {progress.duplicateCards !== undefined &&
              progress.duplicateCards > 0 && (
                <p className="text-yellow-600">
                  ⚠️ {progress.duplicateCards} duplicate card(s) skipped
                </p>
              )}
            {progress.failedItems > 0 && (
              <p className="text-red-600">
                ❌ {progress.failedItems} item(s) failed
              </p>
            )}
          </div>
        )}

        {isError && progress.error && (
          <p className="text-sm text-red-600">{progress.error}</p>
        )}
      </CardContent>
    </Card>
  );
};
