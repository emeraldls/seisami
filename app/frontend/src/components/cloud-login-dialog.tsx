import { FormEvent, useEffect, useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "./ui/dialog";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Button } from "./ui/button";
import { useDesktopAuthStore } from "~/stores/auth-store";
import { apiClient } from "~/lib/api-client";
import { toast } from "sonner";
import { isAxiosError } from "axios";
import { DesktopAuthService } from "~/lib/desktop-auth-service";

interface CloudLoginDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAuthStart?: () => void;
  onAuthEnd?: () => void;
}

interface AuthResponse {
  token: string;
  user_id: string;
  email: string;
}

export const CloudLoginDialog = ({
  open,
  onOpenChange,
  onAuthStart,
  onAuthEnd,
}: CloudLoginDialogProps) => {
  const {
    setToken,
    setError,
    clearError,
    setLoading,
    error: storeError,
  } = useDesktopAuthStore();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [localError, setLocalError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!open) {
      setEmail("");
      setPassword("");
      setLocalError(null);
      clearError();
      setIsSubmitting(false);
    }
  }, [open, clearError]);

  const resolvedError = localError ?? storeError;

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!email.trim() || !password) {
      setLocalError("Email and password are required");
      setError("Email and password are required");
      toast.error("Email and password are required");
      return;
    }

    setLocalError(null);
    clearError();
    setIsSubmitting(true);
    setLoading(true);
    onAuthStart?.();

    try {
      const result = await apiClient.post<AuthResponse>("/auth/signin", {
        email: email.trim(),
        password,
      });

      setToken(result.token, result.user_id, result.email);
      toast.success("Signed in to Seisami Cloud");
      onOpenChange(false);
    } catch (error) {
      let message = "Unable to sign in";
      if (isAxiosError(error)) {
        const data = error.response?.data as { error?: string } | undefined;
        if (data?.error) {
          message = data.error;
        }
      } else if (error instanceof Error) {
        message = error.message;
      }

      setLocalError(message);
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
      setLoading(false);
      onAuthEnd?.();
    }
  };

  const handleDesktopFlow = async () => {
    setLocalError(null);
    clearError();
    setIsSubmitting(true);
    setLoading(true);
    onAuthStart?.();

    try {
      await DesktopAuthService.startAuthFlow();
      toast.info("Continue login in your browser");
      onOpenChange(false);
    } catch (error) {
      let message = "Failed to start desktop login flow";
      if (error instanceof Error) {
        message = error.message;
      }

      setLocalError(message);
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
      setLoading(false);
      onAuthEnd?.();
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Sign in to Seisami Cloud</DialogTitle>
          <DialogDescription>
            Use your Seisami credentials to enable cloud features without leaving the app.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="cloud-login-email">Email</Label>
            <Input
              id="cloud-login-email"
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              disabled={isSubmitting}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="cloud-login-password">Password</Label>
            <Input
              id="cloud-login-password"
              type="password"
              autoComplete="current-password"
              placeholder="••••••••"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              disabled={isSubmitting}
              required
            />
          </div>

          {resolvedError && (
            <p className="text-sm text-red-600" role="alert">
              {resolvedError}
            </p>
          )}

          <div className="flex flex-col gap-2">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Signing in..." : "Sign in"}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={handleDesktopFlow}
              disabled={isSubmitting}
            >
              Continue with browser login
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};
