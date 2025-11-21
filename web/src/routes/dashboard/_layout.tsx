import { useAuthStore } from "@/stores/auth-store";
import { createFileRoute, Outlet, useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import logo from "@/assets/logo.png";
import { LogOut, User } from "lucide-react";

export const Route = createFileRoute("/dashboard/_layout")({
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
    <div className="min-h-screen bg-white dark:bg-black text-black dark:text-white font-sans">
      <header className="border-b border-black dark:border-white sticky top-0 bg-white/80 dark:bg-black/80 backdrop-blur-md z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img src={logo} alt="Seisami" className="h-6 w-6" />
            <h1 className="text-lg font-bold font-mono uppercase tracking-wider">Seisami <span className="text-gray-400 dark:text-gray-600">/ Dashboard</span></h1>
          </div>
          
          <nav className="flex items-center gap-6">
            <div className="flex items-center gap-2 text-sm font-mono text-gray-500 dark:text-gray-400">
              <User size={14} />
              <span>{useAuthStore.getState().email}</span>
            </div>
            
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
              className="flex items-center gap-2 text-sm font-mono uppercase tracking-wider text-black dark:text-white hover:opacity-70 transition-opacity"
            >
              <LogOut size={14} />
              Logout
            </button>
          </nav>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <Outlet />
      </main>
    </div>
  );
}
