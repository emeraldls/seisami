import { useEffect } from "react";
import { Route, Routes } from "react-router-dom";
import BoardManagement from "~/views/board-management";
import { AppLayout } from "~/layouts/app-layout";
import { SidebarProvider } from "~/contexts/sidebar-context";
import { CommandPaletteProvider } from "~/contexts/command-palette-context";
import { OnboardingScreen } from "~/components/onboarding-screen";
import { ErrorBoundary } from "~/components/error-boundary";
import { LoadingScreen } from "~/components/loading";
import { useBoardStore } from "~/stores/board-store";
import Transcriptions from "~/views/transcriptions";
import KanbanView from "~/views/board";
import Settings from "./views/settings";
import { Toaster } from "~/components/ui/sonner";

export default function App() {
  const { hasCompletedOnboarding, fetchBoards, isLoading } = useBoardStore();

  useEffect(() => {
    fetchBoards();
  }, [fetchBoards]);

  if (isLoading && !hasCompletedOnboarding) {
    return <LoadingScreen message="Initializing Seisami..." />;
  }

  if (!hasCompletedOnboarding) {
    return (
      <ErrorBoundary>
        <OnboardingScreen />
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
      <CommandPaletteProvider>
        <SidebarProvider>
          <Routes>
            <Route element={<AppLayout />}>
              <Route element={<KanbanView />} path="/" />
              <Route element={<Transcriptions />} path="/transcriptions" />
              <Route element={<BoardManagement />} path="/boards" />
              <Route element={<Settings />} path="/settings" />
            </Route>
          </Routes>
          <Toaster />
        </SidebarProvider>
      </CommandPaletteProvider>
    </ErrorBoundary>
  );
}
