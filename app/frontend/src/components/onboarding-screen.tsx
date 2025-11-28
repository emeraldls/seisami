import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { useBoardStore } from "~/stores/board-store";
import { Label } from "~/components/ui/label";
import { useDesktopAuthStore } from "~/stores/auth-store";
import { DesktopAuthService } from "~/lib/desktop-auth-service";
import {
  CheckMicrophonePermission,
  RequestMicrophonePermission,
  OpenMicrophoneSettings,
  CheckAccessibilityPermission,
  RequestAccessibilityPermission,
  OpenAccessibilitySettings,
  RestartApp,
} from "../../wailsjs/go/main/App";

type Step = "permissions" | "board";

export function OnboardingScreen() {
  const [boardName, setBoardName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const { createBoard, isLoading, setHasCompletedOnboarding } = useBoardStore();
  const [micGranted, setMicGranted] = useState(false);
  const [accessibilityGranted, setAccessibilityGranted] = useState(false);
  const [requestingPermissions, setRequestingPermissions] = useState(false);
  const [step, setStep] = useState<Step>("permissions");
  const { isAuthenticated, setLoading, setError } = useDesktopAuthStore();
  const [isAuthenticating, setIsAuthenticating] = useState(false);

  const allPermissionsGranted = micGranted && accessibilityGranted;

  useEffect(() => {
    const checkStatuses = async () => {
      try {
        const [mic, accessibility] = await Promise.all([
          CheckMicrophonePermission(),
          CheckAccessibilityPermission(),
        ]);
        setMicGranted(mic === 1);
        setAccessibilityGranted(accessibility === 1);
      } catch (error) {
        console.error("Failed to check permissions", error);
      }
    };

    checkStatuses();

    // Poll for accessibility permission changes every 2 seconds
    // This handles the case where user grants permission in System Settings
    const intervalId = setInterval(async () => {
      try {
        const [mic, accessibility] = await Promise.all([
          CheckMicrophonePermission(),
          CheckAccessibilityPermission(),
        ]);
        setMicGranted(mic === 1);
        if (accessibility === 1 && !accessibilityGranted) {
          setAccessibilityGranted(true);
          toast.info("Accessibility granted. Restarting app...");
          // await RestartApp();
        } else {
          setAccessibilityGranted(accessibility === 1);
        }
      } catch (error) {
        console.error("Failed to check permissions", error);
      }
    }, 2000);

    return () => clearInterval(intervalId);
  }, []);

  const handleRequestPermissions = async () => {
    setRequestingPermissions(true);
    try {
      // Request microphone first
      if (!micGranted) {
        await RequestMicrophonePermission();
        await new Promise((resolve) => setTimeout(resolve, 500)); // Small delay
        const micStatus = await CheckMicrophonePermission();
        setMicGranted(micStatus === 1);
      }

      if (!accessibilityGranted) {
        await RequestAccessibilityPermission();

        // Give user 3 seconds to see the dialog before we start checking
        await new Promise((resolve) => setTimeout(resolve, 3000));

        // Check immediately after the dialog
        const immediateCheck = await CheckAccessibilityPermission();
        if (immediateCheck === 1) {
          setAccessibilityGranted(true);
          setRequestingPermissions(false);
          toast.info("Accessibility granted. Restarting app...");
          // await RestartApp();
          return;
        }
      }
    } catch (error) {
      console.error("Failed to request permissions", error);
    } finally {
      setRequestingPermissions(false);
    }
  };

  const handleRefreshPermissions = async () => {
    try {
      const [mic, accessibility] = await Promise.all([
        CheckMicrophonePermission(),
        CheckAccessibilityPermission(),
      ]);
      setMicGranted(mic === 1);
      setAccessibilityGranted(accessibility === 1);
    } catch (error) {
      console.error("Failed to refresh permissions", error);
    }
  };

  const handleCreateBoard = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!boardName.trim()) return;

    setIsCreating(true);
    const board = await createBoard(boardName.trim());
    if (board) {
      setBoardName("");
    }
    setIsCreating(false);
  };

  const handleEnterApp = () => {
    setHasCompletedOnboarding(true);
  };

  const handleCloudLogin = async () => {
    setIsAuthenticating(true);
    setLoading(true);
    try {
      await DesktopAuthService.startAuthFlow();
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Authentication failed";
      setError(message);
      console.error("Cloud login error:", error);
    } finally {
      setIsAuthenticating(false);
      setLoading(false);
    }
  };

  const renderPermissionsStep = () => (
    <div className="space-y-6">
      <div className="space-y-4">
        <p className="text-sm text-neutral-600 text-center">
          Seisami needs access to your microphone and keyboard to work properly.
        </p>

        <div className="space-y-3">
          <div className="border border-neutral-200 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center ${
                    micGranted ? "bg-green-100" : "bg-neutral-100"
                  }`}
                >
                  {micGranted ? (
                    <svg
                      className="w-5 h-5 text-green-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                  ) : (
                    <svg
                      className="w-5 h-5 text-neutral-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"
                      />
                    </svg>
                  )}
                </div>
                <div>
                  <h4 className="font-medium text-sm">Microphone Access</h4>
                  <p className="text-xs text-neutral-500">
                    {micGranted ? "Granted" : "Required for voice recording"}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div className="border border-neutral-200 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center ${
                    accessibilityGranted ? "bg-green-100" : "bg-neutral-100"
                  }`}
                >
                  {accessibilityGranted ? (
                    <svg
                      className="w-5 h-5 text-green-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                  ) : (
                    <svg
                      className="w-5 h-5 text-neutral-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                      />
                    </svg>
                  )}
                </div>
                <div>
                  <h4 className="font-medium text-sm">Accessibility Access</h4>
                  <p className="text-xs text-neutral-500">
                    {accessibilityGranted
                      ? "Granted"
                      : "Required for Fn key detection"}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        {!allPermissionsGranted && (
          <Button
            className="w-full"
            onClick={handleRequestPermissions}
            disabled={requestingPermissions}
            size="lg"
          >
            {requestingPermissions
              ? "Requesting Access..."
              : "Grant Permissions"}
          </Button>
        )}

        {allPermissionsGranted && (
          <Button className="w-full" onClick={() => setStep("board")} size="lg">
            Continue
          </Button>
        )}
      </div>
    </div>
  );

  const renderBoardStep = () => (
    <div className="space-y-5">
      <form onSubmit={handleCreateBoard} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="boardName">Board Name</Label>
          <Input
            id="boardName"
            type="text"
            placeholder="e.g., Personal Tasks, Work Projects..."
            value={boardName}
            onChange={(e) => setBoardName(e.target.value)}
            className="w-full"
            disabled={isCreating || isLoading}
            autoFocus
          />
        </div>

        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between gap-2">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setStep("permissions")}
            >
              Review permissions
            </Button>
            <Button
              type="submit"
              disabled={!boardName.trim() || isCreating || isLoading}
            >
              {isCreating || isLoading ? "Creating Board..." : "Create Board"}
            </Button>
          </div>
        </div>
      </form>
      <div className="border-t pt-4 text-center space-y-3">
        <div className="text-sm text-neutral-500">
          Already have a Seisami Cloud account?
        </div>
        <Button
          type="button"
          variant="secondary"
          className="w-full"
          onClick={handleCloudLogin}
          disabled={isAuthenticating}
        >
          {isAuthenticating ? "Opening sign-in..." : "Sign in to Cloud"}
        </Button>
        {isAuthenticated && (
          <div className="text-xs text-neutral-500">
            You're signed in. Any boards you have will appear once synced.
          </div>
        )}
        <div className="border-t pt-3">
          <Button
            type="button"
            variant="ghost"
            className="w-full text-neutral-500"
            onClick={handleEnterApp}
          >
            Skip for now
          </Button>
        </div>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-lg">
        <Card className="shadow-xl border border-neutral-200">
          <CardHeader className="space-y-1 text-center">
            <CardTitle className="text-2xl">
              {step === "permissions"
                ? "Welcome to Seisami"
                : "Create Your First Board"}
            </CardTitle>
            <CardDescription>
              {step === "permissions"
                ? "Let's get you set up in just a few clicks."
                : "Start organizing your ideas with a personalized board."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {step === "permissions"
              ? renderPermissionsStep()
              : renderBoardStep()}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
