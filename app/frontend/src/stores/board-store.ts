import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";
import {
  CreateBoard,
  GetBoards,
  GetBoardByID,
  UpdateBoard,
  DeleteBoard,
} from "../../wailsjs/go/main/App";

export interface Board {
  ID: string;
  Name: string;
  CreatedAt: {
    String: string;
    Valid: boolean;
  };
  UpdatedAt: {
    String: string;
    Valid: boolean;
  };
}

// Helper interface for normalized board data used in components
export interface NormalizedBoard {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

// Helper function to normalize board data
export const normalizeBoard = (board: Board): NormalizedBoard => ({
  id: board.ID,
  name: board.Name,
  created_at: board.CreatedAt.Valid ? board.CreatedAt.String : "",
  updated_at: board.UpdatedAt.Valid ? board.UpdatedAt.String : "",
});

// Helper function to normalize board array
export const normalizeBoards = (boards: Board[]): NormalizedBoard[] =>
  boards.map(normalizeBoard);

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
            const rawBoards = Array.isArray(result) ? (result as Board[]) : [];
            const boards = normalizeBoards(rawBoards);
            set({
              boards,
              isLoading: false,
              hasCompletedOnboarding: boards.length > 0,
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
            const board = normalizeBoard(rawBoard as Board);

            set((state) => ({
              boards: [board, ...state.boards],
              currentBoard: board,
              isLoading: false,
              hasCompletedOnboarding: true,
            }));

            return board;
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
            const board = normalizeBoard(rawBoard as Board);

            set((state) => ({
              boards: state.boards.map((b) => (b.id === boardId ? board : b)),
              currentBoard:
                state.currentBoard?.id === boardId ? board : state.currentBoard,
              isLoading: false,
            }));

            return board;
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
                hasCompletedOnboarding: newBoards.length > 0,
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
            const board = normalizeBoard(rawBoard as Board);
            set({
              currentBoard: board,
              isLoading: false,
            });
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
