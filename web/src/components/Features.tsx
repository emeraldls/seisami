import {
  Mic,
  Brain,
  Zap,
  Lock,
  Share2,
  BarChart3,
  Code2,
  Smartphone,
} from "lucide-react";

const features = [
  {
    icon: Mic,
    title: "Voice-First Interface",
    description:
      "Record tasks with a hotkey. Seisami handles transcription and task creation automatically.",
  },
  {
    icon: Brain,
    title: "AI-Powered Processing",
    description:
      "GPT-4 understands context and extracts intents. Creates, updates, and organizes tasks intelligently.",
  },
  {
    icon: Lock,
    title: "Privacy First",
    description:
      "Local-first architecture. Your data stays on your machine. Optional cloud features.",
  },
  {
    icon: Zap,
    title: "Lightning Fast",
    description:
      "Optimized for speed. Instant feedback on every action, minimal latency.",
  },
  {
    icon: Share2,
    title: "Real-Time Collaboration",
    description:
      "Work with your team in real-time. Optional central server for sync across devices.",
  },
  {
    icon: Code2,
    title: "100% Open Source",
    description:
      "Built with love. Full source code available on GitHub. Contribute and customize.",
  },
];

export function Features() {
  return (
    <section
      id="features"
      className="py-20 px-4 sm:px-6 lg:px-8 bg-white dark:bg-black"
    >
      <div className="max-w-2xl mx-auto">
        <div className="text-center mb-14">
          <h2 className="text-3xl sm:text-4xl font-bold mb-4 text-black dark:text-white">
            Powerful Features
          </h2>
          <p className="text-base text-gray-600 dark:text-gray-400">
            Everything you need for voice-driven task management
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {features.map((feature, idx) => {
            const Icon = feature.icon;
            return (
              <div
                key={idx}
                className="p-5 rounded-lg bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 hover:border-gray-300 dark:hover:border-gray-700 transition-all"
              >
                <div className="w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-3">
                  <Icon className="text-gray-600 dark:text-white" size={20} />
                </div>
                <h3 className="font-semibold text-base mb-2 text-black dark:text-white">
                  {feature.title}
                </h3>
                <p className="text-gray-600 dark:text-gray-400 text-sm leading-relaxed">
                  {feature.description}
                </p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
