import React from "react";
import { useEffect, useMemo, useState } from "react";

type DownloadProgressState = {
  percent: number | null;
  downloadedBytes: number;
  totalBytes: number;
};

import { EventsOn } from "../../wailsjs/runtime/runtime";
import { InstallUpdate } from "../../wailsjs/go/main/App";
import { types } from "../../wailsjs/go/models";
import { Download } from "lucide-react";
import { Button } from "./ui/button";
import { cn } from "~/lib/utils";

export const VersionUpdate = ({ collapsed }: { collapsed: boolean }) => {
  const [isDownloadingUpdate, setIsDownloadingUpdate] = useState(false);
  const [availableUpdate, setAvailableUpdate] =
    useState<types.AppVersion | null>(null);
  const [downloadProgress, setDownloadProgress] =
    useState<DownloadProgressState | null>(null);

  useEffect(() => {
    const unsubscribeUpdateAvailable = EventsOn(
      "update:available",
      (data: any) => {
        if (!data) {
          return;
        }

        setAvailableUpdate(types.AppVersion.createFrom(data));
      }
    );

    const unsubscribeDownloadStarted = EventsOn(
      "update:download_started",
      (data: any) => {
        const totalBytes =
          typeof data?.totalBytes === "number" ? data.totalBytes : -1;
        setIsDownloadingUpdate(true);
        setDownloadProgress({
          percent: totalBytes > 0 ? 0 : null,
          downloadedBytes: 0,
          totalBytes,
        });
      }
    );

    const unsubscribeDownloadProgress = EventsOn(
      "update:download_progress",
      (data: any) => {
        const downloadedBytes =
          typeof data?.downloadedBytes === "number" ? data.downloadedBytes : 0;
        const totalBytes =
          typeof data?.totalBytes === "number" ? data.totalBytes : -1;
        const percentValue =
          typeof data?.percent === "number" ? data.percent : null;

        setDownloadProgress({
          percent:
            totalBytes > 0 && percentValue !== null ? percentValue : null,
          downloadedBytes,
          totalBytes,
        });
      }
    );

    const unsubscribeDownloadComplete = EventsOn(
      "update:download_complete",
      (data: any) => {
        setDownloadProgress((prev) =>
          prev
            ? {
                percent: prev.totalBytes > 0 ? 100 : prev.percent,
                downloadedBytes:
                  typeof data?.downloadedBytes === "number"
                    ? data.downloadedBytes
                    : prev.downloadedBytes,
                totalBytes:
                  typeof data?.totalBytes === "number" && data.totalBytes >= 0
                    ? data.totalBytes
                    : prev.totalBytes,
              }
            : prev
        );
      }
    );

    const unsubscribeDownloadError = EventsOn(
      "update:download_error",
      (message: string) => {
        console.error("Update download error:", message);
        setDownloadProgress(null);
        setIsDownloadingUpdate(false);
      }
    );

    return () => {
      unsubscribeUpdateAvailable();
      unsubscribeDownloadStarted();
      unsubscribeDownloadProgress();
      unsubscribeDownloadComplete();
      unsubscribeDownloadError();
    };
  }, []);
  const handleDownloadUpdate = async () => {
    if (!availableUpdate) {
      return;
    }

    setIsDownloadingUpdate(true);
    setDownloadProgress({ percent: 0, downloadedBytes: 0, totalBytes: -1 });
    try {
      await InstallUpdate(availableUpdate);
      setAvailableUpdate(null);
    } catch (error) {
      console.error("Failed to download update:", error);
    } finally {
      setIsDownloadingUpdate(false);
      setDownloadProgress(null);
    }
  };

  const downloadLabel = useMemo(() => {
    if (!availableUpdate) {
      return "";
    }

    if (!isDownloadingUpdate) {
      return `Download ${availableUpdate.version}`;
    }

    if (!downloadProgress) {
      return "Downloading update...";
    }

    const downloadedMB = downloadProgress.downloadedBytes / 1024 / 1024;

    if (downloadProgress.totalBytes > 0) {
      const totalMB = downloadProgress.totalBytes / 1024 / 1024;
      return `Downloading ${downloadedMB.toFixed(2)} / ${totalMB.toFixed(
        2
      )} MB`;
    }

    return `Downloading ${downloadedMB.toFixed(2)} MB`;
  }, [availableUpdate, downloadProgress, isDownloadingUpdate]);

  const downloadFillPercent = useMemo(() => {
    if (!downloadProgress) {
      return null;
    }

    if (downloadProgress.percent !== null) {
      return Math.min(100, Math.max(0, downloadProgress.percent));
    }

    if (downloadProgress.totalBytes > 0) {
      const ratio =
        downloadProgress.downloadedBytes /
        Math.max(downloadProgress.totalBytes, 1);
      return Math.min(100, Math.max(0, ratio * 100));
    }

    return null;
  }, [downloadProgress]);
  return (
    <div>
      {availableUpdate && (
        <Button
          onClick={handleDownloadUpdate}
          disabled={isDownloadingUpdate}
          className={cn(
            "relative w-full overflow-hidden transition-colors",
            isDownloadingUpdate ? "bg-blue-500/10" : undefined
          )}
          size="sm"
        >
          {isDownloadingUpdate && downloadFillPercent !== null && (
            <span
              className="pointer-events-none absolute inset-y-0 left-0 bg-blue-500/70 transition-all duration-300 ease-out"
              style={{ width: `${downloadFillPercent}%` }}
              aria-hidden
            />
          )}
          {isDownloadingUpdate && downloadFillPercent === null && (
            <span
              className="pointer-events-none absolute inset-0 bg-gradient-to-r from-transparent via-blue-400/40 to-transparent animate-pulse"
              aria-hidden
            />
          )}
          <Download
            className={cn(
              "relative z-10 h-4 w-4",
              isDownloadingUpdate ? "text-black" : undefined
            )}
          />
          {!collapsed && (
            <span
              className={cn(
                "relative z-10 text-xs",
                isDownloadingUpdate ? "text-black" : undefined
              )}
            >
              {downloadLabel}
            </span>
          )}
        </Button>
      )}
    </div>
  );
};
