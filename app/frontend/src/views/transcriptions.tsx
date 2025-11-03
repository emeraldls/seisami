import { useState } from "react";
import { GetTranscriptions } from "../../wailsjs/go/main/App";
import { TranscriptionCard } from "~/components/transcription-card";
import { TranscriptionDetailModal } from "~/components/transcription-detail-modal";
import { Transcription } from "~/types/types";
import { useBoardStore } from "~/stores/board-store";
import { useQuery, useQueryClient } from "@tanstack/react-query";

export default function Transcriptions() {
  const queryClient = useQueryClient();
  const [selectedTranscription, setSelectedTranscription] =
    useState<Transcription | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const { currentBoard } = useBoardStore();

  const {
    data: transcriptions,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["transcriptions", currentBoard?.id],
    queryFn: async () => {
      if (!currentBoard) return [];
      const result = await GetTranscriptions(currentBoard.id, 1, 50);
      const formattedTranscriptions: Transcription[] = result.map((t) => ({
        id: t.ID,
        text: t.Transcription,
        timestamp: new Date(t.CreatedAt.String || new Date()),
        isTranscribing: false,
        intent: t.Intent?.String || undefined,
        assistantResponse: t.AssistantResponse?.String || undefined,
        recordingPath: t.RecordingPath?.String || undefined,
      }));
      return formattedTranscriptions;
    },
    enabled: !!currentBoard,
  });

  const updateTranscriptionInCache = (
    updater: (prev: Transcription[]) => Transcription[]
  ) => {
    queryClient.setQueryData(
      ["transcriptions", currentBoard?.id],
      (old: Transcription[] = []) => updater(old)
    );
  };

  const deleteTranscription = (id: string) => {
    updateTranscriptionInCache((prev) => prev.filter((t) => t.id !== id));
  };

  const copyTranscription = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const handleTranscriptionClick = (transcription: Transcription) => {
    setSelectedTranscription(transcription);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedTranscription(null);
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  console.log(transcriptions);

  return (
    <div className="flex-1 flex flex-col">
      <header className="bg-white border-b px-8 py-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl">
              Hold{" "}
              <kbd className="px-2 bg-black/70 text-white rounded text-sm font-mono">
                Fn
              </kbd>{" "}
              to dictate anywhere!
            </h2>
            {currentBoard && (
              <p className="text-sm text-neutral-500 mt-1">
                Working on:{" "}
                <span className="font-medium">{currentBoard.name}</span>
              </p>
            )}
          </div>
        </div>
      </header>
      {/* Waveform is now global and appears in top-right */}

      <div className="flex-1 p-8 overflow-auto">
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold ">History</h3>
              <p className="text-sm text-neutral-500 ">
                {transcriptions?.length} transcripts
              </p>
            </div>
          </div>

          {isLoading ? (
            <div className="flex items-center justify-center p-12">
              <div className="text-center space-y-2">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-neutral-900 mx-auto"></div>
                <p className="text-sm text-neutral-500">
                  Loading transcriptions...
                </p>
              </div>
            </div>
          ) : isError ? (
            <div className="text-center p-12">
              <p className="text-sm text-destructive">
                Failed to load transcriptions: {error?.message}
              </p>
            </div>
          ) : transcriptions?.length === 0 ? (
            <div className="text-center p-12">
              <p className="text-neutral-500">
                No transcriptions yet. Start recording to create your first one!
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {transcriptions &&
                transcriptions.map((transcription) => (
                  <TranscriptionCard
                    key={transcription.id}
                    transcription={transcription}
                    onCopy={copyTranscription}
                    onDelete={deleteTranscription}
                    onClick={handleTranscriptionClick}
                  />
                ))}
            </div>
          )}
        </div>
      </div>

      <TranscriptionDetailModal
        transcription={selectedTranscription}
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        onCopy={copyTranscription}
      />
    </div>
  );
}
