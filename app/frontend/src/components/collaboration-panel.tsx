import { useEffect, useMemo, useState } from "react";
import { Users, Copy, Share2, LogOut, PlusCircle } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Badge } from "~/components/ui/badge";
import {
  useCollaborationStore,
  type CollabStatus,
} from "~/stores/collab-store";
import { cn } from "~/lib/utils";
import { toast } from "sonner";

const statusConfig: Record<
  CollabStatus,
  { label: string; badgeClass: string }
> = {
  disconnected: {
    label: "Disconnected",
    badgeClass: "bg-gray-200 text-gray-700",
  },
  connected: {
    label: "Connected",
    badgeClass: "bg-emerald-100 text-emerald-700",
  },
  "in-room": { label: "In a Room", badgeClass: "bg-blue-100 text-blue-700" },
  busy: { label: "Working...", badgeClass: "bg-amber-100 text-amber-700" },
  error: { label: "Error", badgeClass: "bg-red-100 text-red-700" },
};

/*
  Next step: update user local changes & broadcast to connected clients, if a clients just comes online, 
  im to pull changes for that room & check if the client has local changes, validate conflicts & merge both data, last write wins

  ------ Logic for when a card is updated on one user's end & on the other user's end is either dragging the same card -----
  - Update the current user local state, broadcast changes to connected clients, if a client is currently dragging an item or doing something, i need to track
  a client state to know if they're performing an action & if they're, we shouldnt update their board yet, until their mouse if off the mouse, then i'll pull changes


  - I can implmenent a custom cursor that shows in the other user's end where the current user's cursor is atm in the application, just like what figma does -- coool idea, for this, I will just get the cursor position & keep emitting a user cursor position to connected clients, then on connected clients end, i just draw the cursor & it overlays on their board. Cool
*/

export const CollaborationPanel = () => {
  const [action, setAction] = useState<"create" | "join" | "leave" | null>(
    null
  );
  const [joinInput, setJoinInput] = useState("");
  const {
    initialize,
    status,
    roomId,
    address,
    lastError,
    createRoom,
    joinRoom,
    leaveRoom,
  } = useCollaborationStore();

  useEffect(() => {
    initialize();
  }, [initialize]);

  useEffect(() => {
    if (roomId) {
      setJoinInput(roomId);
    }
  }, [roomId]);

  const statusMeta = useMemo(() => statusConfig[status], [status]);
  const isBusy = status === "busy" && action !== null;
  const isInRoom = status === "in-room" && !!roomId;

  const handleCreateRoom = async () => {
    setAction("create");
    try {
      const created = await createRoom();
      if (created) {
        setJoinInput(created);
      }
    } finally {
      setAction(null);
    }
  };

  const handleJoinRoom = async () => {
    if (!joinInput.trim()) {
      toast.error("Room ID is required", {
        description: "Enter a valid room ID to join",
      });
      return;
    }

    setAction("join");
    try {
      const joined = await joinRoom(joinInput);
      if (joined) {
        setJoinInput(joined);
      }
    } finally {
      setAction(null);
    }
  };

  const handleLeaveRoom = async () => {
    if (!isInRoom) return;
    setAction("leave");
    try {
      await leaveRoom();
      setJoinInput("");
    } finally {
      setAction(null);
    }
  };

  const handleCopyRoomId = async () => {
    if (!roomId) return;
    try {
      await navigator.clipboard.writeText(roomId);
      toast.success("Room ID copied", { description: roomId });
    } catch (error) {
      console.error("Failed to copy room id", error);
      toast.error("Unable to copy room ID", {
        description: "Copy it manually instead",
      });
    }
  };

  const handleShareRoom = async () => {
    if (!roomId) return;
    const shareMessage = `Join my Seisami room: ${roomId}`;

    if (navigator.share) {
      try {
        await navigator.share({
          title: "Join my Seisami room",
          text: shareMessage,
        });
      } catch (error) {
        if ((error as Error).name !== "AbortError") {
          toast.error("Sharing failed", {
            description: (error as Error).message,
          });
        }
      }
      return;
    }

    await handleCopyRoomId();
    toast.info("Share this ID with your teammates", {
      description: shareMessage,
    });
  };

  return (
    <Card className="border-dashed">
      <CardHeader className="pb-0">
        <div className="flex items-start justify-between gap-3">
          <div>
            <CardTitle className="flex items-center gap-2 text-base">
              <Users className="h-4 w-4" />
              Live Collaboration
            </CardTitle>
            <CardDescription>
              Create or join a shared workspace to collaborate in real time.
            </CardDescription>
            {address && (
              <p className="mt-1 text-xs text-muted-foreground">
                Connected to server at{" "}
                <span className="font-mono">{address}</span>
              </p>
            )}
          </div>
          <Badge
            className={cn(
              "h-6 items-center justify-center px-2 text-xs font-medium",
              statusMeta.badgeClass
            )}
          >
            {statusMeta.label}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-4 pt-6">
        <div className="flex flex-wrap items-center gap-3">
          <Button
            onClick={handleCreateRoom}
            variant="default"
            disabled={isBusy || action === "join"}
          >
            <PlusCircle className="mr-2 h-4 w-4" />
            Create new room
          </Button>

          <div className="flex flex-1 min-w-[240px] items-center gap-2">
            <Input
              value={joinInput}
              onChange={(event) => setJoinInput(event.target.value)}
              placeholder="Enter room ID to join"
              disabled={status === "busy" && action === "create"}
            />
            <Button
              variant="secondary"
              onClick={handleJoinRoom}
              disabled={isBusy || !joinInput.trim()}
            >
              Join
            </Button>
          </div>
        </div>

        {isInRoom ? (
          <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border bg-muted/30 p-4">
            <div>
              <p className="text-sm font-semibold">Current room</p>
              <p className="font-mono text-sm text-muted-foreground">
                {roomId}
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Button variant="outline" size="sm" onClick={handleCopyRoomId}>
                <Copy className="mr-1 h-3 w-3" />
                Copy
              </Button>
              <Button variant="outline" size="sm" onClick={handleShareRoom}>
                <Share2 className="mr-1 h-3 w-3" />
                Share
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleLeaveRoom}
                disabled={isBusy && action !== "leave"}
                className="text-destructive hover:text-destructive"
              >
                <LogOut className="mr-1 h-3 w-3" />
                Leave
              </Button>
            </div>
          </div>
        ) : (
          <div className="rounded-xl border border-dashed bg-muted/10 p-4 text-sm text-muted-foreground">
            {status === "error" && lastError ? (
              <span className="text-destructive">{lastError}</span>
            ) : (
              <span>
                Create a new room or join one that a teammate shared with you.
              </span>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
};
