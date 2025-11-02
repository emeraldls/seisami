import React, { useState } from "react";
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
  // CheckAccessibilityPermission,
  // RequestAccessibilityPermission,
  // OpenAccessibilitySettings,
} from "../../wailsjs/go/main/App";
import { query, frontend } from "../../wailsjs/go/models";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

const Settings = () => {
  const queryClient = useQueryClient();
  const [transcriptionMethod, setTranscriptionMethod] =
    useState<string>("cloud");
  const [whisperBinaryPath, setWhisperBinaryPath] = useState<string>("");
  const [whisperModelPath, setWhisperModelPath] = useState<string>("");
  const [openaiApiKey, setOpenaiApiKey] = useState<string>("");
  const [checkingPermission, setCheckingPermission] = useState(false);
  const [checkingAccessibility, setCheckingAccessibility] = useState(false);

  const {
    data: settings,
    isLoading: loading,
    isError,
    error,
  } = useQuery({
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

  // Fetch microphone permission
  const { data: microphonePermission = -2, refetch: refetchMicPermission } =
    useQuery({
      queryKey: ["microphonePermission"],
      queryFn: CheckMicrophonePermission,
      retry: false,
    });

  // Fetch accessibility permission (mocked for now)
  const {
    data: accessibilityPermission = -2,
    refetch: refetchAccessibilityPermission,
  } = useQuery({
    queryKey: ["accessibilityPermission"],
    queryFn: async () => {
      // Temporarily mock this since the function isn't available yet
      console.log("Accessibility permission check not yet available");
      return -2;
    },
    retry: false,
  });

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
    onSuccess: (granted) => {
      refetchMicPermission();
      if (!granted) {
        const openSettings = confirm(
          "Microphone permission was denied. Would you like to open System Settings to grant permission manually?"
        );
        if (openSettings) {
          OpenMicrophoneSettings();
        }
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

  const handleRequestAccessibilityPermission = async () => {
    setCheckingAccessibility(true);
    try {
      // Temporarily mock this since the function isn't available yet
      console.log("Accessibility permission request not yet available");
      // For now, just open the settings
    } catch (error) {
      console.error("Failed to request accessibility permission:", error);
    } finally {
      setCheckingAccessibility(false);
    }
  };

  const getMicrophonePermissionStatus = () => {
    switch (microphonePermission) {
      case 1:
        return {
          text: "Authorized",
          color: "text-green-600",
          bg: "bg-green-50",
        };
      case 0:
        return { text: "Denied", color: "text-red-600", bg: "bg-red-50" };
      case -1:
        return {
          text: "Not Determined",
          color: "text-yellow-600",
          bg: "bg-yellow-50",
        };
      default:
        return {
          text: "Unknown",
          color: "text-neutral-600",
          bg: "bg-neutral-50",
        };
    }
  };

  const getAccessibilityPermissionStatus = () => {
    switch (accessibilityPermission) {
      case 1:
        return {
          text: "Authorized",
          color: "text-green-600",
          bg: "bg-green-50",
        };
      case 0:
        return { text: "Denied", color: "text-red-600", bg: "bg-red-50" };
      default:
        return {
          text: "Not Checked",
          color: "text-neutral-600",
          bg: "bg-neutral-50",
        };
    }
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
            <h2 className="text-xl font-semibold mb-4">
              Microphone Permissions
            </h2>
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg border">
                <div>
                  <h3 className="font-medium">Microphone Access</h3>
                  <p className="text-sm text-muted-foreground">
                    Required for audio recording functionality
                  </p>
                </div>
                <div className="flex items-center space-x-3">
                  <span
                    className={`px-3 py-1 rounded-full text-sm font-medium ${
                      getMicrophonePermissionStatus().color
                    } ${getMicrophonePermissionStatus().bg}`}
                  >
                    {getMicrophonePermissionStatus().text}
                  </span>
                  {microphonePermission !== 1 && (
                    <Button
                      onClick={handleRequestMicrophonePermission}
                      disabled={requestMicPermissionMutation.isPending}
                      variant="outline"
                      size="sm"
                    >
                      {requestMicPermissionMutation.isPending
                        ? "Checking..."
                        : "Request Permission"}
                    </Button>
                  )}
                  {microphonePermission === 0 && (
                    <Button
                      onClick={() => OpenMicrophoneSettings()}
                      variant="outline"
                      size="sm"
                    >
                      Open Settings
                    </Button>
                  )}
                </div>
              </div>
              {microphonePermission === 0 && (
                <div className="p-4 bg-white border border-neutral-200 rounded-lg">
                  <p className="text-sm text-neutral-800">
                    <strong>Microphone access is required</strong> for recording
                    functionality. Please grant permission in System Settings →
                    Privacy & Security → Microphone, then restart the
                    application.
                  </p>
                </div>
              )}
            </div>
          </div>

          <div>
            <h2 className="text-xl font-semibold mb-4">
              Accessibility Permissions
            </h2>
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg border">
                <div>
                  <h3 className="font-medium">Accessibility Access</h3>
                  <p className="text-sm text-muted-foreground">
                    Required for FN key monitoring and keyboard shortcuts
                  </p>
                </div>
                <div className="flex items-center space-x-3">
                  <span
                    className={`px-3 py-1 rounded-full text-sm font-medium ${
                      getAccessibilityPermissionStatus().color
                    } ${getAccessibilityPermissionStatus().bg}`}
                  >
                    {getAccessibilityPermissionStatus().text}
                  </span>
                  {accessibilityPermission !== 1 && (
                    <Button
                      onClick={handleRequestAccessibilityPermission}
                      disabled={checkingAccessibility}
                      variant="outline"
                      size="sm"
                    >
                      {checkingAccessibility
                        ? "Checking..."
                        : "Request Permission"}
                    </Button>
                  )}
                  {accessibilityPermission === 0 && (
                    <Button
                      onClick={() => {
                        console.log("Opening accessibility settings...");
                        alert(
                          "Please go to System Settings → Privacy & Security → Accessibility and add Seisami to the list of allowed applications."
                        );
                      }}
                      variant="outline"
                      size="sm"
                    >
                      Open Settings
                    </Button>
                  )}
                </div>
              </div>
              {accessibilityPermission === 0 && (
                <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                  <p className="text-sm text-yellow-800">
                    <strong>Accessibility access is required</strong> for FN key
                    monitoring. Please grant permission in System Settings →
                    Privacy & Security → Accessibility, add Seisami to the list,
                    then restart the application.
                  </p>
                </div>
              )}
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
              <p className="text-sm text-muted-foreground ml-6">
                Uses Seisami's cloud service for fast and accurate
                transcriptions.
              </p>

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
                <p className="text-sm text-muted-foreground">
                  Your API key is stored locally and never shared.
                </p>
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
