import { useEffect, useMemo, useState } from "react";
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
} from "../../wailsjs/go/main/App";

type PermissionStatus = -2 | -1 | 0 | 1;

const INITIAL_STATUS: PermissionStatus = -2;

const statusMeta = (status: PermissionStatus) => {
  switch (status) {
    case 1:
      return {
        text: "Allowed",
        badgeClass: "bg-black text-white",
      };
    case 0:
      return {
        text: "Denied",
        badgeClass: "bg-neutral-200 text-neutral-900",
      };
    case -1:
      return {
        text: "Not Determined",
        badgeClass: "bg-neutral-200 text-neutral-900",
      };
    default:
      return {
        text: "Checking...",
        badgeClass: "bg-neutral-200 text-neutral-900",
      };
  }
};

type Step = "permissions" | "board";

export function OnboardingScreen() {
  const [boardName, setBoardName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const { createBoard, isLoading, setHasCompletedOnboarding } = useBoardStore();
  const [micStatus, setMicStatus] = useState<PermissionStatus>(INITIAL_STATUS);
  const [accessibilityStatus, setAccessibilityStatus] =
    useState<PermissionStatus>(INITIAL_STATUS);
  const [requestingMic, setRequestingMic] = useState(false);
  const [requestingAccessibility, setRequestingAccessibility] = useState(false);
  const [step, setStep] = useState<Step>("permissions");
  const { isAuthenticated, setLoading, setError } = useDesktopAuthStore();
  const [isAuthenticating, setIsAuthenticating] = useState(false);

  const allPermissionsGranted = useMemo(
    () => micStatus === 1 && accessibilityStatus === 1,
    [micStatus, accessibilityStatus]
  );

  useEffect(() => {
    const fetchStatuses = async () => {
      try {
        const [mic, accessibility] = await Promise.all([
          CheckMicrophonePermission(),
          CheckAccessibilityPermission(),
        ]);
        setMicStatus(mic as PermissionStatus);
        setAccessibilityStatus(accessibility as PermissionStatus);
      } catch (error) {
        console.error("Failed to fetch permission statuses", error);
      }
    };

    fetchStatuses();
  }, []);

  const handleRequestMicrophone = async () => {
    setRequestingMic(true);
    try {
      await RequestMicrophonePermission();
    } catch (error) {
      console.error("Failed to request microphone permission", error);
    } finally {
      try {
        const status = await CheckMicrophonePermission();
        setMicStatus(status as PermissionStatus);
      } catch (error) {
        console.error("Failed to re-check microphone status", error);
      }
      setRequestingMic(false);
    }
  };

  const handleRequestAccessibility = async () => {
    setRequestingAccessibility(true);
    try {
      await RequestAccessibilityPermission();
    } catch (error) {
      console.error("Failed to request accessibility permission", error);
    } finally {
      try {
        const status = await CheckAccessibilityPermission();
        setAccessibilityStatus(status as PermissionStatus);
      } catch (error) {
        console.error("Failed to re-check accessibility status", error);
      }
      setRequestingAccessibility(false);
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
        <div className="border border-neutral-200 rounded-sm p-4">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="text-sm font-semibold uppercase tracking-wider">
                Microphone Access
              </h3>
              <p className="text-sm text-neutral-500 mt-1">
                Required to capture your voice commands during recording.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <span
                className={`px-2.5 py-0.5 text-xs font-medium rounded-full ${
                  statusMeta(micStatus).badgeClass
                }`}
              >
                {statusMeta(micStatus).text}
              </span>
              {micStatus !== 1 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRequestMicrophone}
                  disabled={requestingMic}
                >
                  {requestingMic ? "Requesting..." : "Allow"}
                </Button>
              )}
            </div>
          </div>
          {micStatus === 0 && (
            <div className="mt-3 text-xs text-neutral-600">
              Permission denied. Update it in System Settings.
              <Button
                variant="link"
                size="sm"
                className="px-1"
                onClick={() => OpenMicrophoneSettings()}
              >
                Open settings
              </Button>
            </div>
          )}
        </div>

        <div className="border border-neutral-200 rounded-sm p-4">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="text-sm font-semibold uppercase tracking-wider">
                Accessibility Access
              </h3>
              <p className="text-sm text-neutral-500 mt-1">
                Needed so Seisami can listen for the Fn hotkey while you work.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <span
                className={`px-2.5 py-0.5 text-xs font-medium rounded-full ${
                  statusMeta(accessibilityStatus).badgeClass
                }`}
              >
                {statusMeta(accessibilityStatus).text}
              </span>
              {accessibilityStatus !== 1 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRequestAccessibility}
                  disabled={requestingAccessibility}
                >
                  {requestingAccessibility ? "Requesting..." : "Allow"}
                </Button>
              )}
            </div>
          </div>
          {accessibilityStatus === 0 && (
            <div className="mt-3 text-xs text-neutral-600">
              Permission denied. Allow Seisami under Accessibility settings.
              <Button
                variant="link"
                size="sm"
                className="px-1"
                onClick={() => OpenAccessibilitySettings()}
              >
                Open settings
              </Button>
            </div>
          )}
        </div>
      </div>

      <div className="space-y-2">
        <Button
          className="w-full"
          onClick={() => setStep("board")}
          disabled={!allPermissionsGranted}
        >
          Continue
        </Button>
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
                ? "Confirm Device Access"
                : "Create Your First Board"}
            </CardTitle>
            <CardDescription>
              {step === "permissions"
                ? "Allow Seisami to use your microphone and listen for the Fn key."
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
