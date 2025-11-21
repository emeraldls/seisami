import { Mic, Layout, Cloud, Shield, Zap, ArrowUpRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { Link } from "@tanstack/react-router";

interface BentoCardProps {
  title: string;
  description: string;
  icon: React.ReactNode;
  className?: string;
  pattern?: "dots" | "lines" | "grid" | "cross";
  large?: boolean;
  href?: string;
}

function DitherPattern({ type }: { type: "dots" | "lines" | "grid" | "cross" }) {
  const patterns = {
    dots: "radial-gradient(#000 1px, transparent 1px)",
    lines: "repeating-linear-gradient(45deg, #000 0, #000 1px, transparent 0, transparent 50%)",
    grid: "linear-gradient(#000 1px, transparent 1px), linear-gradient(90deg, #000 1px, transparent 1px)",
    cross: "radial-gradient(circle, transparent 20%, #000 20%, #000 80%, transparent 80%, transparent), radial-gradient(circle, transparent 20%, #000 20%, #000 80%, transparent 80%, transparent) 25px 25px"
  };

  const sizes = {
    dots: "4px 4px",
    lines: "10px 10px",
    grid: "20px 20px",
    cross: "50px 50px"
  };

  return (
    <div 
      className="absolute inset-0 opacity-[0.03] dark:invert pointer-events-none"
      style={{ 
        backgroundImage: patterns[type],
        backgroundSize: sizes[type]
      }} 
    />
  );
}

function BentoCard({ title, description, icon, className, pattern = "dots", large, href }: BentoCardProps) {
  const Content = (
    <>
      <DitherPattern type={pattern} />
      
      <div className="relative z-10 flex flex-col h-full justify-between">
        <div>
          <div className="mb-6 p-3 w-fit border border-black dark:border-white bg-white dark:bg-black group-hover:scale-110 transition-transform duration-300">
            {icon}
          </div>
          
          <h3 className={cn(
            "font-bold mb-4 font-mono uppercase tracking-tight",
            large ? "text-3xl sm:text-4xl" : "text-xl"
          )}>{title}</h3>
        </div>
        
        <div className="mt-auto">
           <p className={cn(
            "text-gray-600 dark:text-gray-400 leading-relaxed",
            large ? "text-lg max-w-md" : "text-sm"
          )}>{description}</p>
          
          {large && (
            <div className="mt-8 flex items-center gap-2 text-sm font-mono uppercase tracking-wider opacity-0 group-hover:opacity-100 transition-opacity transform translate-y-2 group-hover:translate-y-0">
              Learn more <ArrowUpRight size={16} />
            </div>
          )}
        </div>
      </div>

      <div className="absolute bottom-0 right-0 w-32 h-32 bg-gradient-to-tl from-black/5 to-transparent dark:from-white/5 pointer-events-none" />
    </>
  );

  if (href) {
    return (
      <Link to={href} className={cn(
        "group relative p-8 border border-black dark:border-white bg-white dark:bg-black overflow-hidden hover:bg-gray-50 dark:hover:bg-gray-900 transition-all duration-300 block",
        large ? "md:col-span-2 md:row-span-2" : "md:col-span-1",
        className
      )}>
        {Content}
      </Link>
    );
  }

  return (
    <div className={cn(
      "group relative p-8 border border-black dark:border-white bg-white dark:bg-black overflow-hidden hover:bg-gray-50 dark:hover:bg-gray-900 transition-all duration-300",
      large ? "md:col-span-2 md:row-span-2" : "md:col-span-1",
      className
    )}>
      {Content}
    </div>
  );
}

export function BentoGrid() {
  return (
    <section className="py-32 px-4 sm:px-6 lg:px-8 border-b border-border bg-gray-50 dark:bg-black/50">
      <div className="max-w-7xl mx-auto">
        <div className="mb-24">
          <h2 className="text-4xl sm:text-6xl font-bold mb-6 font-mono uppercase tracking-tighter">
            Engineered for<br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-black to-gray-500 dark:from-white dark:to-gray-500">
              Speed & Focus
            </span>
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-400 max-w-2xl leading-relaxed">
            Every interaction is designed to be instantaneous. No spinners, no waiting. Just you and your thoughts.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 auto-rows-[minmax(300px,auto)]">
          <BentoCard
            title="Voice Control"
            description="Speak naturally. Seisami understands context, creates tasks, and organizes your board without you lifting a finger."
            icon={<Mic className="w-8 h-8" />}
            large
            pattern="grid"
            href="/voice-control"
          />
          <BentoCard
            title="Local First"
            description="Your data lives on your device. Zero latency, works offline, complete privacy."
            icon={<Shield className="w-6 h-6" />}
            pattern="dots"
          />
          <BentoCard
            title="Kanban Board"
            description="Classic organization with a modern, minimal twist."
            icon={<Layout className="w-6 h-6" />}
            pattern="lines"
          />
          <BentoCard
            title="Lightning Fast"
            description="Built with Go and React for instant interactions. Native performance."
            icon={<Zap className="w-8 h-8" />}
            large
            pattern="lines"
            className="md:col-span-2 md:row-span-2"
            href="/lightning-fast"
          />
          <BentoCard
            title="Optional Sync"
            description="Sync with the cloud only when you want to. Keep your personal boards personal."
            icon={<Cloud className="w-6 h-6" />}
            pattern="cross"
          />
        </div>
      </div>
    </section>
  );
}
