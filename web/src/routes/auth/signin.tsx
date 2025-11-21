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
import { AlertCircle, ArrowRight } from "lucide-react";
import { DitherBackground } from "@/components/ui/dither-background";
import Logo from "@/assets/logo.png";

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
    <div className="min-h-screen bg-white dark:bg-black text-black dark:text-white font-sans flex items-center justify-center p-4 relative overflow-hidden">
      <DitherBackground opacity={0.15} />
      
      <div className="w-full max-w-md relative z-10">
        <div className="mb-8 text-center">
         <Link to="/" className="inline-flex items-center justify-center w-12 h-12 mb-6">
            <img src={Logo} alt="Seisami Logo" />
          </Link>
          <h1 className="text-3xl font-bold tracking-tighter uppercase mb-2">Welcome Back</h1>
          <p className="text-gray-500 dark:text-gray-400 font-mono text-sm">Enter your credentials to access your workspace</p>
        </div>

        <div className="bg-white dark:bg-black border border-black dark:border-white p-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] dark:shadow-[8px_8px_0px_0px_rgba(255,255,255,1)]">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            {error && (
              <div className="flex gap-3 px-4 py-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 text-sm font-mono">
                <AlertCircle className="h-5 w-5 shrink-0" />
                <p>{error}</p>
              </div>
            )}

            <div className="space-y-2">
              <label htmlFor="email" className="text-xs font-mono uppercase tracking-wider font-bold">
                Email
              </label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                {...register("email")}
                disabled={isLoading}
                className="rounded-none border-black dark:border-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-black focus-visible:border-2 dark:focus-visible:border-white h-12"
              />
              {errors.email && (
                <p className="text-xs text-red-500 font-mono mt-1">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label htmlFor="password" className="text-xs font-mono uppercase tracking-wider font-bold">
                  Password
                </label>
                <Link
                  to="/auth/forgot-password"
                  className="text-xs font-mono text-gray-500 hover:text-black dark:hover:text-white underline underline-offset-4"
                >
                  Forgot Password?
                </Link>
              </div>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                {...register("password")}
                disabled={isLoading}
                className="rounded-none border-black dark:border-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-black focus-visible:border-2 dark:focus-visible:border-white h-12"
              />
              {errors.password && (
                <p className="text-xs text-red-500 font-mono mt-1">
                  {errors.password.message}
                </p>
              )}
            </div>

            <Button 
              type="submit" 
              className="w-full h-12 rounded-none bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 text-sm font-mono uppercase tracking-widest" 
              disabled={isLoading}
            >
              {isLoading ? "Signing in..." : "Sign In"} <ArrowRight className="ml-2 w-4 h-4" />
            </Button>
          </form>

          <div className="mt-8 pt-6 border-t border-gray-100 dark:border-gray-800 text-center">
            <span className="text-sm text-gray-500 dark:text-gray-400">
              Don't have an account?{" "}
            </span>
            <Link
              to="/auth/signup"
              className="text-sm font-bold text-black dark:text-white hover:underline underline-offset-4"
            >
              Create Account
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
