import {
  ControlledBoardProps,
  UncontrolledBoardProps,
} from "@caldwell619/react-kanban";
import randomRgba from "random-rgba";
import { CustomCard } from "../data";
import { ColoredBgText } from "./Card";
import { Plus, MoreHorizontal } from "lucide-react";
import { CreateTicket } from "../../../../wailsjs/go/main/App";
import { useBoardStore } from "~/stores/board-store";
import { TicketResponse } from "~/types/types";

export const renderColumnHeaderControllerBoard: ControlledBoardProps<CustomCard>["renderColumnHeader"] =
  (column) => {
    const { currentBoard } = useBoardStore();

    const onAddCard = async () => {
      if (!currentBoard) return;

      const newTicket = (await CreateTicket(
        "afc6fbc2-7174-42cd-bcef-15c726070098",
        "Hello world Title 2",
        "Hello world description 2",
        "hello world ticket type 2"
      )) as TicketResponse;

      const ticketToAdd: CustomCard = {
        id: newTicket.ID,
        title: newTicket.Title,
        description: newTicket.Description.String ?? "",
        assigneeId: newTicket.AssigneeID.Int64 ?? 0,
        storyPoints: newTicket.StoryPoints.Int64 ?? 0,
        ticketType: newTicket.TicketType ?? "",
        createdAt: new Date(newTicket.CreatedAt.String!),
      };

      // addCard(ticketToAdd, { on: "top" });
    };

    return (
      <div className="flex items-center">
        <div className="flex flex-grow items-center gap-2">
          <ColoredBgText bgColor={randomRgba(20)}>{column.title}</ColoredBgText>
          <span className="text-sm text-gray-600">{column.cards.length}</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            className="p-1 hover:bg-gray-100 rounded"
            aria-label="More"
          >
            <MoreHorizontal className="w-4 h-4 text-gray-500" />
          </button>
          <button
            type="button"
            className="p-1 hover:bg-gray-100 rounded"
            aria-label="Add"
            onClick={onAddCard}
          >
            <Plus className="w-4 h-4 text-gray-500" />
          </button>
        </div>
      </div>
    );
  };

export const renderColumnHeader: UncontrolledBoardProps<CustomCard>["renderColumnHeader"] =
  (column, { addCard }) => {
    const { currentBoard } = useBoardStore();

    const onAddCard = async () => {
      if (!currentBoard) return;

      const newTicket = (await CreateTicket(
        "afc6fbc2-7174-42cd-bcef-15c726070098",
        "Hello world Title 2",
        "Hello world description 2",
        "hello world ticket type 2"
      )) as TicketResponse;

      const ticketToAdd: CustomCard = {
        id: newTicket.ID,
        title: newTicket.Title,
        description: newTicket.Description.String ?? "",
        assigneeId: newTicket.AssigneeID.Int64 ?? 0,
        storyPoints: newTicket.StoryPoints.Int64 ?? 0,
        ticketType: newTicket.TicketType ?? "",
        createdAt: new Date(newTicket.CreatedAt.String!),
      };

      addCard(ticketToAdd, { on: "top" });
    };

    return (
      <div className="flex items-center">
        <div className="flex flex-grow items-center gap-2">
          <ColoredBgText bgColor={randomRgba(20)}>{column.title}</ColoredBgText>
          <span className="text-sm text-gray-600">{column.cards.length}</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            className="p-1 hover:bg-gray-100 rounded"
            aria-label="More"
          >
            <MoreHorizontal className="w-4 h-4 text-gray-500" />
          </button>
          <button
            type="button"
            className="p-1 hover:bg-gray-100 rounded"
            aria-label="Add"
            onClick={onAddCard}
          >
            <Plus className="w-4 h-4 text-gray-500" />
          </button>
        </div>
      </div>
    );
  };
