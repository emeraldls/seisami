import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";
import {
  AlertCircle,
  Globe2,
  Loader2,
  Rocket,
  ShieldCheck,
  Terminal,
  Download,
  Hash
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { ApiClient, type AppVersionInfo } from "@/lib/api-client";

export const Route = createFileRoute("/dashboard/_layout/releases")({
  component: ReleaseManagerPage,
});

function ReleaseManagerPage() {
  const [releaseUrl, setReleaseUrl] = useState("");
  const [versionKey, setVersionKey] = useState("");
  const [notes, setNotes] = useState("");
  const [isPublishing, setIsPublishing] = useState(false);
  const [status, setStatus] = useState<
    | { type: "success"; message: string }
    | { type: "error"; message: string }
    | null
  >(null);
  const [latestVersion, setLatestVersion] = useState<AppVersionInfo | null>(
    null
  );
  const [isLoadingLatest, setIsLoadingLatest] = useState(true);
  const [latestError, setLatestError] = useState<string | null>(null);

  useEffect(() => {
    void loadLatestVersion();
  }, []);

  const latestReleaseMeta = useMemo(() => {
    if (!latestVersion) {
      return null;
    }

    const date = new Date();
    const formattedDate = date.toLocaleString(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
    });

    return {
      version: latestVersion.version,
      url: latestVersion.url,
      sha256: latestVersion.sha256,
      notes: latestVersion.notes,
      fetchedAt: formattedDate,
    };
  }, [latestVersion]);

  const handlePublish = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setStatus(null);

    const trimmedUrl = releaseUrl.trim();
    const trimmedKey = versionKey.trim();
    const trimmedNotes = notes.trim();

    if (!trimmedUrl || !trimmedKey || !trimmedNotes) {
      setStatus({
        type: "error",
        message: "Release URL, version key, and release notes are required.",
      });
      return;
    }

    try {
      new URL(trimmedUrl);
    } catch (err) {
      setStatus({
        type: "error",
        message: "Please provide a valid release URL.",
      });
      return;
    }

    setIsPublishing(true);
    try {
      await ApiClient.publishAppVersion(trimmedUrl, trimmedKey, trimmedNotes);
      setStatus({
        type: "success",
        message:
          "Release published successfully. The desktop app will pick it up shortly.",
      });
      setReleaseUrl("");
      setVersionKey("");
      setNotes("");
      await loadLatestVersion();
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : "Unable to publish the new version.";
      setStatus({ type: "error", message });
    } finally {
      setIsPublishing(false);
    }
  };

  const loadLatestVersion = async () => {
    setIsLoadingLatest(true);
    setLatestError(null);
    try {
      const response = await ApiClient.getLatestAppVersion();
      setLatestVersion(response.data);
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : "Unable to fetch the latest release.";
      setLatestError(message);
      setLatestVersion(null);
    } finally {
      setIsLoadingLatest(false);
    }
  };

  return (
    <div className="space-y-12">
      <div className="space-y-4 border-b border-black dark:border-white pb-8">
        <h1 className="text-4xl font-bold tracking-tighter uppercase">
          Release Manager
        </h1>
        <p className="text-lg text-gray-600 dark:text-gray-400 max-w-2xl font-light">
          Deploy new versions of the Seisami desktop app. Updates are propagated instantly to all connected clients.
        </p>
      </div>

      {status && (
        <div
          role="status"
          className={cn(
            "flex items-start gap-4 border p-6 text-sm font-mono",
            status.type === "success"
              ? "border-green-500 bg-green-50 dark:bg-green-900/10 text-green-700 dark:text-green-400"
              : "border-red-500 bg-red-50 dark:bg-red-900/10 text-red-700 dark:text-red-400"
          )}
        >
          {status.type === "success" ? (
            <ShieldCheck className="mt-0.5 h-5 w-5 shrink-0" />
          ) : (
            <AlertCircle className="mt-0.5 h-5 w-5 shrink-0" />
          )}
          <span className="leading-relaxed">{status.message}</span>
        </div>
      )}

      <div className="grid gap-12 lg:grid-cols-[2fr_1fr]">
        <div className="space-y-8">
          <div className="flex items-center gap-2 mb-6">
            <div className="bg-black dark:bg-white text-white dark:text-black p-1">
              <Rocket size={16} />
            </div>
            <h2 className="text-xl font-bold font-mono uppercase tracking-wider">Deploy New Version</h2>
          </div>

          <form className="space-y-8" onSubmit={handlePublish}>
            <div className="space-y-2">
              <label htmlFor="releaseUrl" className="text-xs font-mono uppercase tracking-wider font-bold">
                Release URL (DMG)
              </label>
              <Input
                id="releaseUrl"
                name="releaseUrl"
                type="url"
                placeholder="https://cdn.your-bucket.app/seisami-v1.2.3.dmg"
                value={releaseUrl}
                onChange={(event) => setReleaseUrl(event.target.value)}
                required
                className="rounded-none border-black dark:border-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-black focus-visible:border-2 dark:focus-visible:border-white h-12 font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="versionKey" className="text-xs font-mono uppercase tracking-wider font-bold">
                Version Secret Key
              </label>
              <Input
                id="versionKey"
                name="versionKey"
                type="password"
                placeholder="••••••••"
                value={versionKey}
                onChange={(event) => setVersionKey(event.target.value)}
                required
                className="rounded-none border-black dark:border-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-black focus-visible:border-2 dark:focus-visible:border-white h-12 font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="releaseNotes" className="text-xs font-mono uppercase tracking-wider font-bold">
                Release Notes / Changelog
              </label>
              <textarea
                id="releaseNotes"
                name="releaseNotes"
                placeholder="- Fixed critical bug in voice processing..."
                value={notes}
                onChange={(event) => setNotes(event.target.value)}
                required
                className={cn(
                  "w-full min-h-[200px] rounded-none border border-black dark:border-white bg-transparent px-4 py-3 text-sm shadow-none outline-none font-mono",
                  "focus-visible:border-2 focus-visible:border-black dark:focus-visible:border-white placeholder:text-gray-400"
                )}
              />
            </div>

            <div className="flex items-center justify-between gap-4 pt-4">
              <Button 
                type="submit" 
                disabled={isPublishing}
                className="h-12 px-8 rounded-none bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 text-sm font-mono uppercase tracking-widest"
              >
                {isPublishing ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    Publishing...
                  </>
                ) : (
                  <>
                    <Rocket className="h-4 w-4 mr-2" />
                    Publish Update
                  </>
                )}
              </Button>
              
              <Button
                type="button"
                variant="outline"
                onClick={() => void loadLatestVersion()}
                disabled={isLoadingLatest}
                className="h-12 px-6 rounded-none border-black dark:border-white hover:bg-gray-50 dark:hover:bg-gray-900 text-sm font-mono uppercase tracking-widest"
              >
                {isLoadingLatest ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  "Refresh"
                )}
              </Button>
            </div>
          </form>
        </div>

        <div className="space-y-8">
          <div className="flex items-center gap-2 mb-6">
            <div className="bg-black dark:bg-white text-white dark:text-black p-1">
              <Globe2 size={16} />
            </div>
            <h2 className="text-xl font-bold font-mono uppercase tracking-wider">Live Status</h2>
          </div>

          <div className="border border-black dark:border-white p-6 bg-gray-50 dark:bg-gray-900/20">
            {isLoadingLatest ? (
              <div className="flex items-center gap-3 text-sm font-mono text-gray-500">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>Fetching latest release...</span>
              </div>
            ) : latestError ? (
              <div className="space-y-2 text-sm font-mono">
                <p className="font-bold text-red-600">
                  Failed to load latest release
                </p>
                <p className="text-gray-500">{latestError}</p>
              </div>
            ) : latestReleaseMeta ? (
              <div className="space-y-6">
                <div>
                  <p className="text-xs font-mono uppercase tracking-wider text-gray-500 mb-1">
                    Current Version
                  </p>
                  <div className="flex items-center gap-2">
                    <span className="text-2xl font-bold tracking-tight">{latestReleaseMeta.version}</span>
                    <span className="px-2 py-0.5 bg-green-100 text-green-800 text-[10px] font-mono uppercase tracking-wider rounded-full">Active</span>
                  </div>
                </div>
                
                <div>
                  <p className="text-xs font-mono uppercase tracking-wider text-gray-500 mb-1">
                    Download URL
                  </p>
                  <a
                    href={latestReleaseMeta.url}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center gap-2 text-sm font-mono text-blue-600 hover:underline break-all"
                  >
                    <Download size={12} />
                    {latestReleaseMeta.url}
                  </a>
                </div>
                
                <div>
                  <p className="text-xs font-mono uppercase tracking-wider text-gray-500 mb-1">
                    SHA256 Checksum
                  </p>
                  <div className="flex items-start gap-2 text-xs font-mono text-gray-600 dark:text-gray-400 break-all bg-white dark:bg-black p-2 border border-gray-200 dark:border-gray-800">
                    <Hash size={12} className="mt-0.5 shrink-0" />
                    {latestReleaseMeta.sha256}
                  </div>
                </div>
                
                {latestReleaseMeta.notes && (
                  <div>
                    <p className="text-xs font-mono uppercase tracking-wider text-gray-500 mb-1">
                      Release Notes
                    </p>
                    <div className="text-sm leading-relaxed text-gray-700 dark:text-gray-300 max-h-40 overflow-y-auto p-3 bg-white dark:bg-black border border-gray-200 dark:border-gray-800">
                      <pre className="whitespace-pre-wrap font-sans text-xs">{latestReleaseMeta.notes}</pre>
                    </div>
                  </div>
                )}
                
                <div className="pt-4 border-t border-gray-200 dark:border-gray-800">
                  <div className="flex items-center gap-2 text-xs font-mono text-gray-400">
                    <Terminal size={12} />
                    Last checked: {latestReleaseMeta.fetchedAt}
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-sm font-mono text-gray-500">
                  No releases have been published yet.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
