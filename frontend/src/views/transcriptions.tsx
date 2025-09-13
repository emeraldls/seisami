import { useState, useEffect, useRef } from "react";
import { Mic, Copy, Trash2 } from "lucide-react";
import { Button } from "~/components/ui/button";
import { Card, CardContent } from "~/components/ui/card";
import { EventsOn, EventsEmit } from "../../wailsjs/runtime/runtime";
import { GetTranscriptions } from "../../wailsjs/go/main/App";
import { PlannerCalendar } from "~/components/planner-calendar";
import { TranscriptionCard } from "~/components/transcription-card";
import { TranscriptionDetailModal } from "~/components/transcription-detail-modal";
import { Transcription } from "~/types/types";
import { useBoardStore } from "~/stores/board-store";

const drawWaveform = (
  ctx: CanvasRenderingContext2D,
  waveformData: number[],
  currentPosition: number,
  canvasWidth: number,
  canvasHeight: number
) => {
  const maxHeight = canvasHeight * 0.8;
  const barWidth = 2;
  const barSpacing = 1;
  const totalBarWidth = barWidth + barSpacing;

  // Draw past waveform (recorded audio)
  ctx.fillStyle = "#3b82f6"; // Blue for recorded
  for (let i = 0; i < currentPosition && i < waveformData.length; i++) {
    const x = i * totalBarWidth;
    const height = Math.max(waveformData[i] * maxHeight, 2);
    const y = (canvasHeight - height) / 2;

    ctx.fillRect(x, y, barWidth, height);
  }

  // Draw current position indicator
  if (currentPosition < waveformData.length) {
    ctx.fillStyle = "#ef4444"; // Red for current
    const x = currentPosition * totalBarWidth;
    const height = Math.max(waveformData[currentPosition] * maxHeight, 2);
    const y = (canvasHeight - height) / 2;

    ctx.fillRect(x, y, barWidth, height);
  }

  // Draw future waveform (faded)
  ctx.fillStyle = "#e5e7eb"; // Gray for future
  for (let i = currentPosition + 1; i < waveformData.length; i++) {
    const x = i * totalBarWidth;
    const height = Math.max(waveformData[i] * maxHeight, 2);
    const y = (canvasHeight - height) / 2;

    ctx.fillRect(x, y, barWidth, height);
  }
};

export default function Transcriptions() {
  const [transcriptions, setTranscriptions] = useState<Transcription[]>([]);
  const [isRecording, setIsRecording] = useState(false);
  const [audioBars, setAudioBars] = useState<number[] | null>(null);
  const [waveformData, setWaveformData] = useState<number[]>([]);
  const [currentPosition, setCurrentPosition] = useState(0);
  const [selectedTranscription, setSelectedTranscription] =
    useState<Transcription | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animationRef = useRef<number>(0);
  const lastUpdateTime = useRef<number>(0);

  const { currentBoard } = useBoardStore();

  useEffect(() => {
    const fetchTranscriptions = async () => {
      if (currentBoard) {
        try {
          const transcriptions = await GetTranscriptions(
            currentBoard.id,
            1,
            50
          );
          const formattedTranscriptions: Transcription[] = transcriptions.map(
            (t) => ({
              id: t.ID,
              text: t.Transcription,
              timestamp: new Date(t.CreatedAt.String || new Date()),
              isTranscribing: false,
              intent: t.Intent?.String || undefined,
              assistantResponse: t.AssistantResponse?.String || undefined,
              recordingPath: t.RecordingPath?.String || undefined,
            })
          );
          setTranscriptions(formattedTranscriptions);
        } catch (error) {
          console.error("Failed to fetch transcriptions:", error);
        }
      }
    };

    fetchTranscriptions();
  }, [currentBoard]);

  useEffect(() => {
    if (!audioBars) return;

    // Add new audio data to waveform
    setWaveformData((prev) => [...prev, ...audioBars]);
    setCurrentPosition((prev) => prev + audioBars.length);
  }, [audioBars]);

  useEffect(() => {
    const draw = () => {
      if (!isRecording || waveformData.length === 0) {
        return;
      }

      if (!canvasRef.current) {
        console.error("canvas not found");
        return;
      }

      const canvas = canvasRef.current;
      const ctx = canvas.getContext("2d");
      if (!ctx) {
        console.error("browser doesnt support canvas");
        return;
      }

      const dpr = window.devicePixelRatio || 1;
      const rect = canvas.getBoundingClientRect();
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;

      ctx.scale(dpr, dpr);
      ctx.clearRect(0, 0, rect.width, rect.height);

      // Draw the waveform
      drawWaveform(ctx, waveformData, currentPosition, rect.width, rect.height);
    };

    const animate = () => {
      const now = performance.now();
      if (now - lastUpdateTime.current > 33) {
        // ~30fps
        draw();
        lastUpdateTime.current = now;
      }

      if (isRecording) {
        animationRef.current = requestAnimationFrame(animate);
      }
    };

    if (isRecording) {
      animate();
    }

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [waveformData, currentPosition, isRecording]);

  useEffect(() => {
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
              setTranscriptions((prev) =>
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

            setTranscriptions((prev) => {
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
      }
    });

    const unsubscribeRecordingStart = EventsOn("recording:start", () => {
      setIsRecording(true);
      setWaveformData([]);
      setCurrentPosition(0);
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
        setWaveformData([]);
        setCurrentPosition(0);
        if (data) {
          try {
            const parsed = JSON.parse(data) as { id: string };
            const newTranscription: Transcription = {
              id: parsed.id,
              text: "Transcribing...",
              timestamp: new Date(),
              isTranscribing: true,
            };
            setTranscriptions((prev) => [newTranscription, ...prev]);
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
            setTranscriptions((prev) => prev.filter((t) => t.id !== parsed.id));
          } catch (e) {
            console.error("Failed to parse transcription:short event data", e);
          }
        }
      }
    );

    return () => {
      unsubscribeTranscription();
      unsubscribeStructuredResponse();
      unsubscribeRecordingStart();
      unsubscribeRecordingStop();
      unsubscribeTranscriptionShort();
      unsubscribeAudioBars();
    };
  }, []);

  const deleteTranscription = (id: string) => {
    setTranscriptions((prev) => prev.filter((t) => t.id !== id));
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
      <canvas
        ref={canvasRef}
        className="w-full h-20 border-b bg-gray-50"
        style={{ display: isRecording ? "block" : "none" }}
      ></canvas>
      {/* <div className="p-4">
        <PlannerCalendar />
      </div> */}

      <div className="flex-1 p-8 overflow-auto">
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold ">History</h3>
              <p className="text-sm text-neutral-500 ">
                {transcriptions.length} transcripts
              </p>
            </div>
          </div>

          {transcriptions.length === 0 ? (
            <div>Nothing yet boy</div>
          ) : (
            <div className="space-y-4">
              {transcriptions.map((transcription) => (
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
