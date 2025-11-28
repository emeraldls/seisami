import { ApiClient } from "@/lib/api-client";
import { useQuery } from "@tanstack/react-query";
import { ArrowUpRight } from "lucide-react";
import { useState } from "react";
import { DownloadAlert } from "./download-alert";

export const FooterSection = () => {
  const currentYear = new Date().getFullYear();

  const { data } = useQuery({
    queryKey: ["latest-app-version"],
    queryFn: () => ApiClient.getLatestAppVersion(),
  });

  const [isDownloadModalOpen, setIsDownloadModalOpen] = useState(false);

  const handleDownload = () => {
    setIsDownloadModalOpen(true);
  };

  return (
    <footer className="bg-black text-white dark:bg-white dark:text-black pt-24 pb-12 px-4 sm:px-6 lg:px-8 overflow-hidden">
      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-16 mb-24">
          <div>
            <h2 className="text-6xl sm:text-8xl font-bold tracking-tighter mb-8 leading-[0.85] uppercase">
              Seisami
            </h2>
            <p className="text-xl text-gray-400 dark:text-gray-600 max-w-md leading-relaxed">
              The voice-first task manager for the modern era. <br />
              Stop typing. Start doing.
            </p>
          </div>

          <div className="flex flex-col justify-end items-start md:items-end">
            <a
              href={data?.data.url ?? "#"}
              className="group flex items-center gap-4 text-2xl sm:text-3xl font-bold font-mono uppercase hover:opacity-70 transition-opacity mb-8"
              onClick={handleDownload}
            >
              Download Beta{" "}
              <ArrowUpRight className="w-8 h-8 group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform" />
            </a>

            <div className="flex gap-8 text-sm font-mono uppercase tracking-wider text-gray-400 dark:text-gray-600">
              <a
                href="https://git.new/seisami"
                className="hover:text-white dark:hover:text-black transition-colors"
              >
                GitHub
              </a>
              <a
                href="https://x.com/tryseisami"
                className="hover:text-white dark:hover:text-black transition-colors"
              >
                Twitter / X
              </a>
              <a
                href="mailto:hello@seisami.com"
                className="hover:text-white dark:hover:text-black transition-colors"
              >
                Contact
              </a>
            </div>
          </div>
        </div>

        <div className="border-t border-white/20 dark:border-black/20 pt-8 flex flex-col sm:flex-row justify-between items-center gap-4 text-xs font-mono text-gray-500">
          <p>&copy; {currentYear} Seisami Inc. Open Source Software.</p>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full" />
            <span>All Systems Operational</span>
          </div>
        </div>
      </div>

      <DownloadAlert
        open={isDownloadModalOpen}
        onOpen={() => {
          setIsDownloadModalOpen(!isDownloadModalOpen);
        }}
        onClose={() => setIsDownloadModalOpen(false)}
      />
    </footer>
  );
};
