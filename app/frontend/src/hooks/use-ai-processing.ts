import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useNavigate } from "react-router-dom";
import { EventsEmit } from "../../wailsjs/runtime/runtime";
import { useRecordingStore } from "~/stores/recording-store";
import { useBoardStore } from "~/stores/board-store";
import type { Transcription } from "~/types/types";

export interface AIEvent {
  type: string;
  data?: any;
  error?: string;
}

export const useAIProcessing = () => {
  const queryClient = useQueryClient();
  const { currentBoard } = useBoardStore();
  const { setProcessingState, setCurrentAction } = useRecordingStore();

  const updateTranscriptionInCache = (
    updater: (prev: Transcription[]) => Transcription[]
  ) => {
    queryClient.setQueryData(
      ["transcriptions", currentBoard?.id],
      (old: Transcription[] = []) => updater(old)
    );
  };

  const handleAIEvent = (event: AIEvent) => {
    console.log("AI Event received:", event);

    switch (event.type) {
      case "ai:transcription_start":
      case "transcription_start":
        setProcessingState("transcribing");
        setCurrentAction("Transcribing with cloud AI...");
        break;

      case "ai:transcription_complete":
      case "transcription_complete":
        if (event.data?.transcription) {
          setProcessingState("processing");
          setCurrentAction("Processing with AI...");

          // Update the transcription in cache
          updateTranscriptionInCache((prev) => {
            const updated = [...prev];
            if (updated.length > 0) {
              updated[0] = {
                ...updated[0],
                text: event.data.transcription,
                isTranscribing: false,
                wordCount: event.data.transcription.split(" ").length,
              };
            }
            return updated;
          });
        }
        break;

      case "ai:processing_start":
      case "ai_processing_start":
        setProcessingState("processing");
        setCurrentAction("Processing with AI...");
        break;

      case "ai:tool_complete":
      case "ai_tool_complete":
        console.log("Tool complete event data:", event.data);
        if (event.data?.toolName || event.data?.tool_name) {
          const toolName = event.data.toolName || event.data.tool_name;
          setProcessingState("processing"); // Ensure we're in processing state
          setCurrentAction(toolName);
          EventsEmit("board:refetch");
        } else {
          console.warn("No toolName found in event data:", event.data);
        }
        break;

      case "ai:tool_error":
      case "ai_tool_error":
        console.log("Tool error event data:", event.data);
        const toolNameError = event.data?.toolName || event.data?.tool_name;
        const errorMsg = event.data?.error || event.error;
        if (toolNameError && errorMsg) {
          setProcessingState("processing"); // Ensure we're in processing state
          setCurrentAction(`Error: ${toolNameError}`);
          toast.error(`❌ ${toolNameError} failed`, {
            description: errorMsg,
            duration: 3000,
          });
        } else {
          console.warn("Missing toolName or error in event data:", event.data);
        }
        break;

      case "ai:processing_complete":
      case "ai_processing_complete":
        if (event.data?.result) {
          // Update transcription with final result
          updateTranscriptionInCache((prev) => {
            const updated = [...prev];
            if (updated.length > 0) {
              updated[0] = {
                ...updated[0],
                assistantResponse: event.data.result,
                intent: event.data.intent,
              };
            }
            return updated;
          });
        }
        // Don't change state here, let ai:done handle it
        break;

      case "ai:done":
      case "done":
        // Show complete state briefly before hiding
        setProcessingState("complete");
        setCurrentAction(null);

        setTimeout(() => {
          setProcessingState("idle");
        }, 1000);
        break;

      case "ai:error":
      case "error":
        setProcessingState("idle");
        setCurrentAction(null);
        toast.error("AI Processing Failed", {
          description: event.error || "An error occurred during AI processing",
          duration: 5000,
        });
        break;

      default:
        console.warn("Unhandled AI event type:", event.type);
    }
  };

  return { handleAIEvent };
};
