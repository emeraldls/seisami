import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ApiClient } from "../../lib/api-client";
import { useAuthStore } from "../../stores/auth-store";
import { SignupFormData, signupSchema } from "../../lib/validation";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { AlertCircle, ArrowRight } from "lucide-react";
import { DitherBackground } from "@/components/ui/dither-background";
import Logo from "@/assets/logo.png";

export const Route = createFileRoute("/auth/signup")({
  component: SignupPage,
});

function SignupPage() {
  const navigate = useNavigate();
  const { setAuth, setLoading, setError, error, isLoading } = useAuthStore();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SignupFormData>({
    resolver: zodResolver(signupSchema),
  });

  const onSubmit = async (data: SignupFormData) => {
    setError(null);
    setLoading(true);

    try {
      const result = await ApiClient.signup(data.email, data.password);
      setAuth(result.token, result.user_id, result.email);
      navigate({ to: "/" });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Signup failed";
      setError(message);
    } finally {
      setLoading(false);
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
          <h1 className="text-3xl font-bold tracking-tighter uppercase mb-2">Join Seisami</h1>
          <p className="text-gray-500 dark:text-gray-400 font-mono text-sm">Start your voice-first productivity journey</p>
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
              <label htmlFor="password" className="text-xs font-mono uppercase tracking-wider font-bold">
                Password
              </label>
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

            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-xs font-mono uppercase tracking-wider font-bold">
                Confirm Password
              </label>
              <Input
                id="confirmPassword"
                type="password"
                placeholder="••••••••"
                {...register("confirmPassword")}
                disabled={isLoading}
                className="rounded-none border-black dark:border-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-black focus-visible:border-2 dark:focus-visible:border-white h-12"
              />
              {errors.confirmPassword && (
                <p className="text-xs text-red-500 font-mono mt-1">
                  {errors.confirmPassword.message}
                </p>
              )}
            </div>

            <Button 
              type="submit" 
              className="w-full h-12 rounded-none bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 text-sm font-mono uppercase tracking-widest" 
              disabled={isLoading}
            >
              {isLoading ? "Creating account..." : "Sign Up"} <ArrowRight className="ml-2 w-4 h-4" />
            </Button>
          </form>

          <div className="mt-8 pt-6 border-t border-gray-100 dark:border-gray-800 text-center">
            <span className="text-sm text-gray-500 dark:text-gray-400">
              Already have an account?{" "}
            </span>
            <Link
              to="/auth/signin"
              search={{
                desktop: undefined,
                redirect: undefined,
                state: undefined,
                type: undefined,
              }}
              className="text-sm font-bold text-black dark:text-white hover:underline underline-offset-4"
            >
              Sign In
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
