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

export interface BoardEventData {
  room_id?: string;
  id: string;
  name: string;
  updated_at?: string;
}

export interface ColumnEventData {
  room_id: string;
  id: string;
  board_id: string;
  name: string;
  position: number;
  created_at?: string;
  updated_at?: string;
}

export interface ColumnDeleteEventData {
  room_id: string;
  id: string;
  board_id: string;
  name: string;
  position: number;
  card_ids?: string[];
}

export interface CardEventData {
  column: {
    room_id: string;
    id: string;
    board_id: string;
    name: string;
    position: number;
  };
  card: {
    id: string;
    name: string;
    description: string;
    column_id: string;
    index: number;
    created_at?: string;
    updated_at?: string;
  };
}

export interface CardDeleteEventData {
  room_id: string;
  column: {
    id: string;
    board_id: string;
    name: string;
    position: number;
  };
  card: {
    id: string;
    column_id: string;
    index: number;
  };
}

export interface CardColumnEventData {
  room_id: string;
  card_id: string;
  old_column: {
    id: string;
    name: string;
    position: number;
  } | null;
  new_column: {
    id: string;
    board_id: string;
    name: string;
    position: number;
  };
  index: number;
}

export type WebSocketEvent =
  | { type: "board:data"; data: BoardEventData }
  | { type: "column:create"; data: ColumnEventData }
  | { type: "column:data"; data: ColumnEventData }
  | { type: "column:delete"; data: ColumnDeleteEventData }
  | { type: "card:create"; data: CardEventData }
  | { type: "card:data"; data: CardEventData }
  | { type: "card:delete"; data: CardDeleteEventData }
  | { type: "card:column"; data: CardColumnEventData };
