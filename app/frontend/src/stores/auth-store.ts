import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface DesktopAuthState {
  isAuthenticated: boolean;
  token: string | null;
  userId: string | null;
  email: string | null;
  isLoading: boolean;
  error: string | null;

  setToken: (token: string, userId: string, email: string) => void;
  logout: () => void;
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
      },

      logout: () => {
        set({
          isAuthenticated: false,
          token: null,
          userId: null,
          email: null,
        });
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
