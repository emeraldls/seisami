import { create } from "zustand";
import { Transcription } from "~/types/types";

const MAX_WAVEFORM_POINTS = 1024;

interface RecordingState {
  isRecording: boolean;
  audioBars: number[] | null;
  waveformData: number[];
  currentPosition: number;
  pendingTranscriptions: Map<string, Transcription>;

  setIsRecording: (recording: boolean) => void;
  setAudioBars: (bars: number[] | null) => void;
  addWaveformData: (bars: number[]) => void;
  resetWaveform: () => void;
  addPendingTranscription: (id: string, transcription: Transcription) => void;
  updatePendingTranscription: (
    id: string,
    updates: Partial<Transcription>
  ) => void;
  removePendingTranscription: (id: string) => void;
  clearPendingTranscriptions: () => void;
}

export const useRecordingStore = create<RecordingState>((set) => ({
  isRecording: false,
  audioBars: null,
  waveformData: [],
  currentPosition: 0,
  pendingTranscriptions: new Map(),

  setIsRecording: (recording) => set({ isRecording: recording }),

  setAudioBars: (bars) => set({ audioBars: bars }),

  addWaveformData: (bars) =>
    set((state) => {
      const combined = [...state.waveformData, ...bars];
      const waveformData =
        combined.length > MAX_WAVEFORM_POINTS
          ? combined.slice(-MAX_WAVEFORM_POINTS)
          : combined;

      return {
        waveformData,
        currentPosition:
          waveformData.length === 0 ? 0 : waveformData.length - 1,
      };
    }),

  resetWaveform: () =>
    set({
      waveformData: [],
      currentPosition: 0,
      audioBars: null,
    }),

  addPendingTranscription: (id, transcription) =>
    set((state) => {
      const newMap = new Map(state.pendingTranscriptions);
      newMap.set(id, transcription);
      return { pendingTranscriptions: newMap };
    }),

  updatePendingTranscription: (id, updates) =>
    set((state) => {
      const newMap = new Map(state.pendingTranscriptions);
      const existing = newMap.get(id);
      if (existing) {
        newMap.set(id, { ...existing, ...updates });
      }
      return { pendingTranscriptions: newMap };
    }),

  removePendingTranscription: (id) =>
    set((state) => {
      const newMap = new Map(state.pendingTranscriptions);
      newMap.delete(id);
      return { pendingTranscriptions: newMap };
    }),

  clearPendingTranscriptions: () => set({ pendingTranscriptions: new Map() }),
}));
