import { EventsOn, BrowserOpenURL } from "../../wailsjs/runtime";
import { useDesktopAuthStore } from "~/stores/auth-store";
import { SetLoginToken } from "../../wailsjs/go/main/App";

const WEB_BASE_URL = import.meta.env.VITE_WEB_URL || "http://localhost:3000";
const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export const DesktopAuthService = {
  startAuthFlow: async (): Promise<void> => {
    try {
      const state = generateRandomState();
      const loginUrl = `${WEB_BASE_URL}/auth/signin?state=${state}&desktop=true`;

      EventsOn(
        "auth:desktop_callback",
        async (data: Record<string, string>) => {
          console.log(data);
          const token = data.token;
          const code = data.code;
          const callbackState = data.state;

          if (callbackState !== state) {
            console.error("State mismatch - potential security issue");
            return;
          }

          try {
            const finalToken = await exchangeCodeForToken(
              token,
              code,
              callbackState
            );

            const authStore = useDesktopAuthStore.getState();
            authStore.setToken(finalToken, "local-user", "local@seisami.app");

            await SetLoginToken(finalToken);
          } catch (err) {
            console.error("Failed to exchange code:", err);
          }
        }
      );

      BrowserOpenURL(loginUrl);
    } catch (error) {
      console.error("Auth flow error:", error);
    }
  },
};

async function exchangeCodeForToken(
  token: string,
  code: string,
  state: string
): Promise<string> {
  const response = await fetch(`${API_BASE_URL}/auth/desktop/exchange`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ code, state }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || "Failed to exchange code for token");
  }

  const data = await response.json();
  return data.token;
}

function generateRandomState(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return Array.from(array, (byte) => byte.toString(16).padStart(2, "0")).join(
    ""
  );
}
