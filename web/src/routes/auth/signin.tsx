import {
  createFileRoute,
  Link,
  useNavigate,
  useSearch,
} from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ApiClient } from "../../lib/api-client";
import { useAuthStore } from "../../stores/auth-store";
import { SigninFormData, signinSchema } from "../../lib/validation";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../components/ui/card";
import { AlertCircle } from "lucide-react";

export const Route = createFileRoute("/auth/signin")({
  component: SigninPage,
  validateSearch(search) {
    return {
      type: (search?.type as string) || undefined,
      state: (search?.state as string) || undefined,
      desktop: (search?.desktop as boolean) || undefined,
      redirect: (search?.redirect as string) || undefined,
    };
  },
});

function SigninPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/auth/signin" });
  const redirectTo = search.redirect || "/";
  const state = search?.state;
  const isDesktop = search?.desktop == true;

  const { setAuth, setLoading, setError, error, isLoading } = useAuthStore();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SigninFormData>({
    resolver: zodResolver(signinSchema),
  });

  const onSubmit = async (data: SigninFormData) => {
    setError(null);
    setLoading(true);

    try {
      const result = await ApiClient.signin(data.email, data.password);
      setAuth(result.token, result.user_id, result.email);

      if (isDesktop && state) {
        await notifyDesktopApp(result.token, state);
      } else {
        navigate({ to: redirectTo });
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : "Signin failed";
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  const notifyDesktopApp = async (token: string, state: string) => {
    try {
      const API_BASE_URL =
        import.meta.env.VITE_API_URL || "http://localhost:8080";

      const response = await fetch(
        `${API_BASE_URL}/auth/desktop/start?state=${state}`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error("Failed to get desktop code");
      }

      // Parse JSON response to get code and expires
      const data = await response.json();
      const { code } = data;

      if (!code) {
        throw new Error("No code in response");
      }

      const deepLink = `seisami://auth/callback?token=${token}&code=${code}&state=${state}`;
      window.location.href = deepLink;
    } catch (err) {
      console.error("Error notifying desktop app:", err);
      setError("Failed to complete desktop authentication");
    }
  };

  return (
    <div className="min-h-screen bg-white dark:bg-black flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Sign In</CardTitle>
          <CardDescription>Welcome back to Seisami</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <div className="flex gap-3 px-4 py-3 bg-red-50 border border-red-200 rounded-md">
                <AlertCircle className="h-5 w-5 text-red-600 shrink-0 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-red-800">{error}</p>
                </div>
              </div>
            )}

            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium">
                Email
              </label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                {...register("email")}
                disabled={isLoading}
              />
              {errors.email && (
                <p className="text-xs text-red-500">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium">
                Password
              </label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                {...register("password")}
                disabled={isLoading}
              />
              {errors.password && (
                <p className="text-xs text-red-500">
                  {errors.password.message}
                </p>
              )}
            </div>

            <div className="text-right">
              <Link
                to="/auth/forgot-password"
                className="text-sm text-primary hover:underline"
              >
                Forgot Password?
              </Link>
            </div>

            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? "Signing in..." : "Sign In"}
            </Button>
          </form>

          <div className="mt-4 text-center text-sm">
            <span className="text-muted-foreground">
              Don't have an account?{" "}
            </span>
            <Link
              to="/auth/signup"
              className="text-primary hover:underline font-medium"
            >
              Sign Up
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
