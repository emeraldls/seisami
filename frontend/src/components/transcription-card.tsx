import React from "react";
import { Transcription } from "~/types/types";
import { Button } from "./ui/button";
import { Copy, Mic, SquareCheck, Trash2 } from "lucide-react";

const formatTime = (date: Date) => {
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
};

export const TranscriptionCard = ({
  transcription,
  onCopy,
  onDelete,
  onClick,
}: {
  transcription: Transcription;
  onCopy: (text: string) => void;
  onDelete: (id: string) => void;
  onClick?: (transcription: Transcription) => void;
}) => {
  return (
    <div
      key={transcription.id}
      className="group border-b py-2 transition-all duration-200 cursor-pointer hover:bg-gray-50 rounded-lg px-3"
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
              <div className="flex justify-between items-start">
                <div className="flex-1 max-w-[80%]">
                  <p className="text-sm text-neutral-800 text-balance">
                    {transcription.text}
                  </p>
                  {transcription.intent && (
                    <p className="text-xs text-blue-600 mt-1 font-medium">
                      Intent: {transcription.intent}
                    </p>
                  )}
                </div>
                <div className="ml-auto text-right">
                  <p className="text-xs text-neutral-400">
                    {transcription.timestamp.toLocaleDateString()} at{" "}
                    {formatTime(transcription.timestamp)}
                  </p>
                  {transcription.assistantResponse && (
                    <p className="text-xs text-green-600 mt-1 flex items-center gap-x-1">
                      <SquareCheck className="w-4 h-4" />
                      AI Processed
                    </p>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
