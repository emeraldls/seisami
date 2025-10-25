import { ArrowRight } from "lucide-react";
import AppUI from "@/assets/seisami-app.png";

export function Hero() {
  return (
    <section className="pt-32 pb-20 px-4 sm:px-6 lg:px-8">
      <div className="max-w-2xl mx-auto text-center mb-16">
        <h1 className="text-5xl sm:text-6xl font-bold tracking-tight mb-6 leading-tight text-black dark:text-white">
          Speak your tasks into existence
        </h1>

        <p className="text-lg text-gray-600 dark:text-gray-400 mb-10 leading-relaxed">
          Record audio with a hotkey. Seisami transcribes and intelligently
          creates tasks using AI. Privacy-first, works offline, optional cloud
          sync.
        </p>

        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          <a
            href="#features"
            className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-lg bg-black dark:bg-white hover:bg-gray-900 dark:hover:bg-gray-100 text-white dark:text-black font-semibold transition-colors text-sm"
          >
            Get Started
            <span className="ml-2 rounded-full bg-gray-200 px-2 py-0.5 text-xs font-semibold text-gray-600 dark:bg-gray-700 dark:text-gray-300">
              Beta
            </span>
          </a>
          <a
            href="https://git.new/seisami"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-lg bg-gray-100 dark:bg-gray-900 hover:bg-gray-200 dark:hover:bg-gray-800 text-black dark:text-white font-semibold transition-colors text-sm"
          >
            GitHub
            <ArrowRight size={18} />
          </a>
        </div>
      </div>

      <div className="relative w-full">
        <div className="max-w-6xl mx-auto">
          <img
            src={AppUI}
            alt="Seisami Application UI"
            className="w-full h-auto rounded-lg object-cover aspect-video"
          />
        </div>
      </div>
    </section>
  );
}
