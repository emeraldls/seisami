import React, { useState, useEffect } from "react";
import { Card } from "~/components/ui/card";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
  GetSettings,
  SaveSettings,
  OpenFileDialog,
} from "../../wailsjs/go/main/App";
import { query, frontend } from "../../wailsjs/go/models";

const Settings = () => {
  const [settings, setSettings] = useState<query.Setting | null>(null);
  const [transcriptionMethod, setTranscriptionMethod] =
    useState<string>("cloud");
  const [whisperBinaryPath, setWhisperBinaryPath] = useState<string>("");
  const [whisperModelPath, setWhisperModelPath] = useState<string>("");
  const [openaiApiKey, setOpenaiApiKey] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const currentSettings = await GetSettings();
      setSettings(currentSettings);
      setTranscriptionMethod(currentSettings.TranscriptionMethod || "cloud");
      setWhisperBinaryPath(currentSettings.WhisperBinaryPath?.String || "");
      setWhisperModelPath(currentSettings.WhisperModelPath?.String || "");
      setOpenaiApiKey(currentSettings.OpenaiApiKey?.String || "");
    } catch (error) {
      console.error("Failed to load settings:", error);
      setTranscriptionMethod("cloud");
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSettings = async () => {
    setSaving(true);
    try {
      await SaveSettings(
        transcriptionMethod,
        whisperBinaryPath || null,
        whisperModelPath || null,
        openaiApiKey || null
      );
      alert("Settings saved successfully!");
    } catch (error) {
      console.error("Failed to save settings:", error);
      alert("Failed to save settings. Please try again.");
    } finally {
      setSaving(false);
    }
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
        <h1 className="text-2xl font-semibold text-gray-900">Settings</h1>
        <p className="text-sm text-gray-500 mt-1">
          Configure your transcription preferences and paths.
        </p>
      </div>

      <Card className="p-6 rounded-sm shadow-none">
        <div className="space-y-6">
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
            <Button onClick={handleSaveSettings} disabled={saving}>
              {saving ? "Saving..." : "Save Settings"}
            </Button>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default Settings;
