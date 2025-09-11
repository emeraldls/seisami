import React from "react";
import { Copy } from "lucide-react";
import { Transcription } from "~/types/types";
import { Button } from "./ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";

interface TranscriptionDetailModalProps {
  transcription: Transcription | null;
  isOpen: boolean;
  onClose: () => void;
  onCopy: (text: string) => void;
}

const formatDateTime = (date: Date) => {
  return {
    date: date.toLocaleDateString([], {
      weekday: "long",
      year: "numeric",
      month: "long",
      day: "numeric",
    }),
    time: date.toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    }),
  };
};

export const TranscriptionDetailModal: React.FC<
  TranscriptionDetailModalProps
> = ({ transcription, isOpen, onClose, onCopy }) => {
  if (!transcription) return null;

  const { date, time } = formatDateTime(transcription.timestamp);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Transcription Details</DialogTitle>
        </DialogHeader>

        <div className="space-y-6">
          {/* Timestamp */}
          <div>
            <h3 className="text-sm font-medium text-gray-900 mb-2">Recorded</h3>
            <p className="text-sm text-gray-600">{date}</p>
            <p className="text-xs text-gray-500">{time}</p>
          </div>

          {/* Transcription Text */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-sm font-medium text-gray-900">
                What you said
              </h3>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onCopy(transcription.text)}
                className="h-8 px-2"
              >
                <Copy className="h-3 w-3 mr-1" />
                Copy
              </Button>
            </div>
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-sm text-gray-900 leading-relaxed">
                "{transcription.text}"
              </p>
              {transcription.wordCount && (
                <p className="text-xs text-gray-500 mt-2">
                  {transcription.wordCount} words
                </p>
              )}
            </div>
          </div>

          {/* Intent */}
          {transcription.intent && (
            <div>
              <h3 className="text-sm font-medium text-gray-900 mb-2">
                AI Understanding
              </h3>
              <div className="bg-blue-50 rounded-lg p-4">
                <p className="text-sm text-blue-800 font-medium">
                  Intent: {transcription.intent}
                </p>
              </div>
            </div>
          )}

          {/* Assistant Response */}
          {transcription.assistantResponse && (
            <div>
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-sm font-medium text-gray-900">
                  AI Response
                </h3>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onCopy(transcription.assistantResponse!)}
                  className="h-8 px-2"
                >
                  <Copy className="h-3 w-3 mr-1" />
                  Copy
                </Button>
              </div>
              <div className="bg-green-50 rounded-lg p-4">
                <p className="text-sm text-green-800 leading-relaxed">
                  {transcription.assistantResponse}
                </p>
              </div>
            </div>
          )}

          {/* Recording Path (for debugging) */}
          {transcription.recordingPath && (
            <div>
              <h3 className="text-xs font-medium text-gray-500 mb-1">
                Recording File
              </h3>
              <p className="text-xs text-gray-400 font-mono bg-gray-100 p-2 rounded">
                {transcription.recordingPath}
              </p>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
