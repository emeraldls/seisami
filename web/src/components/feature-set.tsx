import {
  Mic,
  Cpu,
  Workflow,
  Command,
  Users,
  Brain,
  Shield,
  GitBranch,
  Rocket,
  Zap,
} from "lucide-react";

export const features = [
  {
    icon: Mic,
    title: "Natural-Language Task Engine",
    description:
      "Speak normally. Seisami extracts intent, structure, priorities, and dependencies, not just text.",
  },
  {
    icon: Cpu,
    title: "Local-First Command Core",
    description:
      "All processing happens on-device for speed and privacy. No cloud lag. No data leaks.",
  },
  {
    icon: Workflow,
    title: "Autonomous Workflow Builder",
    description:
      "Give Seisami a goal. It generates boards, stages, and task groups automatically.",
  },
  {
    icon: Command,
    title: "Voice Driven Orchestration",
    description:
      "Trigger a hotkey and manage your entire board hands-free create, move, update, or reorganize tasks by voice.",
  },
  {
    icon: Users,
    title: "Zero-Setup Collaboration",
    description:
      "Invite teammates instantly. Local-first sync keeps everything fast, conflict-free, and reliable.",
  },
  {
    icon: Brain,
    title: "Contextual AI Memory",
    description:
      "Seisami remembers what you've said before references, priorities, and goals and uses them to guide new actions.",
  },
  {
    icon: Shield,
    title: "Explainable AI Actions",
    description:
      "Every AI-driven change includes a clear explanation. No hidden logic. Full transparency.",
  },
  {
    icon: GitBranch,
    title: "Extensible Developer Core",
    description:
      "Customize everything. Add actions, integrate models, or extend Seisami with modular APIs.",
  },
  {
    icon: Rocket,
    title: "Offline-Resilient Sync",
    description:
      "Work offline for days. Sync later. Seisami handles conflict resolution locally with minimal deltas.",
  },
  {
    icon: Zap,
    title: "Ultra-Fast Native Performance",
    description:
      "Built with Go and local indexing. Every command resolves in milliseconds no Electron bloat.",
  },
];

export const FeatureSet = () => {
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
};
