import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface WebAuthState {
  token: string | null;
  userId: string | null;
  email: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  setAuth: (token: string, userId: string, email: string) => void;
  logout: () => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
}

export const useAuthStore = create<WebAuthState>()(
  persist(
    (set) => ({
      token: null,
      userId: null,
      email: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      setAuth: (token: string, userId: string, email: string) => {
        set({
          token,
          userId,
          email,
          isAuthenticated: true,
          error: null,
        });
      },

      logout: () => {
        set({
          token: null,
          userId: null,
          email: null,
          isAuthenticated: false,
          error: null,
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
      name: "auth-storage",
      partialize: (state) => ({
        token: state.token,
        userId: state.userId,
        email: state.email,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
