import { useState, useEffect, useRef } from "react";
import { Mic, Copy, Trash2 } from "lucide-react";
import { Button } from "~/components/ui/button";
import { Card, CardContent } from "~/components/ui/card";
import { EventsOn, EventsEmit } from "../../wailsjs/runtime/runtime";
import { GetBoards } from "../../wailsjs/go/main/App";
import { PlannerCalendar } from "~/components/planner-calendar";
import { TranscriptionCard } from "~/components/transcription-card";
import { Transcription } from "~/types/types";
import { useBoardStore } from "~/stores/board-store";

const drawLineSegment = (
  ctx: CanvasRenderingContext2D,
  x: number,
  height: number,
  width: number,
  isEven: boolean
) => {
  ctx.lineWidth = 1;
  ctx.strokeStyle = "#0a0a0a";
  //   ctx.fillStyle = "#0a0a0a";

  const barWidth = Math.min(width * 0.6, 10);
  const barX = x + (width - barWidth) / 2;
  const radius = barWidth / 2;

  ctx.beginPath();

  if (height > 2) {
    ctx.roundRect(barX, -height, barWidth, height, radius);
    ctx.roundRect(barX, 0, barWidth, height, radius);
  } else {
    ctx.fillRect(barX, -height, barWidth, height);
    ctx.fillRect(barX, 0, barWidth, height);
  }

  ctx.stroke();
};

export default function Transcriptions() {
  const [transcriptions, setTranscriptions] = useState<Transcription[]>([]);
  const [isRecording, setIsRecording] = useState(false);
  const [audioBars, setAudioBars] = useState<number[] | null>(null);
  const [smoothedBars, setSmoothedBars] = useState<number[] | null>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [startDrawing, setStartDrawing] = useState(false);
  const animationRef = useRef<number>(0);
  const lastUpdateTime = useRef<number>(0);

  const { currentBoard } = useBoardStore();

  useEffect(() => {
    const unsubscribeBoards = () => {
      async function fetchBoards() {
        try {
          const boardList = await GetBoards(1, 10);
          console.log(boardList);
        } catch (err) {
          console.error(err);
        }
      }

      fetchBoards();
    };

    return () => {
      unsubscribeBoards();
    };
  }, []);

  useEffect(() => {
    if (!audioBars) return;

    const smoothingFactor = 0.1;

    setSmoothedBars((prevSmoothedBars) => {
      if (!prevSmoothedBars) {
        return [...audioBars];
      }

      return prevSmoothedBars.map((current, index) => {
        const target = audioBars[index] || 0;
        return current + (target - current) * smoothingFactor;
      });
    });
  }, [audioBars]);

  useEffect(() => {
    const draw = () => {
      if (!startDrawing) {
        console.log("start drawing not enabled yet");
        return;
      }

      if (!smoothedBars) {
        console.warn("no audio bars to draw canvas");
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
      const padding = 10;
      const rect = canvas.getBoundingClientRect();
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;

      ctx.scale(dpr, dpr);
      ctx.clearRect(0, 0, rect.width, rect.height);
      ctx.translate(0, rect.height / 2);

      const maxBars = Math.min(smoothedBars.length, Math.floor(rect.width / 6));
      const barsToShow = smoothedBars.slice(0, maxBars);
      const width = rect.width / barsToShow.length + 4;

      for (let i = 0; i < barsToShow.length; i++) {
        const x = width * i;
        let height = barsToShow[i] * (rect.height / 2 - padding);

        height = Math.max(height, 2);

        if (height > rect.height / 2 - padding) {
          height = rect.height / 2 - padding;
        }
        drawLineSegment(ctx, x, height, width, (i + 1) % 2 === 0);
      }
    };

    const animate = () => {
      const now = performance.now();
      if (now - lastUpdateTime.current > 33) {
        // ~30fps
        draw();
        lastUpdateTime.current = now;
      }

      if (startDrawing) {
        animationRef.current = requestAnimationFrame(animate);
      }
    };

    if (startDrawing) {
      animate();
    }

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [smoothedBars, startDrawing]);

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

    const unsubscribeAudioBars = EventsOn("audio_bars", (data) => {
      if (Array.isArray(data) && data.length > 0) {
        console.log(data);
        setAudioBars(data);
        setStartDrawing(true);
      }
    });

    const unsubscribeRecordingStart = EventsOn("recording:start", () => {
      setIsRecording(true);
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
        setSmoothedBars(null);
        setStartDrawing(false);
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
        className="w-full h-40 border-b"
        style={{ display: isRecording || startDrawing ? "block" : "none" }}
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
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
