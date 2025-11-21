import { Github, Moon, Sun } from "lucide-react";
import { useState, useEffect } from "react";
import logo from "../assets/logo.png";
import { cn } from "@/lib/utils";

export const NavbarSection = () => {
  const [isDark, setIsDark] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const isDarkMode =
      localStorage.getItem("theme") === "dark" ||
      (!localStorage.getItem("theme") &&
        window.matchMedia("(prefers-color-scheme: dark)").matches);
    setIsDark(isDarkMode);
    if (isDarkMode) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }

    const handleScroll = () => {
      setScrolled(window.scrollY > 50);
    };

    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const toggleDarkMode = () => {
    const newIsDark = !isDark;
    setIsDark(newIsDark);
    if (newIsDark) {
      document.documentElement.classList.add("dark");
      localStorage.setItem("theme", "dark");
    } else {
      document.documentElement.classList.remove("dark");
      localStorage.setItem("theme", "light");
    }
  };

  return (
    <header className="fixed top-0 w-full z-50 pointer-events-none">
      <div className={cn(
        "transition-all duration-500 ease-in-out flex items-center justify-between pointer-events-auto",
        scrolled 
          ? "mx-auto mt-4 w-[90%] max-w-2xl rounded-full border border-black/10 dark:border-white/10 bg-white/80 dark:bg-black/80 backdrop-blur-md shadow-lg px-6 py-3 translate-y-0 opacity-100"
          : "w-full px-4 sm:px-8 py-6 bg-transparent border-transparent translate-y-0"
      )}>
        <a href="/" className="flex items-center gap-3 group">
          <img src={logo} alt="Seisami Logo" className={cn(
            "transition-all duration-300",
            scrolled ? "h-6 w-6" : "h-8 w-8"
          )} />
          <span className={cn(
            "font-mono font-bold text-black dark:text-white uppercase tracking-wider transition-all",
            scrolled ? "text-sm" : "text-lg"
          )}>
            Seisami
          </span>
        </a>

        <div className="flex items-center gap-2 sm:gap-4">
          <a
            href="https://git.new/seisami"
            target="_blank"
            rel="noopener noreferrer"
            className={cn(
              "transition-colors text-black dark:text-white hover:opacity-70",
              scrolled ? "p-1" : "p-2"
            )}
            title="GitHub Repository"
          >
            <Github size={scrolled ? 16 : 20} />
          </a>
          <a
            href="https://x.com/tryseisami"
            target="_blank"
            rel="noopener noreferrer"
            className={cn(
              "transition-colors text-black dark:text-white hover:opacity-70",
              scrolled ? "p-1" : "p-2"
            )}
            title="Follow on X"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width={scrolled ? "14" : "18"}
              height={scrolled ? "14" : "18"}
              fill="currentColor"
              viewBox="0 0 16 16"
            >
              <path d="M12.6.75h2.454l-5.36 6.142L16 15.25h-4.937l-3.867-5.07-4.425 5.07H.316l5.733-6.57L0 .75h5.063l3.495 4.633L12.601.75Zm-.86 13.028h1.36L4.323 2.145H2.865z" />
            </svg>
          </a>
          <button
            onClick={toggleDarkMode}
            className={cn(
              "transition-colors text-black dark:text-white hover:opacity-70",
              scrolled ? "p-1" : "p-2"
            )}
            title="Toggle dark mode"
          >
            {isDark ? <Sun size={scrolled ? 16 : 20} /> : <Moon size={scrolled ? 16 : 20} />}
          </button>
        </div>
      </div>
    </header>
  );
};
