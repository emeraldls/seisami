import { Outlet, useNavigate } from "react-router-dom";
import { Sidebar } from "~/components/sidebar";
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
import { ErrorListener } from "~/components/error-listener";
import { TopNavbar } from "~/components/top-navbar";
import { useCollaborationStore } from "~/stores/collab-store";
import { ApiClient } from "~/lib/api-client";
import { ReadAudioFile } from "../../wailsjs/go/main/App";
import { useAIProcessing } from "~/hooks/use-ai-processing";

export const AppLayout = () => {
  const { collapsed } = useSidebar();
  const cloudToastIdRef = useRef<string | number | null>(null);
  const tokenSentRef = useRef(false);
  const currentBoardSentRef = useRef(false);
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
    setProcessingState,
    setCurrentAction,
  } = useRecordingStore();

  useEffect(() => {
    if (token && !tokenSentRef.current) {
      console.log("Sending auth token to backend...");
      EventsEmit("auth:set_token", token);
      tokenSentRef.current = true;
    }
  }, [token]);

  useEffect(() => {
    if (currentBoard && !currentBoardSentRef.current) {
      EventsEmit("auth:set_current_board", currentBoard.id);
      currentBoardSentRef.current = true;
    }
  }, [currentBoard]);

  // Helper function to update transcriptions in cache
  const updateTranscriptionInCache = (
    updater: (prev: Transcription[]) => Transcription[]
  ) => {
    queryClient.setQueryData(
      ["transcriptions", currentBoard?.id],
      (old: Transcription[] = []) => updater(old)
    );
  };

  const navigate = useNavigate();
  const { handleAIEvent } = useAIProcessing();

  useEffect(() => {
    const unsubscribeProcessingStart = EventsOn(
      "ai:processing_start",
      (data: any) => {
        setProcessingState("processing");
        setCurrentAction("Analyzing request...");
      }
    );

    const unsubscribeToolComplete = EventsOn(
      "ai:tool_complete",
      (data: any) => {
        setCurrentAction(`${data.toolName}`);
        // Emit event to trigger board refetch
        EventsEmit("board:refetch");
      }
    );

    const unsubscribeToolError = EventsOn("ai:tool_error", (data: any) => {
      setCurrentAction(`Error: ${data.toolName}`);
      toast.error(`❌ ${data.toolName} failed`, {
        description: data.error,
        duration: 3000,
      });
    });

    const unsubscribeProcessingComplete = EventsOn(
      "ai:processing_complete",
      (data: any) => {
        setProcessingState("complete");
        setCurrentAction(null);
        // Hide the waveform after a brief delay
        setTimeout(() => {
          setProcessingState("idle");
        }, 1000);
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
      if (Array.isArray(data) && data.length > 0) {
        setAudioBars(data);
        addWaveformData(data);
      }
    });

    const unsubscribeRecordingStart = EventsOn("recording:start", () => {
      setIsRecording(true);
      setProcessingState("recording");
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
        setProcessingState("transcribing");
        setCurrentAction("Transcribing audio...");
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
        setProcessingState("idle");
        setCurrentAction(null);
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

    const unsubscribeTranscriptionError = EventsOn(
      "transcription:error",
      (data: string) => {
        setIsRecording(false);
        setAudioBars(null);
        setProcessingState("idle");
        setCurrentAction(null);

        if (data) {
          try {
            const parsed = JSON.parse(data) as {
              id: string;
              error: string;
              message: string;
            };

            updateTranscriptionInCache((prev) =>
              prev.filter((t) => t.id !== parsed.id)
            );

            removePendingTranscription(parsed.id);
            if (parsed.error === "authentication_required") {
              toast.error("Login Required for Cloud Transcription", {
                description: (
                  <span className="text-black">
                    Please sign in to use cloud transcription, or add your own
                    OpenAI API key in settings.
                  </span>
                ),
                duration: 6000,
                action: {
                  label: "Open Settings",
                  onClick: () => navigate("/settings"),
                },
              });
            } else if (parsed.error === "api_key_required") {
              toast.error("API Key Required", {
                description: (
                  <span className="text-black">
                    Please configure your OpenAI API key in Settings to use
                    custom transcription.
                  </span>
                ),
                duration: 6000,
                action: {
                  label: "Add API Key",
                  onClick: () => navigate("/settings"),
                },
              });
            } else {
              toast.error("Transcription Failed", {
                description: parsed.message,
                duration: 5000,
              });
            }
          } catch (e) {
            console.error("Failed to parse transcription:error event data", e);
            toast.error("Transcription Failed", {
              description: "An error occurred during transcription.",
              duration: 5000,
            });
          }
        }
      }
    );

    const unsubscribeUseCloud = EventsOn(
      "transcription:use_cloud",
      async (data: string) => {
        try {
          const parsed = JSON.parse(data) as {
            recording_path: string;
            board_id: string;
          };

          console.log(parsed);

          setProcessingState("transcribing");
          setCurrentAction("Uploading to cloud...");

          const resp = await ReadAudioFile(parsed.recording_path);
          const audioBytes = Uint8Array.from(atob(resp.data.toString()), (c) =>
            c.charCodeAt(0)
          );

          const blob = new Blob([new Uint8Array(audioBytes)], {
            type: "audio/wav",
          });

          const audioFile = new File([blob], "recording.wav", {
            type: "audio/wav",
          });

          await ApiClient.transcribeAndProcessAudio(
            audioFile,
            parsed.board_id,
            handleAIEvent
          );
        } catch (error) {
          console.error("Cloud transcription error:", error);
          setProcessingState("idle");
          setCurrentAction(null);
          setAudioBars(null);
          toast.error("Cloud Transcription Failed", {
            description:
              error instanceof Error
                ? error.message
                : "An unexpected error occurred",
            duration: 5000,
          });
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
      unsubscribeTranscriptionError();
      unsubscribeUseCloud();
    };
  }, [currentBoard, queryClient]);

  const { initialize, teardown } = useCollaborationStore();

  useEffect(() => {
    if (currentBoard?.id) {
      initialize(currentBoard.id);
    }

    return () => {
      teardown();
    };
  }, [currentBoard?.id, initialize, teardown]);

  return (
    <>
      <div className="flex relative">
        <Sidebar />
        <div
          className="w-full transition-all duration-300 min-h-screen"
          style={{ paddingLeft: collapsed ? 72 : 256 }}
        >
          <TopNavbar />

          <div className="pt-14">
            <Outlet />
          </div>
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
      <GlobalWaveform />
      <BoardImportDialog
        open={importDialogOpen}
        onOpenChange={setImportDialogOpen}
        boardId={boardIdToImport}
      />
      <ErrorListener />
    </>
  );
};
