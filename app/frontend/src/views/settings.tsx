import React, { useEffect, useState } from "react";
import { Card } from "~/components/ui/card";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
  GetSettings,
  SaveSettings,
  OpenFileDialog,
  CheckMicrophonePermission,
  RequestMicrophonePermission,
  OpenMicrophoneSettings,
  CheckAccessibilityPermission,
  RequestAccessibilityPermission,
  OpenAccessibilitySettings,
  RestartApp,
} from "../../wailsjs/go/main/App";
import { frontend } from "../../wailsjs/go/models";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

const Settings = () => {
  const queryClient = useQueryClient();
  const [transcriptionMethod, setTranscriptionMethod] =
    useState<string>("cloud");
  const [whisperBinaryPath, setWhisperBinaryPath] = useState<string>("");
  const [whisperModelPath, setWhisperModelPath] = useState<string>("");
  const [openaiApiKey, setOpenaiApiKey] = useState<string>("");
  const [micGranted, setMicGranted] = useState(false);
  const [accessibilityGranted, setAccessibilityGranted] = useState(false);

  const { isLoading: loading } = useQuery({
    queryKey: ["settings"],
    queryFn: async () => {
      const currentSettings = await GetSettings();
      setTranscriptionMethod(currentSettings.TranscriptionMethod || "cloud");
      setWhisperBinaryPath(currentSettings.WhisperBinaryPath?.String || "");
      setWhisperModelPath(currentSettings.WhisperModelPath?.String || "");
      setOpenaiApiKey(currentSettings.OpenaiApiKey?.String || "");
      return currentSettings;
    },
  });

  useEffect(() => {
    let isMounted = true;

    const fetchPermissions = async () => {
      try {
        const [mic, accessibility] = await Promise.all([
          CheckMicrophonePermission(),
          CheckAccessibilityPermission(),
        ]);

        if (!isMounted) {
          return;
        }

        setMicGranted(mic === 1);
        setAccessibilityGranted(accessibility === 1);
      } catch (error) {
        console.error("Failed to load permission statuses", error);
      }
    };

    fetchPermissions();

    const intervalId = setInterval(fetchPermissions, 2000);

    return () => {
      isMounted = false;
      clearInterval(intervalId);
    };
  }, []);

  // Save settings mutation
  const saveSettingsMutation = useMutation({
    mutationFn: async () => {
      await SaveSettings(
        transcriptionMethod,
        whisperBinaryPath || null,
        whisperModelPath || null,
        openaiApiKey || null
      );
    },
    onSuccess: () => {
      toast.success("Settings saved successfully!");
      queryClient.invalidateQueries({ queryKey: ["settings"] });
    },
    onError: (error: Error) => {
      toast.error(
        error.message || "Failed to save settings. Please try again."
      );
    },
  });

  // Request microphone permission mutation
  const requestMicPermissionMutation = useMutation({
    mutationFn: RequestMicrophonePermission,
    onSuccess: async (granted) => {
      try {
        const status = await CheckMicrophonePermission();
        const isGranted = status === 1 || granted;
        setMicGranted(isGranted);

        if (!isGranted) {
          const openSettings = confirm(
            "Microphone permission was denied. Would you like to open System Settings to grant permission manually?"
          );
          if (openSettings) {
            OpenMicrophoneSettings();
          }
        }
      } catch (error) {
        console.error("Failed to refresh microphone permission", error);
      }
    },
    onError: (error) => {
      console.error("Failed to request microphone permission:", error);
      toast.error("Failed to request microphone permission");
    },
  });

  const handleRequestMicrophonePermission = async () => {
    requestMicPermissionMutation.mutate();
  };

  // Request accessibility permission mutation
  const requestAccessibilityPermissionMutation = useMutation({
    mutationFn: RequestAccessibilityPermission,
    onSuccess: () => {
      const pollAccessibility = async () => {
        try {
          const status = await CheckAccessibilityPermission();
          if (status === 1) {
            setAccessibilityGranted(true);
            toast.info("Accessibility granted. Restarting app...");
            await RestartApp();
            return true;
          }
          setAccessibilityGranted(false);
        } catch (error) {
          console.error("Failed to refresh accessibility permission", error);
        }
        return false;
      };

      pollAccessibility();

      const intervalId = setInterval(async () => {
        const granted = await pollAccessibility();
        if (granted) {
          clearInterval(intervalId);
        }
      }, 2000);

      setTimeout(() => clearInterval(intervalId), 60000);
    },
    onError: (error) => {
      console.error("Failed to request accessibility permission:", error);
      toast.error("Failed to request accessibility permission");
    },
  });

  const handleRequestAccessibilityPermission = async () => {
    requestAccessibilityPermissionMutation.mutate();
  };

  const handleSaveSettings = async () => {
    saveSettingsMutation.mutate();
  };

  const handleSelectWhisperBinary = async () => {
    try {
      const filters: frontend.FileFilter[] = [];
      const path = await OpenFileDialog("Select Whisper Binary", filters);
      if (path) {
        setWhisperBinaryPath(path);
      }
    } catch (error) {
      console.error("Failed to open file dialog:", error);
    }
  };

  const handleSelectWhisperModel = async () => {
    try {
      const filters: frontend.FileFilter[] = [
        { DisplayName: "Model Files", Pattern: "*.bin" },
        { DisplayName: "All Files", Pattern: "*" },
      ];
      const path = await OpenFileDialog("Select Whisper Model", filters);
      if (path) {
        setWhisperModelPath(path);
      }
    } catch (error) {
      console.error("Failed to open file dialog:", error);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-lg">Loading settings...</div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-2xl font-semibold text-neutral-900">Settings</h1>
        <p className="text-sm text-neutral-500 mt-1">
          Configure your transcription preferences and paths.
        </p>
      </div>

      <Card className="p-6 rounded-sm shadow-none">
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold mb-4">Device Permissions</h2>
            <div className="space-y-3">
              <div className="border border-neutral-200 rounded-lg p-4">
                <div className="flex items-center justify-between gap-4">
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
                      <h3 className="text-sm font-semibold">Microphone</h3>
                      <p className="text-xs text-neutral-500">
                        Enable voice capture for quick recordings.
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className={`text-sm font-medium ${
                        micGranted ? "text-green-600" : "text-neutral-500"
                      }`}
                    >
                      {micGranted ? "Allowed" : "Not allowed"}
                    </span>
                    {!micGranted && (
                      <Button
                        onClick={handleRequestMicrophonePermission}
                        disabled={requestMicPermissionMutation.isPending}
                        size="sm"
                      >
                        {requestMicPermissionMutation.isPending
                          ? "Requesting..."
                          : "Allow"}
                      </Button>
                    )}
                  </div>
                </div>
                {!micGranted && (
                  <Button
                    variant="link"
                    size="sm"
                    className="px-0 mt-3"
                    onClick={() => OpenMicrophoneSettings()}
                  >
                    Open settings
                  </Button>
                )}
              </div>

              <div className="border border-neutral-200 rounded-lg p-4">
                <div className="flex items-center justify-between gap-4">
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
                      <h3 className="text-sm font-semibold">Accessibility</h3>
                      <p className="text-xs text-neutral-500">
                        Needed to listen for the Fn hotkey.
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className={`text-sm font-medium ${
                        accessibilityGranted
                          ? "text-green-600"
                          : "text-neutral-500"
                      }`}
                    >
                      {accessibilityGranted ? "Allowed" : "Not allowed"}
                    </span>
                    {!accessibilityGranted && (
                      <Button
                        onClick={handleRequestAccessibilityPermission}
                        disabled={
                          requestAccessibilityPermissionMutation.isPending
                        }
                        size="sm"
                      >
                        {requestAccessibilityPermissionMutation.isPending
                          ? "Requesting..."
                          : "Allow"}
                      </Button>
                    )}
                  </div>
                </div>
                {!accessibilityGranted && (
                  <Button
                    variant="link"
                    size="sm"
                    className="px-0 mt-3"
                    onClick={() => OpenAccessibilitySettings()}
                  >
                    Open settings
                  </Button>
                )}
              </div>
            </div>
          </div>

          <div>
            <h2 className="text-xl font-semibold mb-4">Transcription Method</h2>
            <div className="space-y-3">
              <div className="flex items-center space-x-2">
                <input
                  type="radio"
                  id="cloud"
                  name="transcriptionMethod"
                  value="cloud"
                  checked={transcriptionMethod === "cloud"}
                  onChange={(e) => setTranscriptionMethod(e.target.value)}
                  className="h-4 w-4"
                />
                <Label
                  htmlFor="cloud"
                  className="text-sm font-medium cursor-pointer"
                >
                  Cloud Transcription (Default)
                </Label>
              </div>
              <div className="ml-6">
                <p className="text-sm text-muted-foreground">
                  Uses Seisami's cloud service for fast and accurate
                  transcriptions.
                </p>
                <p className="text-xs text-amber-600 mt-1">
                  ⚠️ Requires login to use cloud features
                </p>
              </div>

              <div className="flex items-center space-x-2">
                <input
                  type="radio"
                  id="local"
                  name="transcriptionMethod"
                  value="local"
                  checked={transcriptionMethod === "local"}
                  onChange={(e) => setTranscriptionMethod(e.target.value)}
                  className="h-4 w-4"
                />
                <Label
                  htmlFor="local"
                  className="text-sm font-medium cursor-pointer"
                >
                  Local Whisper.cpp
                </Label>
              </div>
              <p className="text-sm text-muted-foreground ml-6">
                Uses your local whisper.cpp installation for offline
                transcriptions.
              </p>

              <div className="flex items-center space-x-2">
                <input
                  type="radio"
                  id="custom"
                  name="transcriptionMethod"
                  value="custom"
                  checked={transcriptionMethod === "custom"}
                  onChange={(e) => setTranscriptionMethod(e.target.value)}
                  className="h-4 w-4"
                />
                <Label
                  htmlFor="custom"
                  className="text-sm font-medium cursor-pointer"
                >
                  Custom OpenAI API
                </Label>
              </div>
              <p className="text-sm text-muted-foreground ml-6">
                Uses your own OpenAI API key for transcriptions.
              </p>
            </div>
          </div>

          {transcriptionMethod === "local" && (
            <div className="space-y-4 border-t pt-4">
              <h3 className="text-lg font-medium">Whisper.cpp Configuration</h3>

              <div className="space-y-2">
                <Label htmlFor="whisperBinary">Whisper Binary Path</Label>
                <div className="flex space-x-2">
                  <Input
                    id="whisperBinary"
                    value={whisperBinaryPath}
                    onChange={(e) => setWhisperBinaryPath(e.target.value)}
                    placeholder="/path/to/whisper-cli"
                    className="flex-1"
                  />
                  <Button onClick={handleSelectWhisperBinary} variant="outline">
                    Browse
                  </Button>
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="whisperModel">Whisper Model Path</Label>
                <div className="flex space-x-2">
                  <Input
                    id="whisperModel"
                    value={whisperModelPath}
                    onChange={(e) => setWhisperModelPath(e.target.value)}
                    placeholder="/path/to/model.bin"
                    className="flex-1"
                  />
                  <Button onClick={handleSelectWhisperModel} variant="outline">
                    Browse
                  </Button>
                </div>
              </div>
            </div>
          )}

          {transcriptionMethod === "custom" && (
            <div className="space-y-4 border-t pt-4">
              <h3 className="text-lg font-medium">OpenAI API Configuration</h3>

              <div className="space-y-2">
                <Label htmlFor="openaiKey">OpenAI API Key</Label>
                <Input
                  id="openaiKey"
                  type="password"
                  value={openaiApiKey}
                  onChange={(e) => setOpenaiApiKey(e.target.value)}
                  placeholder="sk-..."
                  className="w-full"
                />
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">
                    Your API key is stored locally and never shared.
                  </p>
                  <div className="border rounded-sm p-3 mt-2">
                    <p className="text-sm font-medium">
                      Note: This API key will be used for:
                    </p>
                    <ul className="text-sm mt-1 ml-4 list-disc space-y-0.5">
                      <li>Transcribing your audio recordings</li>
                      <li>
                        Processing transcriptions with AI (intent detection,
                        task creation, etc.)
                      </li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="flex justify-end pt-4 border-t">
            <Button
              onClick={handleSaveSettings}
              disabled={saveSettingsMutation.isPending}
            >
              {saveSettingsMutation.isPending ? "Saving..." : "Save Settings"}
            </Button>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default Settings;
