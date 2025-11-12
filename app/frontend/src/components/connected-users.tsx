import { useEffect, useState } from "react";
import { Avatar, AvatarFallback } from "~/components/ui/avatar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import { wsService } from "~/lib/websocket-service";

interface ConnectedUsersProps {
  boardId: string;
}

const getInitials = (userId: string) => {
  return userId.substring(0, 2).toUpperCase();
};

const getAvatarColor = (userId: string) => {
  const colors = [
    "bg-red-500",
    "bg-blue-500",
    "bg-green-500",
    "bg-yellow-500",
    "bg-purple-500",
    "bg-pink-500",
    "bg-indigo-500",
    "bg-orange-500",
  ];
  const index = userId.charCodeAt(0) % colors.length;
  return colors[index];
};

export const ConnectedUsers = ({ boardId }: ConnectedUsersProps) => {
  const [connectedUsers, setConnectedUsers] = useState<string[]>([]);

  useEffect(() => {
    const handleMessage = (message: any) => {
      if (message.type === "user_joined" || message.type === "user_left") {
        const users = message.users || [];
        const uniqueUsers = Array.from(new Set(users)) as string[];
        setConnectedUsers(uniqueUsers);
      }
    };

    const unsubscribe = wsService.onMessage(handleMessage);

    return () => {
      unsubscribe();
    };
  }, [boardId]);

  if (connectedUsers.length === 0) {
    return null;
  }

  const displayUsers = connectedUsers.slice(0, 5);
  const remainingCount = connectedUsers.length - 5;

  return (
    <div className="flex items-center gap-2">
      <div className="flex -space-x-2">
        {displayUsers.map((userId) => (
          <Popover key={userId}>
            <PopoverTrigger asChild>
              <Avatar
                className={`h-8 w-8 border-2 border-background cursor-pointer hover:z-10 transition-transform hover:scale-110 bg-black`}
              >
                <AvatarFallback className=" text-xs text-black font-semibold">
                  {getInitials(userId)}
                </AvatarFallback>
              </Avatar>
            </PopoverTrigger>
            <PopoverContent className="w-64 p-4" align="end">
              <div className="space-y-2">
                <div className="flex items-center gap-3">
                  <Avatar className={`h-10 w-10 `}>
                    <AvatarFallback className="font-semibold">
                      {getInitials(userId)}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="text-sm font-medium">User</p>
                    <p className="text-xs text-muted-foreground truncate">
                      {userId}
                    </p>
                  </div>
                </div>
                <div className="pt-2 border-t">
                  <div className="flex items-center gap-2">
                    <div className="h-2 w-2 rounded-full bg-green-500" />
                    <span className="text-xs text-muted-foreground">
                      Active now
                    </span>
                  </div>
                </div>
              </div>
            </PopoverContent>
          </Popover>
        ))}

        {remainingCount > 0 && (
          <Avatar className="h-8 w-8 border-2 border-background bg-muted">
            <AvatarFallback className="text-xs font-semibold">
              +{remainingCount}
            </AvatarFallback>
          </Avatar>
        )}
      </div>
      <span className="text-sm text-muted-foreground">
        {connectedUsers.length} online
      </span>
    </div>
  );
};
