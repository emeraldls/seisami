import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";
import {
  CreateBoard,
  GetBoards,
  GetBoardByID,
  UpdateBoard,
  DeleteBoard,
} from "../../wailsjs/go/main/App";
import { EventsEmit } from "../../wailsjs/runtime/runtime";
import { useCollaborationStore } from "./collab-store";

// Helper interface for normalized board data used in components
export interface NormalizedBoard {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

interface BoardState {
  boards: NormalizedBoard[];
  currentBoard: NormalizedBoard | null;
  isLoading: boolean;
  error: string | null;
  hasCompletedOnboarding: boolean;

  fetchBoards: (page?: number, pageSize?: number) => Promise<void>;
  createBoard: (name: string) => Promise<NormalizedBoard | null>;
  updateBoard: (
    boardId: string,
    name: string
  ) => Promise<NormalizedBoard | null>;
  deleteBoard: (boardId: string) => Promise<boolean>;
  selectBoard: (boardId: string) => Promise<void>;
  setCurrentBoard: (board: NormalizedBoard | null) => void;
  setHasCompletedOnboarding: (completed: boolean) => void;
  clearError: () => void;
}

export const useBoardStore = create<BoardState>()(
  devtools(
    persist(
      (set, get) => ({
        boards: [],
        currentBoard: null,
        isLoading: false,
        error: null,
        hasCompletedOnboarding: false,

        fetchBoards: async (page = 1, pageSize = 10) => {
          set({ isLoading: true, error: null });
          try {
            const result = await GetBoards(page, pageSize);
            set({
              boards: result,
              isLoading: false,
              hasCompletedOnboarding:
                get().hasCompletedOnboarding || result.length > 0,
            });
          } catch (error) {
            console.error("Failed to fetch boards:", error);
            set({
              error: "Failed to fetch boards",
              isLoading: false,
            });
          }
        },

        createBoard: async (name: string) => {
          set({ isLoading: true, error: null });
          try {
            const rawBoard = await CreateBoard(name);

            const createdAt = rawBoard.created_at;
            const updatedAt = rawBoard.updated_at;

            const payload = {
              id: rawBoard.id,
              name: rawBoard.name,
              created_at: createdAt,
              updated_at: updatedAt,
            };

            EventsEmit("board:create", JSON.stringify(payload));

            set((state) => ({
              boards: [rawBoard, ...state.boards],
              currentBoard: rawBoard,
              isLoading: false,
              hasCompletedOnboarding: true,
            }));

            return rawBoard;
          } catch (error) {
            console.error("Failed to create board:", error);
            set({
              error: "Failed to create board",
              isLoading: false,
            });
            return null;
          }
        },

        updateBoard: async (boardId: string, name: string) => {
          set({ isLoading: true, error: null });
          try {
            const rawBoard = await UpdateBoard(boardId, name);

            const roomId = useCollaborationStore.getState().roomId;
            const updatedAt = rawBoard.updated_at;

            const payload = {
              room_id: roomId,
              id: rawBoard.id,
              name: rawBoard.name,
              updated_at: updatedAt,
            };

            EventsEmit("board:data", JSON.stringify(payload));

            set((state) => ({
              boards: state.boards.map((b) =>
                b.id === boardId ? rawBoard : b
              ),
              currentBoard:
                state.currentBoard?.id === boardId
                  ? rawBoard
                  : state.currentBoard,
              isLoading: false,
            }));

            return rawBoard;
          } catch (error) {
            console.error("Failed to update board:", error);
            set({
              error: "Failed to update board",
              isLoading: false,
            });
            return null;
          }
        },

        deleteBoard: async (boardId: string) => {
          set({ isLoading: true, error: null });
          try {
            console.log("deleteing..");
            await DeleteBoard(boardId);

            set((state) => {
              const newBoards = state.boards.filter((b) => b.id !== boardId);
              const newCurrentBoard =
                state.currentBoard?.id === boardId
                  ? newBoards.length > 0
                    ? newBoards[0]
                    : null
                  : state.currentBoard;

              return {
                boards: newBoards,
                currentBoard: newCurrentBoard,
                isLoading: false,
                hasCompletedOnboarding: state.hasCompletedOnboarding,
              };
            });

            return true;
          } catch (error) {
            console.error("Failed to delete board:", error);
            set({
              error: "Failed to delete board",
              isLoading: false,
            });
            return false;
          }
        },

        selectBoard: async (boardId: string) => {
          set({ isLoading: true, error: null });
          try {
            const rawBoard = await GetBoardByID(boardId);
            set({
              currentBoard: rawBoard,
              isLoading: false,
            });

            // Reinitialize collaboration if authenticated
            const { isAuthenticated } = await import("./auth-store").then(
              (m) => m.useDesktopAuthStore.getState()
            );
            if (isAuthenticated) {
              const { reinitialize } = useCollaborationStore.getState();
              reinitialize(boardId);
            }
          } catch (error) {
            console.error("Failed to select board:", error);
            set({
              error: "Failed to select board",
              isLoading: false,
            });
          }
        },

        setCurrentBoard: (board) => {
          set({ currentBoard: board });
        },

        setHasCompletedOnboarding: (completed) => {
          set({ hasCompletedOnboarding: completed });
        },

        clearError: () => {
          set({ error: null });
        },
      }),
      {
        name: "board-storage",
        partialize: (state) => ({
          hasCompletedOnboarding: state.hasCompletedOnboarding,
          currentBoard: state.currentBoard,
        }),
      }
    ),
    { name: "board-store" }
  )
);
