import { Mic, ArrowRight, CheckCircle2 } from "lucide-react";

export function HowItWorks() {
  const steps = [
    {
      number: "01",
      title: "Speak",
      description: "Press the hotkey and just talk. Describe your task, project updates, or ideas naturally.",
      icon: <Mic className="w-6 h-6" />
    },
    {
      number: "02",
      title: "Process",
      description: "Seisami's local AI engine analyzes your speech, extracting intent, tags, and deadlines.",
      icon: <ArrowRight className="w-6 h-6" />
    },
    {
      number: "03",
      title: "Organize",
      description: "Your board updates instantly. Cards are created, moved, or edited based on your command.",
      icon: <CheckCircle2 className="w-6 h-6" />
    }
  ];

  return (
    <section className="py-32 px-4 sm:px-6 lg:px-8 border-b border-border bg-white dark:bg-black">
      <div className="max-w-7xl mx-auto">
        <div className="mb-24">
          <span className="text-sm font-mono uppercase tracking-widest text-gray-500 mb-4 block">Workflow</span>
          <h2 className="text-4xl sm:text-6xl font-bold mb-6 font-mono uppercase tracking-tighter">
            From thought to<br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-black to-gray-500 dark:from-white dark:to-gray-500">
              Action
            </span>
          </h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-12 relative">
          {/* Connecting Line */}
          <div className="hidden md:block absolute top-12 left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-black/20 dark:via-white/20 to-transparent" />

          {steps.map((step, index) => (
            <div key={index} className="relative group">
              <div className="mb-8 relative">
                <div className="w-24 h-24 bg-gray-50 dark:bg-gray-900 border border-black dark:border-white flex items-center justify-center text-4xl font-mono font-bold group-hover:scale-110 transition-transform duration-300">
                  {step.number}
                </div>
                <div className="absolute -bottom-4 -right-4 w-12 h-12 bg-black dark:bg-white text-white dark:text-black flex items-center justify-center rounded-full">
                  {step.icon}
                </div>
              </div>
              
              <h3 className="text-2xl font-bold mb-4 font-mono uppercase">{step.title}</h3>
              <p className="text-gray-600 dark:text-gray-400 leading-relaxed max-w-sm">
                {step.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
