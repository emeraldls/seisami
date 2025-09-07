import {
  moveCard,
  KanbanBoard,
  UncontrolledBoard,
  OnDragEndNotification,
  ControlledBoard,
  // doingSomething,
} from "@caldwell619/react-kanban";
import "@caldwell619/react-kanban/dist/styles.css";
import {
  CreateColumn,
  GetBoards,
  ListColumnsByBoard,
  ListTicketsByColumn,
  UpdateTicketColumn,
} from "../../wailsjs/go/main/App";

import {
  renderColumnHeader,
  renderCard,
  renderColumnHeaderControllerBoard,
} from "~/components/notion/components";
import { useEffect, useState } from "react";
import { ColumnResponse, TicketResponse } from "~/types/types";
import { Input } from "~/components/ui/input";
import { Button } from "~/components/ui/button";
import { CustomCard, TicketType } from "~/components/notion/data";
import { useBoardStore } from "~/stores/board-store";

export const NotionDemo = () => {
  const [kanbanBoard, setKanbanBoard] = useState<KanbanBoard<CustomCard>>({
    columns: [],
  });

  const { currentBoard } = useBoardStore();

  async function fetchBoard() {
    if (!currentBoard) return;
    try {
      const columns = (await ListColumnsByBoard(
        currentBoard.id
      )) as ColumnResponse[];

      const boardColumns = await Promise.all(
        columns.map(async (col: ColumnResponse) => {
          const tickets = (await ListTicketsByColumn(col.ID)) as
            | TicketResponse[]
            | null;
          const cards: CustomCard[] =
            tickets?.map((t) => ({
              id: t.ID,
              title: t.Title,
              description: t.Description.String || "",
              assigneeId: t.AssigneeID.Int64 || 0,
              storyPoints: t.StoryPoints.Int64 || 0,
              ticketType: t.TicketType as TicketType,
              createdAt: new Date(t.CreatedAt.String!),
              prLink: t.PrLink.String || undefined,
            })) || [];

          return {
            id: col.ID,
            title: col.Title,
            position: col.Position,
            cards: cards,
          };
        })
      );

      setKanbanBoard({ columns: boardColumns });
    } catch (err) {
      console.error(err);
    }
  }

  useEffect(() => {
    const unsubscribeBoards = () => {
      fetchBoard();
    };

    return () => {
      unsubscribeBoards();
    };
  }, []);
  const [newColumnName, setNewColumnName] = useState("");

  const handleAddColumn = async () => {
    if (newColumnName.trim() === "") return;
    const boards = await GetBoards(1, 1);
    if (boards.length > 0) {
      const boardId = boards[0].ID;
      await CreateColumn(boardId, newColumnName);

      setNewColumnName("");
      fetchBoard();
    }
  };

  const handleCardMoveControllerBoard: OnDragEndNotification<
    CustomCard
  > = async (card, source, destination) => {
    setKanbanBoard((currentBoard) => {
      return moveCard(currentBoard, source, destination);
    });
    try {
      await UpdateTicketColumn(card.id, destination?.toColumnId as string);
    } catch (err) {
      console.error(err);
      // TODO: handle revert move
    }
  };

  // in uncontrolled board, this call back function from the lib is wrong, the first parameter is ref the board, second is card, third is then the source
  const handleCardMove = async (
    board: KanbanBoard<CustomCard>,
    card: CustomCard,
    source: { fromColumnId: string; fromPosition: number },
    destination: { toColumnId: string; toosition: number }
  ) => {
    setKanbanBoard((currentBoard) => {
      return moveCard(currentBoard, source, destination);
    });
    try {
      await UpdateTicketColumn(card.id, destination?.toColumnId as string);
    } catch (err) {
      console.error(err);
      // TODO: handle revert move
    }
  };

  // doingSomething();

  return (
    <>
      <div className="h-screen p-4">
        <div className="flex gap-2 mb-4">
          <Input
            value={newColumnName}
            onChange={(e) => setNewColumnName(e.target.value)}
            placeholder="New column name"
            className="w-48"
          />
          <Button onClick={handleAddColumn}>Add Column</Button>
        </div>
        <div className="h-scre h-full">
          {/* <ControlledBoard
            key={JSON.stringify(kanbanBoard.columns.map((c) => c.id))}
            renderColumnHeader={renderColumnHeaderControllerBoard}
            renderCard={renderCard}
            onCardDragEnd={handleCardMoveControllerBoard}
          >
            {kanbanBoard}
          </ControlledBoard> */}
          <UncontrolledBoard
            key={JSON.stringify(kanbanBoard.columns.map((c) => c.id))}
            initialBoard={kanbanBoard}
            renderColumnHeader={renderColumnHeader}
            renderCard={renderCard}
            // @ts-expect-error - the library provides invalid typescript signature which causes linting error
            onCardDragEnd={handleCardMove}
            renderColumnAdder={() => {
              return <div>Hello world</div>;
            }}
            // onNewCardConfirm={(draftCard) => ({
            //   ...draftCard,
            //   id: new Date().getTime(),
            // })}

            onCardNew={console.log}

            // allowAddCard
          />
        </div>
      </div>
    </>
  );
};

const Board = () => {
  return <NotionDemo />;
};

export default Board;
