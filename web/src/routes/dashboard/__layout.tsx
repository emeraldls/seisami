import { createFileRoute, Outlet, useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useAuthStore } from "../../stores/auth-store";

export const Route = createFileRoute("/dashboard/__layout")({
  component: DashboardLayout,
});

function DashboardLayout() {
  const navigate = useNavigate();
  const { token } = useAuthStore();

  useEffect(() => {
    if (!token) {
      navigate({
        to: "/auth/signin",
        search: {
          type: undefined,
          state: undefined,
          desktop: undefined,
          redirect: window.location.pathname,
        },
      });
    }
  }, [token, navigate]);

  if (!token) {
    return null;
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold">Seisami</h1>
            </div>
            <nav className="flex items-center gap-4">
              <span className="text-sm text-muted-foreground">
                {useAuthStore.getState().email}
              </span>
              <button
                onClick={() => {
                  useAuthStore.getState().logout();
                  navigate({
                    to: "/auth/signin",
                    search: {
                      desktop: undefined,
                      redirect: undefined,
                      state: undefined,
                      type: undefined,
                    },
                  });
                }}
                className="text-sm text-muted-foreground hover:text-foreground"
              >
                Logout
              </button>
            </nav>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}
