import { UncontrolledBoardProps } from "@caldwell619/react-kanban";
import { CustomCard } from "../data";

export const renderCard: UncontrolledBoardProps<CustomCard>["renderCard"] = (
  card,
  {}
) => {
  return (
    <div className="w-72 bg-white rounded-lg shadow p-2 my-1 flex flex-col gap-2 border border-gray-200">
      <div className="text-base font-semibold text-gray-900 mb-1">
        {card.title}
      </div>
      <div className="text-xs text-gray-500 mt-1">
        {card.createdAt.toDateString()}
      </div>
    </div>
  );
};

export const ColoredBgText = ({
  bgColor,
  children,
}: {
  bgColor: string;
  children: React.ReactNode;
}) => (
  <span
    className="rounded px-2 py-0.5 text-xs font-medium text-black my-5"
    style={{ backgroundColor: bgColor, marginRight: 10 }}
  >
    {children}
  </span>
);
