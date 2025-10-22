export interface Transcription {
  id: string;
  text: string;
  timestamp: Date;
  isTranscribing?: boolean;
  duration?: number;
  wordCount?: number;
  intent?: string;
  assistantResponse?: string;
  recordingPath?: string;
}
