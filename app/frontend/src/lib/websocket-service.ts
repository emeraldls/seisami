/**
 * WebSocket Service for Real-time Collaboration
 * Handles connection, message sending/receiving, and reconnection logic
 */

import { useDesktopAuthStore } from "~/stores/auth-store";
import { WEBSOCKET_URL } from "./constants";

export type CollabMessage =
  | {
      action: "create" | "join" | "leave" | "broadcast";
      roomId?: string;
      data?: string;
      type?: string;
    }
  | Record<string, unknown>;

export type CollabResponse =
  | {
      status: "created" | "joined" | "left";
      roomId: string;
    }
  | {
      type: string;
      from: string;
      data: string;
    }
  | {
      error: string;
    };

type MessageHandler = (message: CollabResponse) => void;
type ConnectionHandler = () => void;
type ErrorHandler = (error: Error) => void;
type CloseHandler = () => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private authToken: string | null = null;
  private boardId: string | null = null;
  private messageHandlers: Set<MessageHandler> = new Set();
  private connectionHandlers: Set<ConnectionHandler> = new Set();
  private errorHandlers: Set<ErrorHandler> = new Set();
  private closeHandlers: Set<CloseHandler> = new Set();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // ms
  private isIntentionallyClosed = false;

  constructor(url: string = WEBSOCKET_URL) {
    this.url = url;
  }

  /**
   * Set the authentication token before connecting
   */
  setAuthToken(token: string): void {
    this.authToken = token;
  }

  /**
   * Set the board ID for the collaboration session
   */
  setBoardId(boardId: string): void {
    this.boardId = boardId;
  }

  /**
   * Connect to the WebSocket server
   */
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // Build WebSocket URL with auth token and board_id as query parameters
        const wsUrl = new URL(this.url);
        if (this.authToken) {
          wsUrl.searchParams.append("token", this.authToken);
        }
        if (this.boardId) {
          wsUrl.searchParams.append("board_id", this.boardId);
        }

        this.ws = new WebSocket(wsUrl.toString());
        this.isIntentionallyClosed = false;

        this.ws.onopen = () => {
          console.log("WebSocket connected");
          this.reconnectAttempts = 0;
          this.connectionHandlers.forEach((handler) => handler());
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: CollabResponse = JSON.parse(event.data);
            this.messageHandlers.forEach((handler) => handler(message));
          } catch (error) {
            console.error("Error parsing WebSocket message:", error);
          }
        };

        this.ws.onerror = (event) => {
          console.error("WebSocket error:", event);
          const error = new Error("WebSocket error occurred");
          this.errorHandlers.forEach((handler) => handler(error));
          reject(error);
        };

        this.ws.onclose = (event) => {
          console.log("WebSocket closed", event.code, event.reason);
          this.ws = null;

          // Check if it's an authentication error
          if (event.code === 1008) {
            // Policy Violation - typically used for auth errors
            const error = new Error("Authentication failed or token expired");
            this.errorHandlers.forEach((handler) => handler(error));
            this.isIntentionallyClosed = true; // Don't reconnect on auth failure
          }

          this.closeHandlers.forEach((handler) => handler());

          // Attempt reconnection if not intentionally closed or auth error
          if (!this.isIntentionallyClosed) {
            this.attemptReconnect();
          }
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect(): void {
    this.isIntentionallyClosed = true;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Send a message to the server
   */
  send(message: CollabMessage): void {
    if (!this.authToken) {
      console.error("Cannot send message: Auth token is not set");
      return;
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket is not connected. Attempting to connect...");
      this.connect()
        .then(() => {
          this.sendMessage(message);
        })
        .catch((error) => {
          console.error("Failed to reconnect and send message:", error);
          this.errorHandlers.forEach((handler) => handler(error));
        });
      return;
    }

    this.sendMessage(message);
  }

  /**
   * Create a new collaboration room
   */
  createRoom(): void {
    this.send({ action: "create" });
  }

  /**
   * Join an existing collaboration room
   */
  joinRoom(roomId: string): void {
    this.send({ action: "join", roomId });
  }

  /**
   * Leave a collaboration room
   */
  leaveRoom(roomId: string): void {
    this.send({ action: "leave", roomId });
  }

  /**
   * Broadcast a message to all clients in the room
   */
  broadcast(roomId: string, data: string): void {
    this.send({ action: "broadcast", roomId, data });
  }

  /**
   * Register a handler for incoming messages
   */
  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  /**
   * Register a handler for connection events
   */
  onConnect(handler: ConnectionHandler): () => void {
    this.connectionHandlers.add(handler);
    return () => this.connectionHandlers.delete(handler);
  }

  /**
   * Register a handler for error events
   */
  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler);
    return () => this.errorHandlers.delete(handler);
  }

  /**
   * Register a handler for close events
   */
  onClose(handler: CloseHandler): () => void {
    this.closeHandlers.add(handler);
    return () => this.closeHandlers.delete(handler);
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  /**
   * Attempt to reconnect
   */
  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error(
        "Max reconnection attempts reached. Giving up on auto-reconnect."
      );
      const error = new Error("Failed to reconnect after maximum attempts");
      this.errorHandlers.forEach((handler) => handler(error));
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    console.log(
      `Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
    );

    setTimeout(() => {
      this.connect().catch((error) => {
        console.error("Reconnection failed:", error);
      });
    }, delay);
  }

  /**
   * Send the actual message
   */
  private sendMessage(message: CollabMessage): void {
    try {
      this.ws!.send(JSON.stringify(message));
    } catch (error) {
      console.error("Error sending message:", error);
      this.errorHandlers.forEach((handler) =>
        handler(error instanceof Error ? error : new Error(String(error)))
      );
    }
  }
}

// Create and export a singleton instance
export const wsService = new WebSocketService();

export default WebSocketService;
