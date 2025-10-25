import { ArrowRight, Download } from "lucide-react";

export const CTA = () => {
  return (
    <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white dark:bg-black">
      <div className="max-w-2xl mx-auto">
        <div className="bg-white shadow-2xl dark:bg-gray-900 rounded-lg p-10 border border-gray-800 dark:border-gray-700 text-center">
          <h2 className="text-3xl font-bold mb-3">
            Ready to transform your productivity?
          </h2>
          <p className="text-base text-gray-800 dark:text-white mb-8">
            Download the desktop app or explore the source code on GitHub. No
            sign-up required. Works completely offline.
          </p>

          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <a
              href="https://github.com/emeraldls/seisami/releases/download/test-build6/Seisami-test-build6-macos.dmg"
              className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-lg bg-gray-800 text-white hover:bg-gray-700 font-semibold transition-colors text-sm"
            >
              <Download size={18} />
              Download for macOS
            </a>
            <a
              href="https://git.new/seisami"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-lg bg-gray-800 hover:bg-gray-700 text-white font-semibold transition-colors text-sm"
            >
              Star on GitHub
              <ArrowRight size={18} />
            </a>
          </div>
        </div>
      </div>
    </section>
  );
};
