import { useEffect, useRef } from "react";
import { useRecordingStore } from "~/stores/recording-store";
import { X } from "lucide-react";
import { Button } from "~/components/ui/button";

const BAR_WIDTH = 2;
const BAR_SPACING = 1;
const TOTAL_BAR_WIDTH = BAR_WIDTH + BAR_SPACING;
const MIN_BAR_HEIGHT = 2;

const drawWaveform = (
  ctx: CanvasRenderingContext2D,
  waveformData: number[],
  currentPosition: number,
  canvasWidth: number,
  canvasHeight: number
) => {
  ctx.fillStyle = "#ffffff";
  ctx.fillRect(0, 0, canvasWidth, canvasHeight);

  ctx.strokeStyle = "#000000";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(0, canvasHeight / 2);
  ctx.lineTo(canvasWidth, canvasHeight / 2);
  ctx.stroke();

  if (waveformData.length === 0) {
    return;
  }

  const maxHeight = canvasHeight * 0.8;
  const maxVisibleBars = Math.max(1, Math.floor(canvasWidth / TOTAL_BAR_WIDTH));
  const clampedCurrent = Math.min(
    Math.max(currentPosition, 0),
    waveformData.length - 1
  );
  const startIndex = Math.max(0, clampedCurrent - maxVisibleBars + 1);
  const endIndex = Math.min(waveformData.length, startIndex + maxVisibleBars);
  const visibleSpan = endIndex - startIndex;

  for (let i = startIndex; i < endIndex; i++) {
    const amplitude = Math.min(Math.max(Math.abs(waveformData[i]), 0), 1);
    const height = Math.max(amplitude * maxHeight, MIN_BAR_HEIGHT);
    const x = (i - startIndex) * TOTAL_BAR_WIDTH;
    const y = (canvasHeight - height) / 2;

    if (i === clampedCurrent) {
      ctx.fillStyle = "#000000";
    } else {
      const distance = clampedCurrent - i;
      const fade = 1 - Math.min(Math.abs(distance) / visibleSpan, 1);
      const alpha = 0.25 + fade * 0.65;
      ctx.fillStyle = `rgba(0, 0, 0, ${alpha.toFixed(2)})`;
    }

    ctx.fillRect(x, y, BAR_WIDTH, height);
  }
};

export const GlobalWaveform = () => {
  const {
    isRecording,
    waveformData,
    currentPosition,
    setIsRecording,
    processingState,
    currentAction,
    setProcessingState,
  } = useRecordingStore();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animationRef = useRef<number | undefined>(undefined);
  const waveformRef = useRef<number[]>(waveformData);
  const positionRef = useRef<number>(currentPosition);

  useEffect(() => {
    waveformRef.current = waveformData;
  }, [waveformData]);

  useEffect(() => {
    positionRef.current = currentPosition;
  }, [currentPosition]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let animationInterval: number | undefined;

    if (
      processingState === "transcribing" ||
      processingState === "processing"
    ) {
      animationInterval = setInterval(() => {
        const randomAmplitude = 0.1 + Math.random() * 0.15;
        useRecordingStore.getState().addWaveformData([randomAmplitude]);
      }, 100);
    }

    const render = () => {
      const dpr = window.devicePixelRatio || 1;
      const rect = canvas.getBoundingClientRect();
      const targetWidth = rect.width * dpr;
      const targetHeight = rect.height * dpr;

      if (canvas.width !== targetWidth || canvas.height !== targetHeight) {
        canvas.width = targetWidth;
        canvas.height = targetHeight;
      }

      ctx.setTransform(1, 0, 0, 1, 0, 0);
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.scale(dpr, dpr);

      drawWaveform(
        ctx,
        waveformRef.current,
        positionRef.current,
        rect.width,
        rect.height
      );

      const state = useRecordingStore.getState().processingState;
      if (
        state === "recording" ||
        state === "transcribing" ||
        state === "processing"
      ) {
        animationRef.current = requestAnimationFrame(render);
      }
    };

    if (
      processingState === "recording" ||
      processingState === "transcribing" ||
      processingState === "processing"
    ) {
      render();
    } else {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
    }

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
        animationRef.current = undefined;
      }
      if (animationInterval) {
        clearInterval(animationInterval);
      }
    };
  }, [processingState]);

  if (processingState === "idle" || processingState === "complete") return null;

  const getStatusText = () => {
    switch (processingState) {
      case "recording":
        return "REC";
      case "transcribing":
        return "TRANSCRIBING";
      case "processing":
        return currentAction || "PROCESSING";
      default:
        return "REC";
    }
  };

  const handleClose = () => {
    if (processingState === "recording") {
      setIsRecording(false);
    }
    setProcessingState("idle");
  };

  return (
    <div className="fixed top-16 right-6 z-50 w-80 border border-black bg-white shadow-sm">
      <div className="flex items-center justify-between px-3 py-2 border-b border-black bg-black text-white">
        <span className="text-xs font-semibold tracking-widest">
          {getStatusText()}
        </span>
        {processingState === "recording" && (
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6 text-white hover:bg-white/10"
            onClick={handleClose}
          >
            <X className="h-3 w-3" />
          </Button>
        )}
      </div>
      <canvas
        ref={canvasRef}
        className="w-full h-20 bg-white"
        style={{ display: "block" }}
      />
    </div>
  );
};
