import { useDesktopAuthStore } from "~/stores/auth-store";
import { dataSyncService, type SyncPayload } from "~/lib/data-sync-service";
import { ExportDataForSync } from "../../wailsjs/go/main/App";

/**
 * Hook to collect all local data and sync with cloud
 * Call this when user enables cloud features
 */
export const useDataSync = () => {
  const syncToCloud = async () => {
    try {
      // Get auth token
      const authStore = useDesktopAuthStore.getState();
      if (!authStore.token) {
        throw new Error("Not authenticated");
      }

      const exportedData = await ExportDataForSync();

      const payload: SyncPayload = {
        boards: exportedData.boards || [],
        columns: exportedData.columns || [],
        cards: exportedData.cards || [],
        transcriptions: exportedData.transcriptions || [],
        settings: exportedData.settings,
      };

      // Start sync
      await dataSyncService.syncToCloud(payload, authStore.token!);
    } catch (error) {
      console.error("Cloud sync error:", error);
      throw error;
    }
  };

  return { syncToCloud };
};

/**
 * Alternative: Use this to trigger sync from a component
 * Example: Enable Cloud Button onClick handler
 */
export const triggerCloudSync = async () => {
  const authStore = useDesktopAuthStore.getState();
  if (!authStore.token) {
    throw new Error("Not authenticated");
  }

  try {
    // Export data from local database via Wails
    // @ts-ignore - Wails bindings will be generated at runtime
    const exportedData = await window.go.main.App.ExportDataForSync();

    const payload: SyncPayload = {
      boards: exportedData.boards || [],
      columns: exportedData.columns || [],
      cards: exportedData.cards || [],
      transcriptions: exportedData.transcriptions || [],
      settings: exportedData.settings,
    };

    // Start sync
    await dataSyncService.syncToCloud(payload, authStore.token!);
  } catch (error) {
    console.error("Cloud sync error:", error);
    throw error;
  }
};
