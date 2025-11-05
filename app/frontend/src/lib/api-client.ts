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

    // Add request interceptor to attach token
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

    // Add response interceptor for error handling
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
};
