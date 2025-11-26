import { createFileRoute } from "@tanstack/react-router";
import { NavbarSection } from "@/components/navbar-section";
import { FooterSection } from "@/components/footer-section";
import { DitherBackground } from "@/components/ui/dither-background";
import { Zap, Cpu, Database, Layers } from "lucide-react";

export const Route = createFileRoute("/lightning-fast")({
  component: LightningFastPage,
});

function LightningFastPage() {
  return (
    <div className="min-h-screen bg-white dark:bg-black text-black dark:text-white selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black font-sans">
      <NavbarSection />

      <main className="pt-40 pb-20 px-4 sm:px-6 lg:px-8 relative overflow-hidden">
        <DitherBackground opacity={0.15} />

        <div className="max-w-5xl mx-auto relative z-10">
          <div className="flex flex-col items-start mb-20">
            <div className="inline-flex items-center gap-2 px-4 py-1.5 mb-8 border border-black dark:border-white bg-white dark:bg-black text-xs font-mono uppercase tracking-widest shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] dark:shadow-[4px_4px_0px_0px_rgba(255,255,255,1)]">
              <Zap size={14} />
              <span>Performance First</span>
            </div>

            <h1 className="text-6xl sm:text-8xl font-bold tracking-tighter mb-8 leading-[0.85] uppercase">
              Lightning
              <br />
              <span className="text-transparent bg-clip-text bg-linear-to-r from-black via-gray-500 to-transparent dark:from-white dark:via-gray-500">
                Fast
              </span>
            </h1>

            <p className="text-2xl text-gray-600 dark:text-gray-400 leading-relaxed max-w-2xl font-light">
              Built in Go. Designed to feel instant.
              <br />
              <span className="text-black dark:text-white font-medium">
                No loading screens, no server lag, everything happens locally.
              </span>
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-32">
            {[
              {
                icon: <Cpu className="w-6 h-6" />,
                title: "Native Go Core",
                desc: "The backend is written in Go with a tight, minimal runtime. No Electron bloat, no server roundtrips, just fast local execution.",
              },
              {
                icon: <Database className="w-6 h-6" />,
                title: "SQLite",
                desc: "Everything runs on a local SQLite database. Nanosecond-level queries with zero network latency. Your data stays on your machine.",
              },
              {
                icon: <Layers className="w-6 h-6" />,
                title: "Optimized UI",
                desc: "A minimal React layer with aggressive memoization. No wasted renders, no heavy frameworks. Just a responsive interface that stays out of your way.",
              },
            ].map((item, i) => (
              <div
                key={i}
                className="group p-8 border-l border-black dark:border-white hover:bg-gray-50 dark:hover:bg-white/5 transition-colors"
              >
                <div className="mb-6 opacity-50 group-hover:opacity-100 transition-opacity">
                  {item.icon}
                </div>
                <h3 className="text-xl font-bold mb-3 font-mono uppercase tracking-tight">
                  {item.title}
                </h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
                  {item.desc}
                </p>
              </div>
            ))}
          </div>

          <div className="relative border border-black dark:border-white p-8 sm:p-12 bg-white dark:bg-black overflow-hidden">
            <div className="absolute top-0 right-0 p-4 opacity-10">
              <Zap size={200} strokeWidth={0.5} />
            </div>

            <h2 className="text-3xl font-bold mb-12 font-mono uppercase relative z-10">
              Real-world Benchmarks
            </h2>

            <div className="space-y-10 relative z-10 max-w-3xl">
              {[
                {
                  label: "Cold Startup",
                  value: "180ms",
                  note: "Measured on my machine",
                  relative: 100,
                },
                {
                  label: "Voice Processing",
                  value: "500ms",
                  note: "Typical on-device result",
                  relative: 40,
                },
                {
                  label: "SQLite Read",
                  value: "0.4ms",
                  note: "Local query time",
                  relative: 2,
                },
              ].map((bench, i) => (
                <div key={i}>
                  <div className="flex justify-between text-sm font-mono mb-2 uppercase tracking-wider">
                    <span>{bench.label}</span>
                    <span className="font-bold">{bench.value}</span>
                  </div>

                  <div className="h-4 bg-gray-100 dark:bg-gray-900 w-full border border-black/10 dark:border-white/10">
                    <div
                      className="h-full bg-black dark:bg-white transition-all duration-1000 ease-out"
                      style={{ width: `${bench.relative}%` }}
                    />
                  </div>

                  <p className="text-xs font-mono mt-2 text-gray-500">
                    {bench.note}
                  </p>
                </div>
              ))}
            </div>

            <div className="mt-12 flex items-center gap-2 text-xs font-mono text-gray-500 uppercase tracking-widest">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              System Status: Optimal
            </div>
          </div>
        </div>
      </main>

      <FooterSection />
    </div>
  );
}
