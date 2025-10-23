const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export interface AuthResult {
  token: string;
  user_id: string;
  email: string;
}

export interface ForgotPasswordResponse {
  message: string;
  reset_token: string;
  expires_at: string;
  expires_in_seconds: number;
}

export class ApiClient {
  static async signup(email: string, password: string): Promise<AuthResult> {
    const response = await fetch(`${API_BASE_URL}/auth/signup`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Signup failed");
    }

    return response.json();
  }

  static async signin(email: string, password: string): Promise<AuthResult> {
    const response = await fetch(`${API_BASE_URL}/auth/signin`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Signin failed");
    }

    return response.json();
  }

  static async forgotPassword(email: string): Promise<ForgotPasswordResponse> {
    const response = await fetch(`${API_BASE_URL}/auth/forgot-password`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to process forgot password");
    }

    return response.json();
  }

  static async resetPassword(token: string, password: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/auth/reset-password`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ token, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to reset password");
    }
  }
}
