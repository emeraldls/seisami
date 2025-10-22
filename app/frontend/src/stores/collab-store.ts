import { create } from "zustand";
import { toast } from "sonner";
import {
  CreateCollaborationRoom,
  JoinCollaborationRoom,
  LeaveCollaborationRoom,
  GetCollaborationRoomId,
} from "../../wailsjs/go/main/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";

export type CollabStatus =
  | "disconnected"
  | "connected"
  | "in-room"
  | "busy"
  | "error";

type Unsubscribe = () => void;

type CollabEventPayload = {
  roomId?: string;
  address?: string;
  error?: string;
  [key: string]: unknown;
};

interface CollabState {
  status: CollabStatus;
  roomId: string;
  address: string;
  lastError: string | null;
  isInitialized: boolean;
  eventUnsubscribers: Unsubscribe[];
  initialize: () => void;
  teardown: () => void;
  loadCurrentRoom: () => Promise<void>;
  createRoom: () => Promise<string | null>;
  joinRoom: (roomId: string) => Promise<string | null>;
  leaveRoom: () => Promise<string | null>;
  setRoomId: (roomId: string) => void;
}

const normalizeRoomId = (value: string) => value.trim();

export const useCollaborationStore = create<CollabState>((set, get) => ({
  status: "disconnected",
  roomId: "",
  address: "",
  lastError: null,
  isInitialized: false,
  eventUnsubscribers: [],

  initialize: () => {
    if (get().isInitialized) {
      return;
    }

    const unsubscribers: Unsubscribe[] = [
      EventsOn("collab:connected", (data: CollabEventPayload = {}) => {
        set((state) => ({
          status: state.roomId ? "in-room" : "connected",
          address: data.address ?? state.address,
        }));
        if (data.address) {
          toast.success("Connected to collaboration server", {
            description: `Address: ${data.address}`,
          });
        }
      }),
      EventsOn("collab:room_created", (data: CollabEventPayload = {}) => {
        if (!data.roomId) return;
        set({ status: "in-room", roomId: data.roomId, lastError: null });
        toast.success("Room created", {
          description: `Share this ID: ${data.roomId}`,
        });
      }),
      EventsOn("collab:joined", (data: CollabEventPayload = {}) => {
        if (!data.roomId) return;
        set({ status: "in-room", roomId: data.roomId, lastError: null });
        toast.success("Joined room", {
          description: `Connected to room ${data.roomId}`,
        });
      }),
      EventsOn("collab:left", (data: CollabEventPayload = {}) => {
        set({ status: "connected", roomId: "" });
        if (data.roomId) {
          toast.info("Left room", {
            description: `You left room ${data.roomId}`,
          });
        }
      }),
      EventsOn("collab:error", (data: CollabEventPayload = {}) => {
        const message =
          typeof data.error === "string" && data.error.length > 0
            ? data.error
            : "Unknown collaboration error";
        set({ status: "error", lastError: message });
        toast.error("Collaboration error", { description: message });
      }),
    ];

    set({ isInitialized: true, eventUnsubscribers: unsubscribers });
    void get().loadCurrentRoom();
  },

  teardown: () => {
    const { eventUnsubscribers } = get();
    eventUnsubscribers.forEach((unsubscribe) => {
      try {
        unsubscribe();
      } catch (error) {
        console.error("Failed to unsubscribe from collaboration event", error);
      }
    });
    set({ eventUnsubscribers: [], isInitialized: false });
  },

  loadCurrentRoom: async () => {
    try {
      const currentRoomId = await GetCollaborationRoomId();
      set((state) => ({
        roomId: currentRoomId ?? "",
        status: currentRoomId
          ? "in-room"
          : state.status === "disconnected"
          ? "connected"
          : state.status,
      }));
    } catch (error) {
      console.error("Failed to load collaboration room", error);
      set({ lastError: (error as Error).message, status: "error" });
    }
  },

  createRoom: async () => {
    set({ status: "busy", lastError: null });
    try {
      const roomId = await CreateCollaborationRoom();
      const normalized = normalizeRoomId(roomId);
      set({ roomId: normalized, status: "in-room" });
      return normalized;
    } catch (error) {
      const message = (error as Error).message ?? "Unable to create room";
      set({ status: "error", lastError: message });
      toast.error("Failed to create room", { description: message });
      return null;
    }
  },

  joinRoom: async (roomId: string) => {
    const targetRoomId = normalizeRoomId(roomId);
    if (!targetRoomId) {
      const message = "Room ID is required";
      set({ lastError: message, status: "error" });
      toast.error("Unable to join room", { description: message });
      return null;
    }

    set({ status: "busy", lastError: null });

    try {
      const joinedRoomId = await JoinCollaborationRoom(targetRoomId);
      const normalized = normalizeRoomId(joinedRoomId);
      set({ roomId: normalized, status: "in-room" });
      toast.success("Joined room", {
        description: `Connected to ${normalized}`,
      });
      return normalized;
    } catch (error) {
      const message = (error as Error).message ?? "Unable to join room";
      set({ status: "error", lastError: message });
      toast.error("Failed to join room", { description: message });
      return null;
    }
  },

  leaveRoom: async () => {
    const currentRoomId = get().roomId;
    if (!currentRoomId) {
      return null;
    }

    set({ status: "busy", lastError: null });

    try {
      const leftRoomId = await LeaveCollaborationRoom(currentRoomId);
      set({ roomId: "", status: "connected" });
      toast.info("Left room", { description: `You left ${leftRoomId}` });
      return leftRoomId;
    } catch (error) {
      const message = (error as Error).message ?? "Unable to leave room";
      set({ status: "error", lastError: message });
      toast.error("Failed to leave room", { description: message });
      return null;
    }
  },

  setRoomId: (roomId: string) => {
    set({ roomId: normalizeRoomId(roomId) });
  },
}));
