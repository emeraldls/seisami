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
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import {
  CreateColumn,
  CreateTicket,
  ListColumnsByBoard,
  ListTicketsByColumn,
  UpdateTicketColumn,
  UpdateColumn,
  DeleteColumn,
  DeleteTicket,
} from "../../wailsjs/go/main/App";
import { useBoardStore } from "~/stores/board-store";
import { TicketResponse } from "~/types/types";
import { Avatar, AvatarFallback } from "~/components/ui/avatar";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";

type Feature = {
  id: string;
  name: string;
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
  const [dragStartTime, setDragStartTime] = useState<number | null>(null);
  const [draggedCardId, setDraggedCardId] = useState<string | null>(null);
  const [editingColumnId, setEditingColumnId] = useState<string | null>(null);
  const [editingColumnName, setEditingColumnName] = useState("");

  const { currentBoard } = useBoardStore();

  const fetchBoard = useCallback(async () => {
    if (!currentBoard) return;
    try {
      const columnsData = await ListColumnsByBoard(currentBoard.id);

      const transformedColumns: Column[] = columnsData.map((c) => ({
        id: c.ID.toString(),
        name: c.Title,
        color: "#10B981", //TODO: add a column in *column* table to include color
      }));

      setColumns(transformedColumns);

      const allFeatures: Feature[] = [];
      for (const col of columnsData) {
        const tickets = (await ListTicketsByColumn(col.ID)) as
          | TicketResponse[]
          | null;
        if (tickets) {
          const featuresForColumn: Feature[] = tickets.map((t) => ({
            id: t.ID.toString(),
            name: t.Title,
            startAt: new Date(t.CreatedAt.String!),
            endAt: new Date(t.CreatedAt.String!),
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
          await UpdateTicketColumn(newFeature.id, newFeature.column);
        } catch (err) {
          console.error("Failed to update ticket column", err);
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
      await CreateTicket(columnId, newCardName, "", "Task");
      setNewCardName("");
      setAddingCardInStatus(null);
      fetchBoard();
    } catch (err) {
      console.error("Failed to create ticket", err);
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
      await DeleteTicket(cardId);
      setIsCardDialogOpen(false);
      setSelectedCard(null);
      fetchBoard();
    } catch (err) {
      console.error("Failed to delete card", err);
    }
  };

  const getCardData = (feature: Feature) => {
    return {
      ...feature,
      description: "This is a sample description for the task.",
      attachments: [
        {
          id: "2",
          name: "requirements.pdf",
          type: "pdf",
          size: "1.2 MB",
          addedDate: "5 Jan 2019, 14:23",
          isCover: false,
        },
      ],
      assignees: [
        { id: "1", name: "John Doe", initials: "JD" },
        { id: "2", name: "Jane Smith", initials: "JS" },
      ],
    };
  };

  if (!currentBoard) {
    return (
      <div className="p-4 h-full flex items-center justify-center">
        <p className="text-muted-foreground">Please select a board to view</p>
      </div>
    );
  }

  return (
    <div className="p-4 h-full w-full overflow-x-auto">
      {columns.length === 0 ? (
        <div className="h-full flex items-center justify-center">
          <div className="text-center">
            <div className="mb-4">
              <div className="w-24 h-24 mx-auto bg-muted/20 rounded-lg flex items-center justify-center mb-4">
                <Plus className="h-8 w-8 text-muted-foreground" />
              </div>
              <h3 className="text-lg font-semibold mb-2">No columns yet</h3>
              <p className="text-muted-foreground mb-6 max-w-sm">
                Get started by creating your first column to organize your tasks
                and workflow.
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
                  <div className="flex items-center justify-between w-full">
                    <div className="flex items-center gap-2">
                      <div
                        className="h-2 w-2 rounded-full"
                        style={{ backgroundColor: column.color }}
                      />
                      {editingColumnId === column.id ? (
                        <Input
                          value={editingColumnName}
                          onChange={(e) => setEditingColumnName(e.target.value)}
                          onBlur={handleSaveColumnEdit}
                          onKeyDown={(e) => {
                            if (e.key === "Enter") handleSaveColumnEdit();
                            if (e.key === "Escape") {
                              setEditingColumnId(null);
                              setEditingColumnName("");
                            }
                          }}
                          className="h-6 text-sm font-semibold bg-transparent border-none p-0 focus:ring-0"
                          autoFocus
                        />
                      ) : (
                        <span>{column.name}</span>
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
                </KanbanHeader>
                {features.filter((f) => f.column === column.id).length === 0 ? (
                  <div className="p-4 text-center">
                    <div className="w-16 h-16 mx-auto bg-muted/10 rounded-lg flex items-center justify-center mb-3">
                      <Plus className="h-6 w-6 text-muted-foreground" />
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">
                      No cards in this column yet
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setAddingCardInStatus(column.id)}
                      className="gap-2"
                    >
                      <Plus className="h-3 w-3" />
                      Add first card
                    </Button>
                  </div>
                ) : (
                  <KanbanCards id={column.id}>
                    {(feature: Feature) => (
                      <KanbanCard
                        column={column.id}
                        id={feature.id}
                        key={feature.id}
                        name={feature.name}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex flex-col gap-1">
                            <p className="m-0 flex-1 font-medium text-sm">
                              {feature.name}
                            </p>
                          </div>
                          <Avatar className="h-4 w-4 shrink-0">
                            <AvatarFallback>
                              {feature.name.slice(0, 2).toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                        </div>
                        <p className="m-0 text-muted-foreground text-xs">
                          {shortDateFormatter.format(feature.startAt)} -{" "}
                          {dateFormatter.format(feature.endAt)}
                        </p>
                      </KanbanCard>
                    )}
                  </KanbanCards>
                )}

                {addingCardInStatus === column.id ? (
                  <div className="p-2">
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
                    <Button
                      size="sm"
                      className="mt-2"
                      onClick={() => handleAddCard(column.id)}
                    >
                      Add Card
                    </Button>
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

      <Dialog open={isCardDialogOpen} onOpenChange={setIsCardDialogOpen}>
        <DialogContent
          className="max-w-4xl max-h-[90vh] overflow-y-auto"
          aria-describedby="sinzu"
        >
          {selectedCard && (
            <>
              <DialogHeader>
                <div className="flex items-center justify-between">
                  <DialogTitle className="text-xl font-semibold">
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
                  <h3 className="text-sm font-medium mb-2 flex items-center gap-2">
                    <User className="h-4 w-4" />
                    Description
                  </h3>
                  <div className="bg-muted/20 rounded-lg p-4">
                    <p className="text-sm text-muted-foreground leading-relaxed">
                      {getCardData(selectedCard).description}
                    </p>
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
                    <Paperclip className="h-4 w-4" />
                    Attachments
                  </h3>
                  <div className="space-y-2">
                    {getCardData(selectedCard).attachments.map((attachment) => (
                      <div
                        key={attachment.id}
                        className="flex items-center gap-3 p-3 border rounded-lg hover:bg-muted/20 transition-colors"
                      >
                        <div className="flex-shrink-0 w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center">
                          <span className="text-xs font-medium text-orange-600">
                            {attachment.type.toUpperCase()}
                          </span>
                        </div>
                        <div className="flex-grow min-w-0">
                          <div className="flex items-center gap-2">
                            <p className="font-medium text-sm truncate">
                              {attachment.name}
                            </p>
                            {attachment.isCover && (
                              <span className="text-xs bg-orange-100 text-orange-600 px-2 py-1 rounded">
                                Cover
                              </span>
                            )}
                          </div>
                          <p className="text-xs text-muted-foreground">
                            Added {attachment.addedDate} • {attachment.size}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-medium mb-3">Assignees</h3>
                  <div className="flex gap-2">
                    {getCardData(selectedCard).assignees.map((assignee) => (
                      <Avatar key={assignee.id} className="h-8 w-8">
                        <AvatarFallback className="text-xs">
                          {assignee.initials}
                        </AvatarFallback>
                      </Avatar>
                    ))}
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
