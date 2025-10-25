import { toast } from "sonner";

interface SyncBoard {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

interface SyncColumn {
  id: string;
  board_id: string;
  name: string;
  position: number;
  created_at: string;
  updated_at: string;
}

interface SyncCard {
  id: string;
  column_id: string;
  title: string;
  description?: string;
  attachments?: string;
  created_at: string;
  updated_at: string;
}

interface SyncTranscription {
  id: string;
  board_id: string;
  transcription: string;
  recording_path?: string;
  intent?: string;
  assistant_response?: string;
  created_at: string;
  updated_at: string;
}

interface SyncSettings {
  id: string;
  transcription_method: string;
  whisper_binary_path?: string;
  whisper_model_path?: string;
  openai_api_key?: string;
  created_at: string;
  updated_at: string;
}

interface SyncPayload {
  boards: SyncBoard[];
  columns: SyncColumn[];
  cards: SyncCard[];
  transcriptions: SyncTranscription[];
  settings?: SyncSettings;
}

interface SyncProgress {
  status: "idle" | "preparing" | "uploading" | "completed" | "error";
  totalItems: number;
  processedItems: number;
  failedItems: number;
  duplicateBoards?: number;
  duplicateColumns?: number;
  duplicateCards?: number;
  message?: string;
  error?: string;
  percentComplete?: number;
}

type SyncProgressCallback = (progress: SyncProgress) => void;

class DataSyncService {
  private serverUrl: string = "http://localhost:8080";
  private progressCallbacks: Set<SyncProgressCallback> = new Set();

  setServerUrl(url: string) {
    this.serverUrl = url;
  }

  onProgress(callback: SyncProgressCallback) {
    this.progressCallbacks.add(callback);
    return () => {
      this.progressCallbacks.delete(callback);
    };
  }

  private notifyProgress(progress: SyncProgress) {
    this.progressCallbacks.forEach((callback) => callback(progress));
  }

  async syncToCloud(payload: SyncPayload, token: string): Promise<void> {
    try {
      const totalItems =
        payload.boards.length +
        payload.columns.length +
        payload.cards.length +
        payload.transcriptions.length +
        (payload.settings ? 1 : 0);

      this.notifyProgress({
        status: "preparing",
        totalItems,
        processedItems: 0,
        failedItems: 0,
        message: "Preparing to upload your data...",
      });

      // Small delay for UX feedback
      await new Promise((resolve) => setTimeout(resolve, 500));

      this.notifyProgress({
        status: "uploading",
        totalItems,
        processedItems: 0,
        failedItems: 0,
        percentComplete: 0,
        message: "Uploading to cloud...",
      });

      const response = await fetch(`${this.serverUrl}/sync/upload`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Upload failed");
      }

      // Simulate progress updates while processing happens server-side
      let processedItems = 0;
      const interval = setInterval(() => {
        processedItems = Math.min(
          processedItems + Math.floor(Math.random() * totalItems * 0.1),
          totalItems - 1
        );
        const percentComplete = Math.floor((processedItems / totalItems) * 100);

        this.notifyProgress({
          status: "uploading",
          totalItems,
          processedItems,
          failedItems: 0,
          percentComplete,
          message: `Uploading... ${percentComplete}%`,
        });
      }, 300);

      // Wait for server processing (poll for status)
      const maxAttempts = 30; // 30 seconds max
      let attempts = 0;
      let finalStatus: SyncProgress | null = null;

      while (attempts < maxAttempts) {
        await new Promise((resolve) => setTimeout(resolve, 1000));
        attempts++;

        try {
          const statusResponse = await fetch(`${this.serverUrl}/sync/status`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });

          if (statusResponse.ok) {
            const status = await statusResponse.json();

            if (status.status === "completed" || status.status === "idle") {
              clearInterval(interval);

              finalStatus = {
                status: "completed",
                totalItems,
                processedItems: totalItems,
                failedItems: status.failed_items || 0,
                duplicateBoards: status.duplicate_boards || 0,
                duplicateColumns: status.duplicate_columns || 0,
                duplicateCards: status.duplicate_cards || 0,
                percentComplete: 100,
                message:
                  status.message ||
                  "Your data has been successfully synced to the cloud!",
              };

              this.notifyProgress(finalStatus);
              toast.success("Cloud Sync Complete", {
                description: finalStatus.message,
              });
              return;
            }
          }
        } catch (error) {
          // Polling error - continue
        }
      }

      clearInterval(interval);

      // If we get here, sync completed with timeout
      if (finalStatus === null) {
        finalStatus = {
          status: "completed",
          totalItems,
          processedItems: totalItems,
          failedItems: 0,
          percentComplete: 100,
          message: "Cloud sync uploaded. Server is processing your data.",
        };

        this.notifyProgress(finalStatus);
        toast.success("Sync Uploaded", {
          description:
            "Your data is being processed on the cloud. Check back soon!",
        });
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Sync failed";

      this.notifyProgress({
        status: "error",
        totalItems: 0,
        processedItems: 0,
        failedItems: 0,
        error: errorMessage,
        message: errorMessage,
      });

      toast.error("Cloud Sync Failed", {
        description: errorMessage,
      });

      throw error;
    }
  }
}

export const dataSyncService = new DataSyncService();
export type { SyncPayload, SyncProgress };
