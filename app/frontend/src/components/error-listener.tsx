import React, { useEffect } from "react";
import { EventsOn, EventsEmit } from "../../wailsjs/runtime/runtime";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { useBoardStore } from "~/stores/board-store";

export const ErrorListener = () => {
  const { currentBoard } = useBoardStore();
  useEffect(() => {
    const unsubscribeSyncError = EventsOn("sync:error", (message: string) => {
      toast.error("Sync Error", {
        description: message,
        duration: 5000,
      });
    });

    const unsubscribeSyncPullError = EventsOn(
      "sync:pull_error",
      (message: string) => {
        toast.error("Failed to Pull Data", {
          description: message,
          duration: 5000,
        });
      }
    );

    const unsubscribeSyncPushError = EventsOn(
      "sync:push_error",
      (message: string) => {
        toast.error("Failed to Push Data", {
          description: message,
          duration: 5000,
        });
      }
    );

    const unsubscribeSyncLocalUpdateError = EventsOn(
      "sync:local_update_error",
      (message: string) => {
        toast.error("Local Update Failed", {
          description: message,
          duration: 5000,
        });
      }
    );

    const unsubscribeSyncStateError = EventsOn(
      "sync:state_error",
      (message: string) => {
        toast.error("Sync State Error", {
          description: message,
          duration: 5000,
        });
      }
    );

    const unsubscribeBootstrapError = EventsOn(
      "bootstrap:error",
      (message: string) => {
        toast.error("Bootstrap Failed", {
          description: message,
          duration: 5000,
        });
      }
    );

    const unsubscribeBootstrapBoardError = EventsOn(
      "bootstrap:board_error",
      (message: string) => {
        toast.error("Board Sync Error", {
          description: message,
          duration: 4000,
        });
      }
    );

    const unsubscribeBootstrapColumnError = EventsOn(
      "bootstrap:column_error",
      (message: string) => {
        toast.error("Column Sync Error", {
          description: message,
          duration: 4000,
        });
      }
    );

    const unsubscribeBootstrapCardError = EventsOn(
      "bootstrap:card_error",
      (message: string) => {
        toast.error("Card Sync Error", {
          description: message,
          duration: 4000,
        });
      }
    );

    const unsubscribeBootstrapTranscriptionError = EventsOn(
      "bootstrap:transcription_error",
      (message: string) => {
        toast.error("Transcription Sync Error", {
          description: message,
          duration: 4000,
        });
      }
    );

    const unsubscribeBootstrapInitError = EventsOn(
      "bootstrap:init_error",
      (message: string) => {
        toast.error("Bootstrap Init Failed", {
          description: message,
          duration: 5000,
        });
      }
    );

    const unsubscribeBootstrapStateError = EventsOn(
      "bootstrap:state_error",
      (message: string) => {
        toast.error("Bootstrap State Error", {
          description: message,
          duration: 4000,
        });
      }
    );

    const unsubscribeImportError = EventsOn(
      "import:error",
      (message: string) => {
        toast.error("Import Failed", {
          description: message,
          duration: 5000,
        });
      }
    );

    const unsubscribeImportColumnError = EventsOn(
      "import:column_error",
      (message: string) => {
        toast.error("Column Import Error", {
          description: message,
          duration: 4000,
        });
      }
    );

    const unsubscribeImportCardError = EventsOn(
      "import:card_error",
      (message: string) => {
        toast.error("Card Import Error", {
          description: message,
          duration: 4000,
        });
      }
    );

    const unsubscribeImportTranscriptionError = EventsOn(
      "import:transcription_error",
      (message: string) => {
        toast.error("Transcription Import Error", {
          description: message,
          duration: 4000,
        });
      }
    );

    // Success listeners for positive feedback
    const unsubscribeSyncStarted = EventsOn(
      "sync:started",
      (data: { table: string }) => {
        toast.info(`Syncing ${data.table}...`, {
          description: "Synchronizing data with cloud",
          duration: 2000,
        });
      }
    );

    const unsubscribeSyncCompleted = EventsOn(
      "sync:completed",
      (data: { table: string; pushed: boolean; pulled: boolean }) => {
        const actions = [];
        if (data.pushed) actions.push("pushed");
        if (data.pulled) actions.push("pulled");

        if (actions.length > 0) {
          toast.success(`${data.table} synced`, {
            description: `Successfully ${actions.join(" and ")} changes`,
            duration: 3000,
          });
        }
      }
    );

    const unsubscribeBootstrapStarted = EventsOn(
      "bootstrap:started",
      (message: string) => {
        toast.info("Bootstrapping Cloud", {
          description: message,
          duration: 3000,
        });
      }
    );

    const unsubscribeBootstrapCompleted = EventsOn(
      "bootstrap:completed",
      (message: string) => {
        toast.success("Cloud Bootstrap Complete", {
          description: message,
          duration: 4000,
        });
      }
    );

    const unsubscribeImportStarted = EventsOn(
      "import:started",
      (data: { boardId: string }) => {
        toast.info("Importing Board", {
          description: `Fetching board data...`,
          duration: 2000,
        });
      }
    );

    const unsubscribeImportCompleted = EventsOn(
      "import:completed",
      (data: {
        boardId: string;
        boardName: string;
        columnsCount: number;
        cardsCount: number;
        transcriptions: number;
      }) => {
        toast.success(`Board "${data.boardName}" Imported`, {
          description: `${data.columnsCount} columns, ${data.cardsCount} cards, ${data.transcriptions} transcriptions`,
          duration: 5000,
        });
      }
    );

    return () => {
      // Sync errors
      unsubscribeSyncError();
      unsubscribeSyncPullError();
      unsubscribeSyncPushError();
      unsubscribeSyncLocalUpdateError();
      unsubscribeSyncStateError();
      // Bootstrap errors
      unsubscribeBootstrapError();
      unsubscribeBootstrapBoardError();
      unsubscribeBootstrapColumnError();
      unsubscribeBootstrapCardError();
      unsubscribeBootstrapTranscriptionError();
      unsubscribeBootstrapInitError();
      unsubscribeBootstrapStateError();
      // Import errors
      unsubscribeImportError();
      unsubscribeImportColumnError();
      unsubscribeImportCardError();
      unsubscribeImportTranscriptionError();
      // Success events
      unsubscribeSyncStarted();
      unsubscribeSyncCompleted();
      unsubscribeBootstrapStarted();
      unsubscribeBootstrapCompleted();
      unsubscribeImportStarted();
      unsubscribeImportCompleted();
    };
  }, [currentBoard]);
  return <></>;
};
