export interface Transcription {
  id: string;
  text: string;
  timestamp: Date;
  isTranscribing?: boolean;
  duration?: number;
  wordCount?: number;
}

export type TicketType = "bug" | "feature" | "enhancement";

export interface Column {
  id: string; // UUID
  boardId: string; // FK -> Board
  title: string; // Column title
  createdAt: string; // ISO datetime
  updatedAt: string; // ISO datetime
}

export interface Ticket {
  id: string; // UUID
  columnId: string; // FK -> Column
  title: string; // Card title
  description?: string; // Optional description
  assigneeId?: number; // Optional assignee
  storyPoints?: number; // Optional story points
  prLink?: string; // Optional PR link
  ticketType: TicketType; // bug | feature | enhancement
  createdAt: string; // ISO datetime
  updatedAt: string; // ISO datetime
}

export interface BoardResponse {
  ID: string;
  CreatedAt: StringNull;
  UpdatedAt: StringNull;
  Name: string;
}

export interface ColumnResponse {
  BoardID: string;
  CreatedAt: StringNull;
  ID: string;
  Position: number;
  Title: string;
  UpdatedAt: StringNull;
}

type StringNull = {
  String: string | null;
} & {
  Valid: boolean;
};

type NumberNull = {
  Int64: number | null;
} & {
  Valid: boolean;
};

export interface TicketResponse {
  ID: string;
  ColumnID: string;
  Title: string;
  Description: StringNull;
  AssigneeID: NumberNull;
  StoryPoints: NumberNull;
  PrLink: StringNull;
  TicketType: string;
  CreatedAt: StringNull;
  UpdatedAt: StringNull;
}
