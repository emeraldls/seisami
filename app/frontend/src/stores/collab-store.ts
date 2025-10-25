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
  initialize: () => void;
  teardown: () => void;
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

    const authStore = useDesktopAuthStore.getState();
    if (!authStore.isAuthenticated || !authStore.token) {
      set({
        status: "unauthenticated",
        lastError: "Please log in to use collaboration features",
        isInitialized: true,
      });
      toast.error("Authentication required", {
        description: "Please log in to use collaboration features",
      });
      return;
    }

    const unsubscribers: Unsubscribe[] = [];

    wsService.setAuthToken(authStore.token);

    const connectPromise = wsService.connect();

    const unsubConnect = wsService.onConnect(() => {
      set((state) => ({
        status: state.roomId ? "in-room" : "connected",
        address: "127.0.0.1:8080",
      }));
      toast.success("Connected to collaboration server", {
        description: "Ready to create or join rooms",
      });
    });
    unsubscribers.push(unsubConnect);

    const unsubMessage = wsService.onMessage((message: CollabResponse) => {
      if ("status" in message && message.status === "created") {
        const roomId = message.roomId;
        set({ status: "in-room", roomId, lastError: null });
        toast.success("Room created", {
          description: `Share this ID: ${roomId}`,
        });
      }

      if ("status" in message && message.status === "joined") {
        const roomId = message.roomId;
        set({ status: "in-room", roomId, lastError: null });
        toast.success("Joined room", {
          description: `Connected to room ${roomId}`,
        });
      }

      if ("status" in message && message.status === "left") {
        const roomId = message.roomId;
        set({ status: "connected", roomId: "" });
        toast.info("Left room", {
          description: `You left room ${roomId}`,
        });
      }

      if ("type" in message && message.type === "message") {
        console.log("Broadcast message:", message);
        // You can handle broadcast messages here
        // For now, just log them
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
        toast.error("Collaboration error", { description: message });
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
    set({ eventUnsubscribers: [], isInitialized: false });
  },

  createRoom: async () => {
    set({ status: "busy", lastError: null });
    try {
      wsService.createRoom();

      const roomId = await new Promise<string>((resolve, reject) => {
        const timeout = setTimeout(() => {
          unsubscribe();
          reject(new Error("Room creation timeout"));
        }, 5000);

        const unsubscribe = wsService.onMessage((message) => {
          if ("status" in message && message.status === "created") {
            clearTimeout(timeout);
            unsubscribe();
            resolve(message.roomId);
          }
        });
      });

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
      wsService.joinRoom(targetRoomId);

      const joinedRoomId = await new Promise<string>((resolve, reject) => {
        const timeout = setTimeout(() => {
          unsubscribe();
          reject(new Error("Join room timeout"));
        }, 5000);

        const unsubscribe = wsService.onMessage((message) => {
          if ("status" in message && message.status === "joined") {
            clearTimeout(timeout);
            unsubscribe();
            resolve(message.roomId);
          }
        });
      });

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
      wsService.leaveRoom(currentRoomId);

      const leftRoomId = await new Promise<string>((resolve, reject) => {
        const timeout = setTimeout(() => {
          unsubscribe();
          reject(new Error("Leave room timeout"));
        }, 5000);

        const unsubscribe = wsService.onMessage((message) => {
          if ("status" in message && message.status === "left") {
            clearTimeout(timeout);
            unsubscribe();
            resolve(message.roomId);
          }
        });
      });

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
