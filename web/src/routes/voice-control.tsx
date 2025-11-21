import { createFileRoute } from "@tanstack/react-router";
import { NavbarSection } from "@/components/navbar-section";
import { FooterSection } from "@/components/footer-section";
import { DitherBackground } from "@/components/ui/dither-background";
import { Mic, ArrowRight, Command, MessageSquare } from "lucide-react";

export const Route = createFileRoute("/voice-control")({
  component: VoiceControlPage,
});

function VoiceControlPage() {
  return (
    <div className="min-h-screen bg-white dark:bg-black text-black dark:text-white selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black">
      <NavbarSection />
      
      <main className="pt-32 pb-20 px-4 sm:px-6 lg:px-8 relative overflow-hidden">
        <DitherBackground opacity={0.1} />
        
        <div className="max-w-4xl mx-auto relative z-10">
          <div className="inline-flex items-center gap-2 px-3 py-1 mb-8 border border-black dark:border-white bg-white dark:bg-black text-xs font-mono uppercase tracking-wider">
            <Mic size={12} />
            Core Feature
          </div>
          
          <h1 className="text-5xl sm:text-7xl font-bold tracking-tighter mb-8 leading-[0.9] uppercase">
            Voice Control
          </h1>
          
          <p className="text-xl text-gray-600 dark:text-gray-400 mb-16 leading-relaxed max-w-2xl">
            Stop typing. Start doing. Seisami's voice engine is designed to understand intent, not just words.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-24">
            <div className="p-8 border border-black dark:border-white bg-white dark:bg-black">
              <Command className="w-8 h-8 mb-6" />
              <h3 className="text-xl font-bold mb-4 font-mono uppercase">Global Hotkey</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Press <span className="font-mono bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">Fn key</span> anywhere to start listening. Seisami captures audio instantly without bringing the window to focus.
              </p>
            </div>
            
            <div className="p-8 border border-black dark:border-white bg-white dark:bg-black">
              <MessageSquare className="w-8 h-8 mb-6" />
              <h3 className="text-xl font-bold mb-4 font-mono uppercase">Natural Language</h3>
              <p className="text-gray-600 dark:text-gray-400">
                "Move the design task to done and add a comment about the new logo." Seisami parses this and executes the actions.
              </p>
            </div>
          </div>

          <div className="border-t border-black dark:border-white pt-16">
            <h2 className="text-3xl font-bold mb-8 font-mono uppercase">Example Commands</h2>
            <ul className="space-y-6 font-mono text-sm sm:text-base">
              <li className="flex items-start gap-4">
                <ArrowRight className="w-5 h-5 mt-0.5 shrink-0" />
                <span>"Create a new card in Backlog called 'Research Competitors' due next Friday."</span>
              </li>
              <li className="flex items-start gap-4">
                <ArrowRight className="w-5 h-5 mt-0.5 shrink-0" />
                <span>"Add a checklist to the 'Launch Plan' card with items: Email blast, Social media, Blog post."</span>
              </li>
              <li className="flex items-start gap-4">
                <ArrowRight className="w-5 h-5 mt-0.5 shrink-0" />
                <span>"Archive all cards in the Done column."</span>
              </li>
            </ul>
          </div>
        </div>
      </main>

      <FooterSection />
    </div>
  );
}
