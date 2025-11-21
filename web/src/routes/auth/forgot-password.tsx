import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { ApiClient } from "../../lib/api-client";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { AlertCircle, ArrowRight, KeyRound, CheckCircle2 } from "lucide-react";
import { DitherBackground } from "@/components/ui/dither-background";

export const Route = createFileRoute("/auth/forgot-password")({
  component: ForgotPasswordPage,
});

function ForgotPasswordPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(false);
  const [resetToken, setResetToken] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!email) {
      setError("Email is required");
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      setError("Please enter a valid email");
      return;
    }

    setLoading(true);
    try {
      const result = await ApiClient.forgotPassword(email);
      setResetToken(result.reset_token);
      setSuccess(true);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to process request"
      );
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div className="min-h-screen bg-white dark:bg-black text-black dark:text-white font-sans flex items-center justify-center p-4 relative overflow-hidden">
        <DitherBackground opacity={0.15} />
        
        <div className="w-full max-w-md relative z-10">
          <div className="mb-8 text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 mb-6 bg-black dark:bg-white text-white dark:text-black">
              <CheckCircle2 size={24} />
            </div>
            <h1 className="text-3xl font-bold tracking-tighter uppercase mb-2">Check Your Email</h1>
            <p className="text-gray-500 dark:text-gray-400 font-mono text-sm">We've sent you a reset token</p>
          </div>

          <div className="bg-white dark:bg-black border border-black dark:border-white p-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] dark:shadow-[8px_8px_0px_0px_rgba(255,255,255,1)]">
            <div className="space-y-6">
              <div className="bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-800 p-4">
                <p className="text-sm text-gray-600 dark:text-gray-300 font-mono">
                  A password reset token has been sent to <strong className="text-black dark:text-white">{email}</strong>
                </p>
              </div>

              <div className="bg-black/5 dark:bg-white/5 border border-black/10 dark:border-white/10 p-4">
                <p className="text-sm text-gray-600 dark:text-gray-300">
                  <strong className="uppercase text-xs tracking-wider block mb-2">Your reset token:</strong>
                  <code className="block p-3 bg-white dark:bg-black border border-black/20 dark:border-white/20 text-sm break-all font-mono select-all">
                    {resetToken}
                  </code>
                </p>
              </div>

              <Button
                onClick={() => navigate({ to: "/auth/reset-password" })}
                className="w-full h-12 rounded-none bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 text-sm font-mono uppercase tracking-widest"
              >
                Go to Reset Password <ArrowRight className="ml-2 w-4 h-4" />
              </Button>

              <div className="text-center">
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
                  Back to Sign In
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white dark:bg-black text-black dark:text-white font-sans flex items-center justify-center p-4 relative overflow-hidden">
      <DitherBackground opacity={0.15} />
      
      <div className="w-full max-w-md relative z-10">
        <div className="mb-8 text-center">
          <div className="inline-flex items-center justify-center w-12 h-12 mb-6 bg-black dark:bg-white text-white dark:text-black">
            <KeyRound size={24} />
          </div>
          <h1 className="text-3xl font-bold tracking-tighter uppercase mb-2">Forgot Password</h1>
          <p className="text-gray-500 dark:text-gray-400 font-mono text-sm">Enter your email to recover your account</p>
        </div>

        <div className="bg-white dark:bg-black border border-black dark:border-white p-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] dark:shadow-[8px_8px_0px_0px_rgba(255,255,255,1)]">
          <form onSubmit={handleSubmit} className="space-y-6">
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
                value={email}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setEmail(e.target.value)
                }
                disabled={loading}
                className="rounded-none border-black dark:border-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-black focus-visible:border-2 dark:focus-visible:border-white h-12"
              />
            </div>

            <Button 
              type="submit" 
              className="w-full h-12 rounded-none bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 text-sm font-mono uppercase tracking-widest" 
              disabled={loading}
            >
              {loading ? "Sending..." : "Send Reset Token"} <ArrowRight className="ml-2 w-4 h-4" />
            </Button>
          </form>

          <div className="mt-8 pt-6 border-t border-gray-100 dark:border-gray-800 text-center">
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
              Back to Sign In
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
