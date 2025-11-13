import { create } from "zustand";
import { persist } from "zustand/middleware";
import { SetLoginToken, ClearLoginToken } from "../../wailsjs/go/main/App";

export interface DesktopAuthState {
  isAuthenticated: boolean;
  token: string | null;
  userId: string | null;
  email: string | null;
  isLoading: boolean;
  error: string | null;

  setToken: (token: string, userId: string, email: string) => void;
  logout: (clearLocalData?: boolean) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
}

export const useDesktopAuthStore = create<DesktopAuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      token: null,
      userId: null,
      email: null,
      isLoading: false,
      error: null,

      setToken: (token: string, userId: string, email: string) => {
        set({
          isAuthenticated: true,
          token,
          userId,
          email,
          error: null,
        });

        SetLoginToken(token).catch((err) => {
          console.error("Failed to sync token with backend:", err);
        });
      },

      logout: (clearLocalData = false) => {
        set({
          isAuthenticated: false,
          token: null,
          userId: null,
          email: null,
        });

        ClearLoginToken().catch((err) => {
          console.error("Failed to clear token from backend:", err);
        });

        if (clearLocalData) {
          try {
            localStorage.removeItem("desktop-auth-storage");
            localStorage.removeItem("board-storage");
            console.log("All local data has been cleared");
          } catch (err) {
            console.error("Failed to clear local storage:", err);
          }
        }
      },

      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      },

      setError: (error: string | null) => {
        set({ error });
      },

      clearError: () => {
        set({ error: null });
      },
    }),
    {
      name: "desktop-auth-storage",
      partialize: (state) => ({
        token: state.token,
        userId: state.userId,
        email: state.email,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
