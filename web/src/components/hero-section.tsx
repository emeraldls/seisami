import { ArrowRight, Apple, Download } from "lucide-react";
import { DitherBackground } from "./ui/dither-background";
import { ApiClient } from "@/lib/api-client";
import { useQuery } from "@tanstack/react-query";
import seisamiDemo from "@/assets/seisami-landing-output.mp4";
import { useEffect, useState } from "react";
import { DownloadAlert } from "./download-alert";

export const HeroSection = () => {
  const { data } = useQuery({
    queryKey: ["latest-app-version"],
    queryFn: () => ApiClient.getLatestAppVersion(),
  });

  const [videoSrc, setVideoSrc] = useState<string | null>(null);
  const [isDownloadModalOpen, setIsDownloadModalOpen] = useState(false);

  const handleDownload = () => {
    setIsDownloadModalOpen(true);
  };

  useEffect(() => {
    setVideoSrc(seisamiDemo);
  }, []);

  return (
    <>
      <section className="relative pt-20 min-h-[90vh] flex flex-col items-center justify-center px-4 sm:px-6 lg:px-8 overflow-hidden border-b border-border">
        <DitherBackground opacity={0.15} />

        <div className="relative z-10 max-w-4xl mx-auto text-center">
          {data?.data && (
            <div className="inline-flex items-center gap-2 px-3 py-1 mb-8 border border-black dark:border-white bg-white dark:bg-black text-xs font-mono tracking-wider">
              <span className="w-2 h-2 bg-black dark:bg-white animate-pulse" />
              {data.data.version} Public Beta
            </div>
          )}

          <h1 className="text-6xl sm:text-7xl md:text-8xl font-bold tracking-tighter mb-8 leading-[0.9] text-black dark:text-white uppercase">
            Control Your
            <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-b from-black to-gray-500 dark:from-white dark:to-gray-500">
              Workflow
            </span>
          </h1>

          <p className="text-lg sm:text-xl text-gray-600 dark:text-gray-400 mb-12 max-w-2xl mx-auto leading-relaxed font-light">
            Voice-first task management for the modern era.
            <br />
            Local-first, privacy-focused, optional cloud sync.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <a
              href={data?.data.url ?? "#"}
              onClick={handleDownload}
              className="group relative px-8 py-4 bg-black dark:bg-white text-white dark:text-black font-mono text-sm uppercase tracking-widest hover:bg-gray-900 dark:hover:bg-gray-100 transition-all active:translate-y-0.5"
            >
              <span className="relative z-10 flex items-center gap-2">
                <Apple size={18} className="mb-0.5" /> Download for macOS
              </span>
              <div className="absolute inset-0 border-2 border-transparent group-hover:border-black dark:group-hover:border-white translate-x-1 translate-y-1 transition-transform group-hover:translate-x-1.5 group-hover:translate-y-1.5" />
            </a>

            <a
              href="https://git.new/seisami"
              target="_blank"
              rel="noopener noreferrer"
              className="group px-8 py-4 bg-transparent border border-black dark:border-white text-black dark:text-white font-mono text-sm uppercase tracking-widest hover:bg-black hover:text-white dark:hover:bg-white dark:hover:text-black transition-all"
            >
              <span className="flex items-center gap-2">
                View Source <ArrowRight size={16} />
              </span>
            </a>
          </div>

          <div className="my-6 flex items-center justify-center gap-2 text-xs text-gray-500 font-mono">
            <Download size={12} />
            <span>Universal Binary (Intel & Apple Silicon)</span>
          </div>
        </div>

        <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-full max-w-5xl h-64 opacity-50 pointer-events-none">
          <div className="w-full h-full bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-gray-200 via-transparent to-transparent dark:from-gray-800" />
        </div>

        <div className="w-full max-w-4xl mx-auto">
          <div className="relative border-2 border-black dark:border-white bg-white dark:bg-black p-2">
            <div className="absolute -top-1 -left-1 w-3 h-3 bg-black dark:bg-white" />
            <div className="absolute -top-1 -right-1 w-3 h-3 bg-black dark:bg-white" />
            <div className="absolute -bottom-1 -left-1 w-3 h-3 bg-black dark:bg-white" />
            <div className="absolute -bottom-1 -right-1 w-3 h-3 bg-black dark:bg-white" />

            <video className="w-full h-auto" autoPlay loop muted playsInline>
              {videoSrc && <source src={videoSrc} type="video/mp4" />}
            </video>
          </div>
          <p className="mt-4 text-xs mb-8 text-gray-500 dark:text-gray-400 font-mono text-center uppercase tracking-wider">
            See the app in action
          </p>
        </div>
      </section>

      <DownloadAlert
        open={isDownloadModalOpen}
        onOpen={() => {
          setIsDownloadModalOpen(!isDownloadModalOpen);
        }}
        onClose={() => setIsDownloadModalOpen(false)}
      />
    </>
  );
};
