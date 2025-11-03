import { Outlet } from "react-router-dom";
import { Sidebar } from "~/components/sidebar";
import { CloudSyncProgress } from "~/components/cloud-sync-progress";
import { useEffect, useRef, useState } from "react";
import { useSidebar } from "~/contexts/sidebar-context";
import { EventsOn, EventsEmit } from "../../wailsjs/runtime/runtime";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import { Info, MessageCircle } from "lucide-react";
import { useDesktopAuthStore } from "~/stores/auth-store";
import { BoardImportDialog } from "~/components/board-import-dialog";
import { GlobalWaveform } from "~/components/global-waveform";
import { useRecordingStore } from "~/stores/recording-store";
import { useBoardStore } from "~/stores/board-store";
import { Transcription } from "~/types/types";
import { useQueryClient } from "@tanstack/react-query";

export const AppLayout = () => {
  const { collapsed } = useSidebar();
  const cloudToastIdRef = useRef<string | number | null>(null);
  const tokenSentRef = useRef(false);
  const { token } = useDesktopAuthStore();
  const [importDialogOpen, setImportDialogOpen] = useState(false);
  const [boardIdToImport, setBoardIdToImport] = useState<string | undefined>();
  const queryClient = useQueryClient();
  const { currentBoard } = useBoardStore();
  const {
    setIsRecording,
    setAudioBars,
    addWaveformData,
    resetWaveform,
    addPendingTranscription,
    updatePendingTranscription,
    removePendingTranscription,
  } = useRecordingStore();

  useEffect(() => {
    if (token && !tokenSentRef.current) {
      console.log("Sending auth token to backend...");
      EventsEmit("auth:set_token", token);
      tokenSentRef.current = true;
    }
  }, [token]);

  // Helper function to update transcriptions in cache
  const updateTranscriptionInCache = (
    updater: (prev: Transcription[]) => Transcription[]
  ) => {
    queryClient.setQueryData(
      ["transcriptions", currentBoard?.id],
      (old: Transcription[] = []) => updater(old)
    );
  };

  useEffect(() => {
    const unsubscribeProcessingStart = EventsOn(
      "ai:processing_start",
      (data: any) => {
        toast.info("🤖 AI is analyzing your request...", {
          description: "Processing your voice command",
          duration: 2000,
        });
      }
    );

    const unsubscribeToolComplete = EventsOn(
      "ai:tool_complete",
      (data: any) => {
        toast.success(`✅ Completed: ${data.toolName}`, {
          description:
            data.result?.length > 100
              ? `${data.result.substring(0, 100)}...`
              : data.result,
          duration: 2000,
        });
      }
    );

    const unsubscribeToolError = EventsOn("ai:tool_error", (data: any) => {
      toast.error(`❌ Failed: ${data.toolName}`, {
        description: data.error,
        duration: 3000,
      });
    });

    const unsubscribeProcessingComplete = EventsOn(
      "ai:processing_complete",
      (data: any) => {
        toast.success("🎉 AI finished processing!", {
          description: data.intent
            ? `Intent: ${data.intent}`
            : "Task completed successfully",
          duration: 3000,
        });
      }
    );

    const unsubscribeAiError = EventsOn("ai:error", (data: any) => {
      toast.error("🚨 AI Error", {
        description: data.error,
        duration: 4000,
      });
    });

    const unsubscribeCloudSetupStarted = EventsOn("cloud:setup_started", () => {
      cloudToastIdRef.current = toast.loading("☁️ Connecting to cloud...", {
        description: "Setting up your workspace",
      });
    });

    const unsubscribeCloudSetupSuccess = EventsOn("cloud:setup_success", () => {
      const toastId = cloudToastIdRef.current ?? undefined;
      toast.success("☁️ Cloud ready", {
        id: toastId,
        description: "Your workspace is connected to Seisami Cloud.",
      });
      cloudToastIdRef.current = null;
    });

    const unsubscribeCloudSetupFailed = EventsOn(
      "cloud:setup_failed",
      (error: unknown) => {
        const description =
          typeof error === "string"
            ? error
            : (error as { message?: string })?.message ??
              "Unable to complete cloud setup.";

        const toastId = cloudToastIdRef.current ?? undefined;
        toast.error("☁️ Cloud setup failed", {
          id: toastId,
          description,
          duration: 4000,
        });
        cloudToastIdRef.current = null;
      }
    );

    // Listen for board import deep link event
    const unsubscribeBoardImport = EventsOn(
      "board:import_request",
      (data: { board_id: string }) => {
        console.log("Board import requested:", data.board_id);
        setBoardIdToImport(data.board_id);
        setImportDialogOpen(true);
      }
    );

    // Global transcription event listeners
    const unsubscribeTranscription = EventsOn(
      "transcription",
      (data: string) => {
        if (data) {
          try {
            const parsed = JSON.parse(data) as {
              id: string;
              transcription: string;
            };
            if (parsed.transcription) {
              updateTranscriptionInCache((prev) =>
                prev.map((t) =>
                  t.id === parsed.id
                    ? {
                        ...t,
                        text: parsed.transcription,
                        isTranscribing: false,
                        wordCount: parsed.transcription.split(" ").length,
                      }
                    : t
                )
              );
              updatePendingTranscription(parsed.id, {
                text: parsed.transcription,
                isTranscribing: false,
                wordCount: parsed.transcription.split(" ").length,
              });
            }
          } catch (e) {
            console.error("Failed to parse transcription event data", e);
          }
        }
      }
    );

    const unsubscribeStructuredResponse = EventsOn(
      "structured_response",
      (data: string) => {
        if (data) {
          try {
            const parsed = JSON.parse(data);
            console.log("Structured Response received:", parsed);

            updateTranscriptionInCache((prev) => {
              const updated = [...prev];
              if (updated.length > 0) {
                updated[0] = {
                  ...updated[0],
                  intent: parsed.intent,
                  assistantResponse: parsed.result,
                };
              }
              return updated;
            });
          } catch (e) {
            console.error("Failed to parse structured response data", e);
          }
        }
      }
    );

    const unsubscribeAudioBars = EventsOn("audio_bars", (data) => {
      console.log(data);
      if (Array.isArray(data) && data.length > 0) {
        setAudioBars(data);
        addWaveformData(data);
      }
    });

    const unsubscribeRecordingStart = EventsOn("recording:start", () => {
      setIsRecording(true);
      resetWaveform();
      if (currentBoard) {
        console.log("emitting event: ", currentBoard.id);
        EventsEmit("board:id", currentBoard.id);
      }
    });

    const unsubscribeRecordingStop = EventsOn(
      "recording:stop",
      (data: string) => {
        setIsRecording(false);
        setAudioBars(null);
        resetWaveform();
        if (data) {
          try {
            const parsed = JSON.parse(data) as { id: string };
            const newTranscription: Transcription = {
              id: parsed.id,
              text: "Transcribing...",
              timestamp: new Date(),
              isTranscribing: true,
            };
            updateTranscriptionInCache((prev) => [newTranscription, ...prev]);
            addPendingTranscription(parsed.id, newTranscription);
          } catch (e) {
            console.error("Failed to parse recording:stop event data", e);
          }
        }
      }
    );

    const unsubscribeTranscriptionShort = EventsOn(
      "transcription:short",
      (data: string) => {
        if (data) {
          try {
            const parsed = JSON.parse(data) as { id: string };
            updateTranscriptionInCache((prev) =>
              prev.filter((t) => t.id !== parsed.id)
            );
            removePendingTranscription(parsed.id);
          } catch (e) {
            console.error("Failed to parse transcription:short event data", e);
          }
        }
      }
    );

    return () => {
      unsubscribeProcessingStart();
      unsubscribeToolComplete();
      unsubscribeToolError();
      unsubscribeProcessingComplete();
      unsubscribeAiError();
      unsubscribeCloudSetupStarted();
      unsubscribeCloudSetupSuccess();
      unsubscribeCloudSetupFailed();
      unsubscribeBoardImport();
      unsubscribeTranscription();
      unsubscribeStructuredResponse();
      unsubscribeAudioBars();
      unsubscribeRecordingStart();
      unsubscribeRecordingStop();
      unsubscribeTranscriptionShort();
    };
  }, [currentBoard, queryClient]);

  return (
    <>
      <div className="flex relative">
        <Sidebar />
        <div
          className="w-full transition-all duration-300 min-h-screen"
          style={{ paddingLeft: collapsed ? 72 : 256 }}
        >
          <Outlet />
        </div>

        <div className="fixed bottom-6 right-6 z-50">
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="rounded-full w-12 h-12 shadow-lg hover:shadow-xl transition-shadow"
              >
                <svg
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <circle cx="12" cy="12" r="10" />
                  <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
                  <path d="M12 17h.01" />
                </svg>
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-48 p-2" align="end" side="top">
              <div>
                <Button
                  variant="ghost"
                  className="w-full justify-start text-sm"
                >
                  <Info /> Help Center
                </Button>
                <Button
                  variant="ghost"
                  className="w-full justify-start text-sm"
                >
                  <MessageCircle /> Report a Bug
                </Button>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>
      <CloudSyncProgress />
      <GlobalWaveform />
      <BoardImportDialog
        open={importDialogOpen}
        onOpenChange={setImportDialogOpen}
        boardId={boardIdToImport}
      />
    </>
  );
};
