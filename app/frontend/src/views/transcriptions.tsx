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
        id: t.id,
        text: t.transcription,
        timestamp: new Date(t.created_at || new Date()),
        isTranscribing: false,
        intent: t.intent,
        assistantResponse: t.assistant_response,
        recordingPath: t.recording_path,
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

  return (
    <div className="flex-1 flex flex-col">
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
