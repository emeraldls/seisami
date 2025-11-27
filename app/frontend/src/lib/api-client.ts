import axios, { AxiosInstance, AxiosRequestConfig } from "axios";
import { useDesktopAuthStore } from "~/stores/auth-store";
import { CLOUD_API_URL } from "./constants";

class ApiClientClass {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: CLOUD_API_URL,
      headers: {
        "Content-Type": "application/json",
      },
    });

    this.client.interceptors.request.use(
      (config) => {
        const token = useDesktopAuthStore.getState().token;
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Token expired or invalid
          //   useDesktopAuthStore.getState().logout();
        }
        return Promise.reject(error);
      }
    );
  }

  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  async post<T>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  async put<T>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }

  async patch<T>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<T> {
    const response = await this.client.patch<T>(url, data, config);
    return response.data;
  }
}

export const apiClient = new ApiClientClass();

export interface BoardMember {
  user_id: string;
  email: string;
  role: "owner" | "member";
  joined_at: string;
}

export interface BoardMetadata {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
  columns_count: number;
  cards_count: number;
  transcriptions_count: number;
}

export type NotificationType = "info" | "in_app" | "external_url";

export interface Notification {
  id: string;
  user_id: string;
  title: string;
  message: string;
  type: NotificationType;
  target: string | null;
  read: boolean;
  created_at: string;
}

export interface NotificationsPaginatedResponse {
  message: string;
  data: Notification[];
  total_count: number;
  total_pages: number;
  current_page: number;
  page_size: number;
}

export const ApiClient = {
  async inviteUserToBoard(email: string, boardId: string): Promise<void> {
    return apiClient.post("/board/invite", {
      email,
      board_id: boardId,
    });
  },

  async removeUserFromBoard(id: string, boardId: string): Promise<void> {
    return apiClient.post("/board/remove", {
      user_id: id,
      board_id: boardId,
    });
  },

  async getBoardMembers(
    boardId: string
  ): Promise<{ data: BoardMember[]; message: string }> {
    return apiClient.get(`/board/${boardId}/members`);
  },

  async getBoardMetadata(
    boardId: string
  ): Promise<{ data: BoardMetadata; message: string }> {
    return apiClient.get(`/board/${boardId}/metadata`);
  },

  async getNotifications(
    offset: number = 0
  ): Promise<NotificationsPaginatedResponse> {
    return apiClient.get(`/notifications/all?offset=${offset}`);
  },

  async markNotificationAsRead(notificationId: string): Promise<void> {
    return apiClient.post(`/notifications/${notificationId}/read`);
  },

  async getConnectedUsers(boardId: string): Promise<{ data: string[] }> {
    return apiClient.get(`/board/${boardId}/connected-users`);
  },

  async transcribeAndProcessAudio(
    audioFile: File,
    boardId: string,
    onEvent: (event: { type: string; data?: any; error?: string }) => void
  ): Promise<void> {
    const token = useDesktopAuthStore.getState().token;
    if (!token) {
      throw new Error("Authentication required");
    }

    const formData = new FormData();
    formData.append("audio", audioFile);
    formData.append("board_id", boardId);

    const response = await fetch(`${CLOUD_API_URL}/ai/transcribe-and-process`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    if (!response.body) {
      throw new Error("No response body");
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    let currentEventType = "";

    try {
      while (true) {
        const { done, value } = await reader.read();

        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        for (const line of lines) {
          if (!line.trim()) continue;

          console.log("SSE Line:", line);

          if (line.startsWith("event: ")) {
            currentEventType = line.slice(7).trim();
          } else if (line.startsWith("data: ")) {
            const data = line.slice(6).trim();

            if (data === "[DONE]") {
              onEvent({ type: "done" });
              return;
            }

            try {
              const parsed = JSON.parse(data);
              onEvent({
                type: currentEventType || "unknown",
                data: parsed,
              });
            } catch (e) {
              console.error("Failed to parse SSE data:", data, e);
            }

            currentEventType = "";
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  },
};
