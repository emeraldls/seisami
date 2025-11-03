import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import App from "./App";
import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useDesktopAuthStore } from "./stores/auth-store";
import { SetLoginToken } from "../wailsjs/go/main/App";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5 * 60 * 1000,
    },
  },
});

const initializeAuth = async () => {
  const { token } = useDesktopAuthStore.getState();
  if (token) {
    try {
      await SetLoginToken(token);
      console.log("Login token initialized from storage");
    } catch (error) {
      console.error("Failed to set login token:", error);
    }
  }
};

initializeAuth();

const container = document.getElementById("root");

const root = createRoot(container!);

root.render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </QueryClientProvider>
  </React.StrictMode>
);
