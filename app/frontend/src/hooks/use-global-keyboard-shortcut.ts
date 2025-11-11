import { useEffect } from "react";

export const useGlobalKeyboardShortcut = (
  key: string,
  callback: () => void,
  options?: { ctrlKey?: boolean; metaKey?: boolean; shiftKey?: boolean }
) => {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const matchesKey = e.key.toLowerCase() === key.toLowerCase();
      const matchesCtrl = options?.ctrlKey ? e.ctrlKey : true;
      const matchesMeta = options?.metaKey ? e.metaKey : true;
      const matchesShift = options?.shiftKey ? e.shiftKey : !e.shiftKey;

      if (matchesKey && (e.metaKey || e.ctrlKey) && !e.shiftKey) {
        e.preventDefault();
        callback();
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [key, callback, options]);
};
