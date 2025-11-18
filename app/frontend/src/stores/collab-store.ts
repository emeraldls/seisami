import { create } from "zustand";
import { toast } from "sonner";
import { wsService, type CollabResponse } from "../lib/websocket-service";
import { useDesktopAuthStore } from "./auth-store";

export type CollabStatus =
  | "disconnected"
  | "connected"
  | "in-room"
  | "busy"
  | "error"
  | "unauthenticated";

type Unsubscribe = () => void;

interface CollabState {
  status: CollabStatus;
  roomId: string;
  address: string;
  lastError: string | null;
  isInitialized: boolean;
  eventUnsubscribers: Unsubscribe[];
  initialize: (boardId: string) => void;
  reinitialize: (boardId: string) => void;
  teardown: () => void;
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

  reinitialize: (boardId: string) => {
    const state = get();
    if (state.isInitialized) {
      state.teardown();
    }
    state.initialize(boardId);
  },

  initialize: (boardId: string) => {
    if (get().isInitialized) {
      return;
    }

    if (!boardId) {
      set({
        status: "error",
        lastError: "Board ID is required for collaboration",
        isInitialized: true,
      });
      toast.error("Collaboration error", {
        description: "Board ID is required",
      });
      return;
    }

    const authStore = useDesktopAuthStore.getState();
    if (!authStore.isAuthenticated || !authStore.token) {
      return;
    }

    const unsubscribers: Unsubscribe[] = [];

    wsService.setAuthToken(authStore.token);
    wsService.setBoardId(boardId);

    const connectPromise = wsService.connect();

    const unsubConnect = wsService.onConnect(() => {
      // TODO: look into this
      set({
        status: "in-room",
        roomId: boardId,
        address: "127.0.0.1:8080",
      });

      toast.success("Connected to collaboration", {
        description: "Real-time collaboration is active for this board",
      });
    });
    unsubscribers.push(unsubConnect);

    const unsubMessage = wsService.onMessage((message: CollabResponse) => {
      if ("type" in message && message.type === "message") {
        console.log("Broadcast message:", message);
        // Handle broadcast messages here
      }

      if ("error" in message) {
        const errorMsg = message.error;
        set({ status: "error", lastError: errorMsg });
        toast.error("Collaboration error", { description: errorMsg });
      }
    });
    unsubscribers.push(unsubMessage);

    const unsubError = wsService.onError((error: Error) => {
      const message = error.message || "Unknown connection error";

      if (message.includes("Authentication failed")) {
        set({ status: "unauthenticated", lastError: message });
        toast.error("Authentication failed", {
          description: "Your session has expired. Please log in again.",
        });
      } else {
        set({ status: "error", lastError: message });
        // toast.error("Collaboration error", { description: message });
      }
    });
    unsubscribers.push(unsubError);

    const unsubClose = wsService.onClose(() => {
      set({ status: "disconnected", lastError: "Connection closed" });
    });
    unsubscribers.push(unsubClose);

    set({ isInitialized: true, eventUnsubscribers: unsubscribers });

    connectPromise.catch((error) => {
      console.error("Failed to connect:", error);
      set({
        status: "error",
        lastError: error instanceof Error ? error.message : String(error),
      });
    });
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
    wsService.disconnect();
    set({
      eventUnsubscribers: [],
      isInitialized: false,
      status: "disconnected",
      roomId: "",
    });
  },

  setRoomId: (roomId: string) => {
    set({ roomId: normalizeRoomId(roomId) });
  },
}));
