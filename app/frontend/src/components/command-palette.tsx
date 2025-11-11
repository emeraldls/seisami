import { useEffect, useState, useRef, useMemo } from "react";
import { Search, FileText, Layout, Plus, Folder } from "lucide-react";
import { cn } from "~/lib/utils";
import { useNavigate } from "react-router-dom";

type CommandItem = {
  id: string;
  label: string;
  description?: string;
  icon?: React.ComponentType<{ className?: string }>;
  onSelect: () => void;
  group: "cards" | "columns" | "actions" | "navigation";
};

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  cards?: Array<{
    id: string;
    name: string;
    description?: string;
    column: string;
  }>;
  columns?: Array<{ id: string; name: string }>;
  boardId?: string;
}

export const CommandPalette = ({
  isOpen,
  onClose,
  cards = [],
  columns = [],
  boardId,
}: CommandPaletteProps) => {
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const allCommands: CommandItem[] = useMemo(() => {
    const commands: CommandItem[] = [];

    commands.push(
      {
        id: "nav-home",
        label: "Home",
        icon: Layout,
        onSelect: () => {
          navigate("/");
          onClose();
        },
        group: "navigation",
      },
      {
        id: "nav-transcripts",
        label: "Transcripts",
        icon: FileText,
        onSelect: () => {
          navigate("/transcriptions");
          onClose();
        },
        group: "navigation",
      },
      {
        id: "nav-boards",
        label: "Manage Boards",
        icon: Folder,
        onSelect: () => {
          navigate("/boards");
          onClose();
        },
        group: "navigation",
      }
    );

    // Actions
    commands.push({
      id: "action-create-card",
      label: "Create new card",
      icon: Plus,
      onSelect: () => {
        // TODO: Implement card creation modal
        onClose();
      },
      group: "actions",
    });

    // Cards
    cards.forEach((card) => {
      const column = columns.find((c) => c.id === card.column);
      commands.push({
        id: `card-${card.id}`,
        label: card.name,
        description: column ? `in ${column.name}` : undefined,
        icon: FileText,
        onSelect: () => {
          const cardElement = document.querySelector(
            `[data-card-id="${card.id}"]`
          );
          if (cardElement) {
            cardElement.scrollIntoView({
              behavior: "smooth",
              block: "center",
              inline: "center",
            });
            // Brief highlight effect
            cardElement.classList.add("ring-2", "ring-primary", "shadow-lg");
            setTimeout(() => {
              cardElement.classList.remove(
                "ring-2",
                "ring-primary",
                "shadow-lg"
              );
            }, 2000);
          }
          onClose();
        },
        group: "cards",
      });
    });

    // Columns
    columns.forEach((column) => {
      commands.push({
        id: `column-${column.id}`,
        label: column.name,
        description: "Column",
        icon: Layout,
        onSelect: () => {
          const columnElement = document.querySelector(
            `[data-column-id="${column.id}"]`
          );
          if (columnElement) {
            columnElement.scrollIntoView({
              behavior: "smooth",
              block: "start",
              inline: "center",
            });
          }
          onClose();
        },
        group: "columns",
      });
    });

    return commands;
  }, [cards, columns, navigate, onClose]);

  const fuzzyMatch = (text: string, query: string): boolean => {
    const lowerText = text.toLowerCase();
    const lowerQuery = query.toLowerCase();

    if (lowerText.includes(lowerQuery)) return true;

    let queryIndex = 0;
    for (
      let i = 0;
      i < lowerText.length && queryIndex < lowerQuery.length;
      i++
    ) {
      if (lowerText[i] === lowerQuery[queryIndex]) {
        queryIndex++;
      }
    }
    return queryIndex === lowerQuery.length;
  };

  const filteredCommands = useMemo(() => {
    if (!query.trim()) return allCommands;

    return allCommands.filter((cmd) => {
      const searchText = `${cmd.label} ${cmd.description || ""}`;
      return fuzzyMatch(searchText, query);
    });
  }, [query, allCommands]);

  const groupedCommands = useMemo(() => {
    const groups: Record<string, CommandItem[]> = {
      navigation: [],
      actions: [],
      cards: [],
      columns: [],
    };

    filteredCommands.forEach((cmd) => {
      groups[cmd.group].push(cmd);
    });

    return groups;
  }, [filteredCommands]);

  useEffect(() => {
    if (isOpen) {
      setQuery("");
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [isOpen]);

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return;

      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelectedIndex((prev) =>
          Math.min(prev + 1, filteredCommands.length - 1)
        );
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelectedIndex((prev) => Math.max(prev - 1, 0));
      } else if (e.key === "Enter") {
        e.preventDefault();
        if (filteredCommands[selectedIndex]) {
          filteredCommands[selectedIndex].onSelect();
        }
      } else if (e.key === "Escape") {
        e.preventDefault();
        onClose();
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, selectedIndex, filteredCommands, onClose]);

  useEffect(() => {
    const selectedElement = document.querySelector(
      `[data-command-index="${selectedIndex}"]`
    );
    if (selectedElement) {
      selectedElement.scrollIntoView({ block: "nearest", behavior: "smooth" });
    }
  }, [selectedIndex]);

  if (!isOpen) return null;

  const renderGroup = (groupName: string, items: CommandItem[]) => {
    if (items.length === 0) return null;

    const groupLabels: Record<string, string> = {
      navigation: "Navigation",
      actions: "Actions",
      cards: "Cards",
      columns: "Columns",
    };

    let currentIndex = 0;
    for (const key in groupedCommands) {
      if (key === groupName) break;
      currentIndex +=
        groupedCommands[key as keyof typeof groupedCommands].length;
    }

    return (
      <div key={groupName} className="pb-2">
        <div className="px-3 py-1.5 text-xs font-medium text-muted-foreground">
          {groupLabels[groupName]}
        </div>
        {items.map((item, idx) => {
          const absoluteIndex = currentIndex + idx;
          const Icon = item.icon;
          return (
            <button
              key={item.id}
              data-command-index={absoluteIndex}
              onClick={item.onSelect}
              className={cn(
                "w-full flex items-center gap-3 px-3 py-2 text-sm text-left transition-colors",
                absoluteIndex === selectedIndex
                  ? "bg-accent text-accent-foreground"
                  : "hover:bg-accent/50"
              )}
            >
              {Icon && (
                <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
              )}
              <div className="flex-1 min-w-0">
                <div className="truncate">{item.label}</div>
                {item.description && (
                  <div className="text-xs text-muted-foreground truncate">
                    {item.description}
                  </div>
                )}
              </div>
            </button>
          );
        })}
      </div>
    );
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh] bg-black/50"
      onClick={onClose}
    >
      <div
        className="w-full max-w-2xl bg-background rounded-lg shadow-2xl border overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 px-4 py-3 border-b">
          <Search className="h-5 w-5 text-muted-foreground shrink-0" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search for cards, columns, or actions..."
            className="flex-1 bg-transparent outline-none text-sm"
          />
          <kbd className="hidden sm:inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
            ESC
          </kbd>
        </div>

        <div className="max-h-[400px] overflow-y-auto py-2">
          {filteredCommands.length === 0 ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              No results found
            </div>
          ) : (
            <>
              {renderGroup("navigation", groupedCommands.navigation)}
              {renderGroup("actions", groupedCommands.actions)}
              {renderGroup("cards", groupedCommands.cards)}
              {renderGroup("columns", groupedCommands.columns)}
            </>
          )}
        </div>
      </div>
    </div>
  );
};
