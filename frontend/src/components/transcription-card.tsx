import React from "react";
import { Transcription } from "~/types/types";
import { Button } from "./ui/button";
import { Copy, Mic, Trash2 } from "lucide-react";

const formatTime = (date: Date) => {
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
};

export const TranscriptionCard = ({
  transcription,
  onCopy,
  onDelete,
}: {
  transcription: Transcription;
  onCopy: (text: string) => void;
  onDelete: (id: string) => void;
}) => {
  return (
    <div
      key={transcription.id}
      className="group border-b py-2 transition-all duration-200"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="leading-relaxed text-pretty">
            {transcription.isTranscribing ? (
              <span className="text-sm animate-shimmer">
                Transcribing <span className="animate-pulse">...</span>
              </span>
            ) : (
              <div className="flex justify-between items-center">
                <p className="text-sm text-neutral-800 text-balance max-w-[80%]">
                  {transcription.text}
                </p>
                <p className="ml-auto text-xs text-neutral-400 ">
                  {transcription.timestamp.toLocaleDateString()} at{" "}
                  {formatTime(transcription.timestamp)}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
