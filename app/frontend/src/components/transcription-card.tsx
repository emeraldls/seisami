import React from "react";
import { Transcription } from "~/types/types";
import { Button } from "./ui/button";
import { Copy, Mic, SquareCheck, Trash2, RefreshCw } from "lucide-react";
import { ReprocessTranscription } from "../../wailsjs/go/main/App";
import { toast } from "sonner";
import { useBoardStore } from "~/stores/board-store";

const formatTime = (date: Date) => {
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
};

const truncateText = (text: string, maxLength: number = 200) => {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength).trim() + "...";
};

export const TranscriptionCard = ({
  transcription,
  onClick,
}: {
  transcription: Transcription;
  onCopy: (text: string) => void;
  onDelete: (id: string) => void;
  onClick?: (transcription: Transcription) => void;
}) => {
  const { currentBoard } = useBoardStore();
  const [isReprocessing, setIsReprocessing] = React.useState(false);

  const handleReprocess = async (e: React.MouseEvent) => {
    e.stopPropagation();

    if (!currentBoard?.id) {
      toast.error("No board selected");
      return;
    }

    if (transcription.isTranscribing) {
      toast.info("Please wait for transcription to complete");
      return;
    }

    setIsReprocessing(true);
    toast.info("🤖 Reprocessing transcription...", {
      description: "AI is analyzing your request again",
      duration: 2000,
    });

    try {
      await ReprocessTranscription(
        transcription.id,
        transcription.text,
        currentBoard.id
      );
      toast.success("✅ Transcription reprocessed successfully", {
        duration: 3000,
      });
    } catch (error) {
      console.error("Failed to reprocess transcription:", error);
      toast.error("❌ Failed to reprocess transcription", {
        description: error instanceof Error ? error.message : "Unknown error",
        duration: 4000,
      });
    } finally {
      setIsReprocessing(false);
    }
  };

  return (
    <div
      key={transcription.id}
      className="group border-b py-2 transition-all duration-200 cursor-pointer hover:bg-neutral-50 rounded-lg px-3"
      onClick={() => onClick?.(transcription)}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="leading-relaxed text-pretty">
            {transcription.isTranscribing ? (
              <span className="text-sm animate-shimmer">
                Transcribing <span className="animate-pulse">...</span>
              </span>
            ) : (
              <div className="flex justify-between items-start gap-3">
                <div className="flex-1 max-w-[80%]">
                  <p className="text-sm text-neutral-800">
                    {truncateText(transcription.text)}
                  </p>
                  {transcription.intent && (
                    <p className="text-xs text-blue-600 mt-1 font-medium">
                      Intent: {transcription.intent}
                    </p>
                  )}
                </div>
                <div className="ml-auto text-right flex items-start gap-2">
                  <div>
                    <p className="text-xs text-neutral-400">
                      {transcription.timestamp.toLocaleDateString()} at{" "}
                      {formatTime(transcription.timestamp)}
                    </p>
                    {transcription.assistantResponse && (
                      <p className="text-xs text-green-600 mt-1 flex items-center gap-x-1 justify-end">
                        <SquareCheck className="w-4 h-4" />
                        AI Processed
                      </p>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                    onClick={handleReprocess}
                    disabled={isReprocessing}
                    title="Reprocess with AI"
                  >
                    <RefreshCw
                      className={`h-3 w-3 ${
                        isReprocessing ? "animate-spin" : ""
                      }`}
                    />
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
