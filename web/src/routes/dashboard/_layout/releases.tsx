import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useMemo, useState } from "react";
import {
  AlertCircle,
  Globe2,
  Loader2,
  Rocket,
  ShieldCheck,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
    <div className="space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Desktop Release Manager
        </h1>
        <p className="text-sm text-muted-foreground max-w-2xl">
          Provide a signed DMG download link and the release key to publish a
          new Seisami desktop build.
        </p>
      </div>

      {status && (
        <div
          role="status"
          className={cn(
            "flex items-start gap-3 rounded-md border p-4 text-sm",
            status.type === "success"
              ? "border-emerald-500/40 bg-emerald-500/10 text-emerald-700"
              : "border-destructive/40 bg-destructive/10 text-destructive"
          )}
        >
          {status.type === "success" ? (
            <ShieldCheck className="mt-0.5 h-4 w-4" />
          ) : (
            <AlertCircle className="mt-0.5 h-4 w-4" />
          )}
          <span>{status.message}</span>
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-xl">
              <Rocket className="h-5 w-5" /> Release a new version
            </CardTitle>
            <CardDescription>
              Paste the DMG (or installer) URL and the shared secret to trigger
              a release.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-6" onSubmit={handlePublish}>
              <div className="space-y-2">
                <label htmlFor="releaseUrl" className="text-sm font-medium">
                  Release URL
                </label>
                <Input
                  id="releaseUrl"
                  name="releaseUrl"
                  type="url"
                  placeholder="https://cdn.your-bucket.app/seisami-v1.2.3.dmg"
                  value={releaseUrl}
                  onChange={(event) => setReleaseUrl(event.target.value)}
                  required
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="versionKey" className="text-sm font-medium">
                  Version secret key
                </label>
                <Input
                  id="versionKey"
                  name="versionKey"
                  type="password"
                  placeholder="••••••••"
                  value={versionKey}
                  onChange={(event) => setVersionKey(event.target.value)}
                  required
                />
              </div>

              <div className="space-y-2">
                <label htmlFor="releaseNotes" className="text-sm font-medium">
                  Release notes / changelog
                </label>
                <textarea
                  id="releaseNotes"
                  name="releaseNotes"
                  placeholder=""
                  value={notes}
                  onChange={(event) => setNotes(event.target.value)}
                  required
                  className={cn(
                    "placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground",
                    "min-h-40 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none",
                    "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
                    "aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive"
                  )}
                />
              </div>

              <div className="flex items-center justify-between gap-3">
                <Button type="submit" disabled={isPublishing}>
                  {isPublishing ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      <span>Publishing…</span>
                    </>
                  ) : (
                    <>
                      <Rocket className="h-4 w-4" />
                      <span>Publish update</span>
                    </>
                  )}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => void loadLatestVersion()}
                  disabled={isLoadingLatest}
                >
                  {isLoadingLatest ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      <span>Refreshing…</span>
                    </>
                  ) : (
                    <>Refresh latest</>
                  )}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Globe2 className="h-4 w-4" /> Latest release
              </CardTitle>
              <CardDescription>
                The most recent version available to desktop clients.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoadingLatest ? (
                <div className="flex items-center gap-3 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>Fetching latest release…</span>
                </div>
              ) : latestError ? (
                <div className="space-y-2 text-sm">
                  <p className="font-medium text-destructive">
                    Failed to load latest release
                  </p>
                  <p className="text-muted-foreground">{latestError}</p>
                </div>
              ) : latestReleaseMeta ? (
                <div className="space-y-3 text-sm">
                  <div>
                    <p className="text-xs uppercase text-muted-foreground">
                      Version
                    </p>
                    <p className="text-lg font-semibold">
                      {latestReleaseMeta.version}
                    </p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs uppercase text-muted-foreground">
                      Download URL
                    </p>
                    <a
                      href={latestReleaseMeta.url}
                      target="_blank"
                      rel="noreferrer"
                      className="truncate text-sm text-primary hover:underline"
                    >
                      {latestReleaseMeta.url}
                    </a>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs uppercase text-muted-foreground">
                      SHA256
                    </p>
                    <p className="font-mono text-xs break-all">
                      {latestReleaseMeta.sha256}
                    </p>
                  </div>
                  {latestReleaseMeta.notes && (
                    <div className="space-y-1">
                      <p className="text-xs uppercase text-muted-foreground">
                        Notes
                      </p>
                      <p className="text-sm leading-relaxed">
                        {latestReleaseMeta.notes}
                      </p>
                    </div>
                  )}
                  <p className="text-xs text-muted-foreground">
                    Checked {latestReleaseMeta.fetchedAt}
                  </p>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">
                  No releases have been published yet.
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
