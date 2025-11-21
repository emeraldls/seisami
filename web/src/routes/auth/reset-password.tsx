import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { ApiClient } from "../../lib/api-client";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { AlertCircle, ArrowRight, Lock, CheckCircle2 } from "lucide-react";
import { DitherBackground } from "@/components/ui/dither-background";

export const Route = createFileRoute("/auth/reset-password")({
  component: ResetPasswordPage,
});

function ResetPasswordPage() {
  const navigate = useNavigate();
  const [token, setToken] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(false);

  const validateForm = (): boolean => {
    if (!token || !password || !confirmPassword) {
      setError("All fields are required");
      return false;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return false;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return false;
    }

    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!validateForm()) {
      return;
    }

    setLoading(true);
    try {
      await ApiClient.resetPassword(token, password);
      setSuccess(true);
      setTimeout(() => {
        navigate({ 
          to: "/auth/signin",
          search: {
            desktop: undefined,
            redirect: undefined,
            state: undefined,
            type: undefined,
          }
        });
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to reset password");
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
            <h1 className="text-3xl font-bold tracking-tighter uppercase mb-2">Success!</h1>
            <p className="text-gray-500 dark:text-gray-400 font-mono text-sm">Your password has been reset</p>
          </div>

          <div className="bg-white dark:bg-black border border-black dark:border-white p-8 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] dark:shadow-[8px_8px_0px_0px_rgba(255,255,255,1)]">
            <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 p-4 mb-6">
              <p className="text-sm text-green-800 dark:text-green-400 font-mono">
                Your password has been successfully updated. Redirecting to sign in...
              </p>
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
            <Lock size={24} />
          </div>
          <h1 className="text-3xl font-bold tracking-tighter uppercase mb-2">Reset Password</h1>
          <p className="text-gray-500 dark:text-gray-400 font-mono text-sm">Enter your reset token and new password</p>
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
              <label htmlFor="token" className="text-xs font-mono uppercase tracking-wider font-bold">
                Reset Token
              </label>
              <Input
                id="token"
                type="text"
                placeholder="Paste your reset token here"
                value={token}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setToken(e.target.value)
                }
                disabled={loading}
                className="rounded-none border-black dark:border-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-black focus-visible:border-2 dark:focus-visible:border-white h-12"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="password" className="text-xs font-mono uppercase tracking-wider font-bold">
                New Password
              </label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setPassword(e.target.value)
                }
                disabled={loading}
                className="rounded-none border-black dark:border-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-black focus-visible:border-2 dark:focus-visible:border-white h-12"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-xs font-mono uppercase tracking-wider font-bold">
                Confirm Password
              </label>
              <Input
                id="confirmPassword"
                type="password"
                placeholder="••••••••"
                value={confirmPassword}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  setConfirmPassword(e.target.value)
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
              {loading ? "Resetting..." : "Reset Password"} <ArrowRight className="ml-2 w-4 h-4" />
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
