import { useState, useCallback, useEffect } from "react";
import {
  KanbanProvider,
  KanbanBoard,
  KanbanHeader,
  KanbanCards,
  KanbanCard,
} from "~/components/ui/shadcn-io/kanban";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import {
  Plus,
  Paperclip,
  X,
  User,
  MoreHorizontal,
  Edit,
  Trash2,
  Save,
  Upload,
  Search,
  ChevronUp,
  ChevronDown,
  Mic,
  MicOff,
  Zap,
  Calendar,
  Clock,
  Tag,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import {
  CreateColumn,
  CreateCard,
  ListColumnsByBoard,
  ListCardsByColumn,
  UpdateCardColumn,
  UpdateCard,
  UpdateColumn,
  DeleteColumn,
  DeleteCard,
} from "../../wailsjs/go/main/App";
import { useBoardStore } from "~/stores/board-store";
import { Avatar, AvatarFallback } from "~/components/ui/avatar";
import { Badge } from "~/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";

type Feature = {
  id: string;
  name: string;
  description: string;
  attachments: string;
  startAt: Date;
  endAt: Date;
  column: string;
};

type Column = {
  id: string;
  name: string;
  color: string;
};

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
  year: "numeric",
});

const shortDateFormatter = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
});

export default function KanbanView() {
  const [columns, setColumns] = useState<Column[]>([]);
  const [features, setFeatures] = useState<Feature[]>([]);
  const [newCardName, setNewCardName] = useState("");
  const [addingCardInStatus, setAddingCardInStatus] = useState<string | null>(
    null
  );
  const [newColumnName, setNewColumnName] = useState("");
  const [isAddingColumn, setIsAddingColumn] = useState(false);
  const [selectedCard, setSelectedCard] = useState<Feature | null>(null);
  const [isCardDialogOpen, setIsCardDialogOpen] = useState(false);
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [editingDescription, setEditingDescription] = useState("");
  const [newAttachment, setNewAttachment] = useState("");
  const [dragStartTime, setDragStartTime] = useState<number | null>(null);
  const [draggedCardId, setDraggedCardId] = useState<string | null>(null);
  const [editingColumnId, setEditingColumnId] = useState<string | null>(null);
  const [editingColumnName, setEditingColumnName] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<Feature[]>([]);
  const [currentSearchIndex, setCurrentSearchIndex] = useState(0);
  const [highlightedCardId, setHighlightedCardId] = useState<string | null>(
    null
  );

  const { currentBoard } = useBoardStore();

  const fetchBoard = useCallback(async () => {
    if (!currentBoard) return;
    try {
      const columnsData = await ListColumnsByBoard(currentBoard.id);

      const transformedColumns: Column[] = columnsData.map((c) => ({
        id: c.ID.toString(),
        name: c.Name,
        color: "#10B981", //TODO: add a column in *column* table to include color
      }));

      setColumns(transformedColumns);

      const allFeatures: Feature[] = [];
      for (const col of columnsData) {
        const tickets = await ListCardsByColumn(col.ID);
        if (tickets) {
          const featuresForColumn: Feature[] = tickets.map((t) => ({
            id: t.ID.toString(),
            name: t.Title,
            description: t.Description.Valid ? t.Description.String : "",
            attachments: t.Attachments.Valid ? t.Attachments.String : "",
            startAt: new Date(t.CreatedAt.String!),
            endAt: new Date(
              t.UpdatedAt.Valid ? t.UpdatedAt.String! : t.CreatedAt.String!
            ),
            column: col.ID.toString(),
          }));
          allFeatures.push(...featuresForColumn);
        }
      }
      setFeatures(allFeatures);
    } catch (err) {
      console.error(err);
    }
  }, [currentBoard]);

  useEffect(() => {
    fetchBoard();
  }, [fetchBoard]);

  const handleDataChange = useCallback(
    async (newFeatures: Feature[]) => {
      for (const newFeature of newFeatures) {
        try {
          await UpdateCardColumn(newFeature.id, newFeature.column);
        } catch (err) {
          console.error("Failed to update card column", err);
          return;
        }
      }

      setFeatures(newFeatures);
    },
    [features]
  );

  const handleDragStart = (event: any) => {
    console.log("start dragging...");
    setDragStartTime(Date.now());
    setDraggedCardId(event.active.id);
  };

  const handleDragEnd = () => {
    console.log("stopped dragging");
    const dragDuration = dragStartTime ? Date.now() - dragStartTime : 0;

    if (dragDuration < 200 && draggedCardId) {
      const clickedFeature = features.find((f) => f.id === draggedCardId);
      if (clickedFeature) {
        setSelectedCard(clickedFeature);
        setIsCardDialogOpen(true);
      }
    }

    setDragStartTime(null);
    setDraggedCardId(null);
  };

  const handleAddCard = async (columnId: string) => {
    if (newCardName.trim() === "" || !currentBoard) return;

    try {
      await CreateCard(columnId, newCardName, "");
      setNewCardName("");
      setAddingCardInStatus(null);
      fetchBoard();
    } catch (err) {
      console.error("Failed to create card", err);
    }
  };

  const handleAddColumn = async () => {
    if (newColumnName.trim() === "" || !currentBoard) return;

    try {
      await CreateColumn(currentBoard.id, newColumnName);
      setNewColumnName("");
      setIsAddingColumn(false);
      fetchBoard();
    } catch (err) {
      console.error("Failed to create column", err);
    }
  };

  const handleEditColumn = (column: Column) => {
    setEditingColumnId(column.id);
    setEditingColumnName(column.name);
  };

  const handleSaveColumnEdit = async () => {
    if (editingColumnName.trim() === "" || !editingColumnId) return;

    try {
      await UpdateColumn(editingColumnId, editingColumnName);
      setEditingColumnId(null);
      setEditingColumnName("");
      fetchBoard();
    } catch (err) {
      console.error("Failed to update column", err);
    }
  };

  const handleDeleteColumn = async (columnId: string) => {
    try {
      await DeleteColumn(columnId);
      fetchBoard();
    } catch (err) {
      console.error("Failed to delete column", err);
    }
  };

  const handleDeleteCard = async (cardId: string) => {
    try {
      await DeleteCard(cardId);
      setIsCardDialogOpen(false);
      setSelectedCard(null);
      fetchBoard();
    } catch (err) {
      console.error("Failed to delete card", err);
    }
  };

  const handleStartEditDescription = () => {
    setEditingDescription(selectedCard?.description || "");
    setIsEditingDescription(true);
  };

  const handleSaveDescription = async () => {
    if (!selectedCard) return;

    try {
      await UpdateCard(selectedCard.id, selectedCard.name, editingDescription);
      setIsEditingDescription(false);
      fetchBoard();
    } catch (err) {
      console.error("Failed to update description", err);
    }
  };

  const handleCancelEditDescription = () => {
    setIsEditingDescription(false);
    setEditingDescription("");
  };

  const handleAddAttachment = async () => {
    if (!selectedCard || !newAttachment.trim()) return;

    try {
      // TODO: Need to add UpdateCardAttachments function to backend
      // Current UpdateCard only handles title and description
      console.log("Add attachment functionality needs backend support");
      setNewAttachment("");
      // fetchBoard();
    } catch (err) {
      console.error("Failed to add attachment", err);
    }
  };

  const handleRemoveAttachment = async (attachmentToRemove: string) => {
    if (!selectedCard) return;

    try {
      // TODO: Need to add UpdateCardAttachments function to backend
      // Current UpdateCard only handles title and description
      console.log("Remove attachment functionality needs backend support");
      // fetchBoard();
    } catch (err) {
      console.error("Failed to remove attachment", err);
    }
  };

  const performSearch = useCallback(
    (query: string) => {
      if (!query.trim()) {
        setSearchResults([]);
        setHighlightedCardId(null);
        setCurrentSearchIndex(0);
        return;
      }

      const lowercaseQuery = query.toLowerCase();
      const results: Feature[] = [];

      features.forEach((feature) => {
        const matchesTitle = feature.name
          .toLowerCase()
          .includes(lowercaseQuery);
        const matchesDescription = feature.description
          .toLowerCase()
          .includes(lowercaseQuery);

        const column = columns.find((col) => col.id === feature.column);
        const matchesColumn = column?.name
          .toLowerCase()
          .includes(lowercaseQuery);

        if (matchesTitle || matchesDescription || matchesColumn) {
          results.push(feature);
        }
      });

      setSearchResults(results);
      setCurrentSearchIndex(0);

      if (results.length > 0) {
        setHighlightedCardId(results[0].id);
        scrollToCard(results[0].id);
      } else {
        setHighlightedCardId(null);
      }
    },
    [features, columns]
  );

  const navigateSearchResults = (direction: "next" | "prev") => {
    if (searchResults.length === 0) return;

    let newIndex = currentSearchIndex;

    if (direction === "next") {
      newIndex = (currentSearchIndex + 1) % searchResults.length;
    } else {
      newIndex =
        currentSearchIndex === 0
          ? searchResults.length - 1
          : currentSearchIndex - 1;
    }

    setCurrentSearchIndex(newIndex);
    setHighlightedCardId(searchResults[newIndex].id);
    scrollToCard(searchResults[newIndex].id);
  };

  const scrollToCard = (cardId: string) => {
    const cardElement = document.querySelector(`[data-card-id="${cardId}"]`);
    if (cardElement) {
      cardElement.scrollIntoView({
        behavior: "smooth",
        block: "center",
        inline: "center",
      });
    }
  };

  const highlightText = (text: string, query: string) => {
    if (!query.trim()) return text;

    const regex = new RegExp(
      `(${query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`,
      "gi"
    );
    const parts = text.split(regex);

    return parts.map((part, index) =>
      regex.test(part) ? (
        <mark
          key={index}
          className="bg-yellow-200 text-yellow-900 px-1 rounded"
        >
          {part}
        </mark>
      ) : (
        part
      )
    );
  };

  const clearSearch = () => {
    setSearchQuery("");
    setSearchResults([]);
    setHighlightedCardId(null);
    setCurrentSearchIndex(0);
  };

  useEffect(() => {
    performSearch(searchQuery);
  }, [searchQuery, performSearch]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (searchResults.length === 0) return;

      if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        navigateSearchResults("next");
      } else if (e.key === "Enter" && e.shiftKey) {
        e.preventDefault();
        navigateSearchResults("prev");
      } else if (e.key === "Escape" && searchQuery) {
        clearSearch();
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [searchResults.length, searchQuery]);

  const getCardData = (feature: Feature) => {
    let attachments: string[] = [];
    if (feature.attachments && feature.attachments.trim()) {
      attachments = feature.attachments
        .split(",")
        .map((att) => att.trim())
        .filter((att) => att.length > 0);
    }

    return {
      ...feature,
      attachments,
    };
  };

  if (!currentBoard) {
    return (
      <div className=" min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="neural-pattern w-24 h-24 mx-auto  rounded-lg flex items-center justify-center mb-4">
            <Zap className="h-8 w-8" />
          </div>
          <h3 className="text-lg font-medium mb-2">No Board Selected</h3>
          <p className="text-gray-500">
            Please select a board to view your tasks
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className=" min-h-screen">
      <div className="bg-white border-b p-4">
        <div className="max-w-full flex items-center gap-4">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search cards, descriptions, columns..."
              className="pl-10 pr-10 border-gray-200 "
            />
            {searchQuery && (
              <Button
                variant="ghost"
                size="sm"
                onClick={clearSearch}
                className="absolute right-1 top-1/2 transform -translate-y-1/2 h-6 w-6 p-0"
              >
                <X className="h-3 w-3" />
              </Button>
            )}
          </div>

          {searchResults.length > 0 && (
            <div className="flex items-center gap-2">
              <span className="text-small text-gray-500 px-2">
                {currentSearchIndex + 1} of {searchResults.length}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigateSearchResults("prev")}
                className="h-8 w-8 p-0"
                title="Previous result (Shift+Enter)"
              >
                <ChevronUp className="h-3 w-3" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigateSearchResults("next")}
                className="h-8 w-8 p-0"
                title="Next result (Ctrl/Cmd+Enter)"
              >
                <ChevronDown className="h-3 w-3" />
              </Button>
            </div>
          )}

          {searchQuery && searchResults.length === 0 && (
            <span className="text-small text-gray-500">No results found</span>
          )}
        </div>
      </div>

      <div className="p-6 h-full w-full overflow-x-auto">
        {columns.length === 0 ? (
          <div className="h-full flex items-center justify-center">
            <div className="text-center">
              <div className="mb-4">
                <div className="w-24 h-24 mx-auto bg-muted/20 rounded-lg flex items-center justify-center mb-4">
                  <Plus className="h-8 w-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-semibold mb-2">No columns yet</h3>
                <p className="text-muted-foreground mb-6 max-w-sm">
                  Get started by creating your first column to organize your
                  tasks and workflow.
                </p>
              </div>
              <Button onClick={() => setIsAddingColumn(true)} className="gap-2">
                <Plus className="h-4 w-4" />
                Create your first column
              </Button>
              {isAddingColumn && (
                <div className="mt-4 max-w-sm mx-auto">
                  <Input
                    autoFocus
                    value={newColumnName}
                    onChange={(e) => setNewColumnName(e.target.value)}
                    onBlur={() => setIsAddingColumn(false)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") handleAddColumn();
                      if (e.key === "Escape") setIsAddingColumn(false);
                    }}
                    placeholder="Enter column name..."
                    className="mb-2"
                  />
                  <div className="flex gap-2">
                    <Button size="sm" onClick={handleAddColumn}>
                      Add Column
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => setIsAddingColumn(false)}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="flex gap-4 min-w-max">
            <KanbanProvider
              columns={columns}
              data={features}
              onDataChange={handleDataChange}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
              modifiers={[]}
            >
              {(column) => (
                <KanbanBoard
                  id={column.id}
                  key={column.id}
                  className="w-80 flex-shrink-0"
                >
                  <KanbanHeader>
                    <div>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8  flex items-center justify-center">
                            <Tag className="h-4 w-4" />
                          </div>
                          {editingColumnId === column.id ? (
                            <Input
                              value={editingColumnName}
                              onChange={(e) =>
                                setEditingColumnName(e.target.value)
                              }
                              onBlur={handleSaveColumnEdit}
                              onKeyDown={(e) => {
                                if (e.key === "Enter") handleSaveColumnEdit();
                                if (e.key === "Escape") {
                                  setEditingColumnId(null);
                                  setEditingColumnName("");
                                }
                              }}
                              className="h-6 text-sm ring-1 rounded-sm p-2 font-semibold bg-transparent border-none focus:ring-0"
                              autoFocus
                            />
                          ) : (
                            <span
                              className={
                                searchQuery &&
                                column.name
                                  .toLowerCase()
                                  .includes(searchQuery.toLowerCase())
                                  ? "bg-neutral-200 text-yellow-900 px-1 rounded"
                                  : ""
                              }
                            >
                              {searchQuery
                                ? highlightText(column.name, searchQuery)
                                : column.name}
                            </span>
                          )}
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-6 w-6 p-0"
                            >
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => handleEditColumn(column)}
                            >
                              <Edit className="h-4 w-4 mr-2" />
                              Edit name
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => handleDeleteColumn(column.id)}
                              className="text-red-600"
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              Delete column
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </div>
                  </KanbanHeader>
                  {features.filter((f) => f.column === column.id).length ===
                  0 ? (
                    <div className="p-6 text-center">
                      <div className="neural-pattern w-16 h-16 mx-auto rounded-lg flex items-center justify-center mb-4">
                        <Zap className="h-6 w-6 " />
                      </div>
                      <p className="text-sm text-gray-500 mb-4">
                        No cards in this column yet
                      </p>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setAddingCardInStatus(column.id)}
                        className="gap-2 hover"
                      >
                        <Plus className="h-3 w-3" />
                        Add first card
                      </Button>
                    </div>
                  ) : (
                    <KanbanCards id={column.id}>
                      {(feature: Feature) => {
                        const isHighlighted = highlightedCardId === feature.id;
                        const isInSearchResults = searchResults.some(
                          (result) => result.id === feature.id
                        );

                        return (
                          <KanbanCard
                            column={column.id}
                            id={feature.id}
                            key={feature.id}
                            name={feature.name}
                            className={`hover:shadow-lg transition-all duration-200 ${
                              isHighlighted
                                ? "ring-2 shadow-lg"
                                : isInSearchResults
                            }`}
                          >
                            <div
                              className="flex items-start justify-between gap-2 p-1"
                              data-card-id={feature.id}
                            >
                              <div className="flex flex-col gap-2 flex-1">
                                <div className="flex items-start justify-between gap-2">
                                  <p className="m-0 flex-1 font-medium text-sm">
                                    {searchQuery
                                      ? highlightText(feature.name, searchQuery)
                                      : feature.name}
                                  </p>
                                </div>

                                {feature.description && (
                                  <p className="m-0 text-xs text-gray-500 line-clamp-2">
                                    {searchQuery &&
                                    feature.description
                                      .toLowerCase()
                                      .includes(searchQuery.toLowerCase())
                                      ? highlightText(
                                          feature.description,
                                          searchQuery
                                        )
                                      : feature.description}
                                  </p>
                                )}

                                <div className="flex items-center justify-between text-xs">
                                  <div className="flex items-center gap-1 text-gray-400">
                                    <Calendar className="h-3 w-3" />
                                    <span>
                                      {shortDateFormatter.format(
                                        feature.startAt
                                      )}
                                    </span>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </KanbanCard>
                        );
                      }}
                    </KanbanCards>
                  )}

                  {addingCardInStatus === column.id ? (
                    <div className="p-4 ounded-lg border mx-2 mb-2">
                      <Input
                        autoFocus
                        value={newCardName}
                        onChange={(e) => setNewCardName(e.target.value)}
                        onBlur={() => setAddingCardInStatus(null)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") handleAddCard(column.id);
                          if (e.key === "Escape") setAddingCardInStatus(null);
                        }}
                        placeholder="Enter card title..."
                      />
                      <div className="flex gap-2 mt-3">
                        <Button
                          size="sm"
                          onClick={() => handleAddCard(column.id)}
                          className=" text-white"
                        >
                          <Plus className="h-3 w-3 mr-1" />
                          Add Card
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setAddingCardInStatus(null)}
                          className="text-gray-500 hover:text-gray-700"
                        >
                          Cancel
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <Button
                      variant="ghost"
                      className="m-2 justify-start"
                      onClick={() => setAddingCardInStatus(column.id)}
                    >
                      <Plus className="h-4 w-4 mr-2" /> Add a card
                    </Button>
                  )}
                </KanbanBoard>
              )}
            </KanbanProvider>

            <div className="p-2 w-80 flex-shrink-0">
              {isAddingColumn ? (
                <div>
                  <Input
                    autoFocus
                    value={newColumnName}
                    onChange={(e) => setNewColumnName(e.target.value)}
                    onBlur={() => setIsAddingColumn(false)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") handleAddColumn();
                      if (e.key === "Escape") setIsAddingColumn(false);
                    }}
                    placeholder="Enter column name..."
                  />
                  <Button size="sm" className="mt-2" onClick={handleAddColumn}>
                    Add Column
                  </Button>
                </div>
              ) : (
                <Button
                  variant="ghost"
                  className="w-full justify-start"
                  onClick={() => setIsAddingColumn(true)}
                >
                  <Plus className="h-4 w-4 mr-2" /> Add another list
                </Button>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Card Detail Dialog */}
      <Dialog open={isCardDialogOpen} onOpenChange={setIsCardDialogOpen}>
        <DialogContent
          className="max-w-4xl max-h-[90vh] overflow-y-auto"
          aria-describedby="sinzu"
        >
          {selectedCard && (
            <>
              <DialogHeader>
                <div className="flex items-center justify-between">
                  <DialogTitle className="text-md font-semibold">
                    {selectedCard.name}
                  </DialogTitle>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDeleteCard(selectedCard.id)}
                    className="text-red-600 hover:text-red-700 hover:bg-red-50"
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete
                  </Button>
                </div>
              </DialogHeader>

              <div className="space-y-6">
                <div>
                  <h3 className="text-sm font-medium flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <User className="h-4 w-4" />
                      Description
                    </div>
                    {!isEditingDescription && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={handleStartEditDescription}
                        className="h-6 w-6 p-0"
                      >
                        <Edit className="h-3 w-3" />
                      </Button>
                    )}
                  </h3>
                  <div className="bg-muted/20 rounded-lg p-4">
                    {isEditingDescription ? (
                      <div className="space-y-2">
                        <textarea
                          value={editingDescription}
                          onChange={(e) =>
                            setEditingDescription(e.target.value)
                          }
                          className="w-full min-h-24 p-2 text-sm bg-background border rounded resize-none focus:outline-none focus:ring-2 focus:ring-ring"
                          placeholder="Enter description..."
                        />
                        <div className="flex gap-2">
                          <Button size="sm" onClick={handleSaveDescription}>
                            Save
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={handleCancelEditDescription}
                          >
                            Cancel
                          </Button>
                        </div>
                      </div>
                    ) : (
                      <p className="text-sm text-muted-foreground leading-relaxed">
                        {selectedCard.description || "No description provided."}
                      </p>
                    )}
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
                    <Paperclip className="h-4 w-4" />
                    Attachments
                  </h3>

                  <div className="mb-4 flex gap-2">
                    <Input
                      value={newAttachment}
                      onChange={(e) => setNewAttachment(e.target.value)}
                      placeholder="Enter file name or URL..."
                      className="flex-1"
                    />
                    <Button
                      size="sm"
                      onClick={handleAddAttachment}
                      disabled={!newAttachment.trim()}
                    >
                      <Upload className="h-3 w-3 mr-1" />
                      Add
                    </Button>
                  </div>

                  <div className="space-y-2">
                    {getCardData(selectedCard).attachments.length > 0 ? (
                      getCardData(selectedCard).attachments.map(
                        (attachment: string, index: number) => (
                          <div
                            key={index}
                            className="flex items-center gap-3 p-3 border rounded-lg hover:bg-muted/20 transition-colors"
                          >
                            <div className="flex-shrink-0 w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center">
                              <span className="text-xs font-medium text-orange-600">
                                FILE
                              </span>
                            </div>
                            <div className="flex-grow min-w-0">
                              <div className="flex items-center gap-2">
                                <p className="font-medium text-sm truncate">
                                  {attachment}
                                </p>
                              </div>
                            </div>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleRemoveAttachment(attachment)}
                              className="h-6 w-6 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
                            >
                              <X className="h-3 w-3" />
                            </Button>
                          </div>
                        )
                      )
                    ) : (
                      <div className="text-center py-8">
                        <Paperclip className="h-8 w-8 mx-auto text-muted-foreground/40 mb-2" />
                        <p className="text-sm text-muted-foreground">
                          No attachments
                        </p>
                      </div>
                    )}
                  </div>
                </div>

                <div className="pt-4 border-t">
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>
                      Created: {dateFormatter.format(selectedCard.startAt)}
                    </span>
                    <span>
                      Updated: {dateFormatter.format(selectedCard.endAt)}
                    </span>
                  </div>
                </div>
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
