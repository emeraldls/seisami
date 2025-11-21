import { Plus } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";

const faqs = [
  {
    question: "Is it really free?",
    answer: "Yes, the local version is completely free and open source. We only charge for cloud sync and team features."
  },
  {
    question: "Do I need an internet connection?",
    answer: "No. Seisami is local-first. It works perfectly offline. You only need internet if you want to sync across devices."
  },
  {
    question: "Where is my data stored?",
    answer: "On your device, in a standard SQLite database. You have full access to it. If you enable sync, it's encrypted and stored on our servers."
  },
  {
    question: "Which languages are supported?",
    answer: "Currently English, with support for Spanish, French, and German coming soon."
  },
  {
    question: "Is there a Windows/Linux version?",
    answer: "We are currently in beta for macOS. Windows and Linux versions are on the roadmap for Q4 2025."
  }
];

function FAQItem({ question, answer }: { question: string; answer: string }) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="border-b border-black/10 dark:border-white/10 last:border-0">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full py-6 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-white/5 transition-colors px-4"
      >
        <span className="font-mono font-bold uppercase tracking-tight pr-8">{question}</span>
        <Plus className={cn("w-4 h-4 transition-transform duration-200", isOpen && "rotate-45")} />
      </button>
      <div className={cn(
        "overflow-hidden transition-all duration-300 ease-in-out",
        isOpen ? "max-h-48 opacity-100" : "max-h-0 opacity-0"
      )}>
        <p className="pb-6 px-4 text-gray-600 dark:text-gray-400 leading-relaxed">
          {answer}
        </p>
      </div>
    </div>
  );
}

export function FAQ() {
  return (
    <section className="py-32 px-4 sm:px-6 lg:px-8 border-b border-border bg-white dark:bg-black">
      <div className="max-w-3xl mx-auto">
        <div className="mb-16 text-center">
          <h2 className="text-3xl sm:text-4xl font-bold mb-4 font-mono uppercase">Common Questions</h2>
        </div>

        <div className="border border-black dark:border-white">
          {faqs.map((faq, index) => (
            <FAQItem key={index} {...faq} />
          ))}
        </div>
      </div>
    </section>
  );
}
