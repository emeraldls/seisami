import { useEffect, useMemo } from "react";
import { Users, Copy, Share2 } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Button } from "~/components/ui/button";
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
  unauthenticated: {
    label: "Not Authenticated",
    badgeClass: "bg-purple-100 text-purple-700",
  },
};

/*
  Next step: update user local changes & broadcast to connected clients, if a clients just comes online, 
  im to pull changes for that room & check if the client has local changes, validate conflicts & merge both data, last write wins

  ------ Logic for when a card is updated on one user's end & on the other user's end is either dragging the same card -----
  - Update the current user local state, broadcast changes to connected clients, if a client is currently dragging an item or doing something, i need to track
  a client state to know if they're performing an action & if they're, we shouldnt update their board yet, until their mouse if off the mouse, then i'll pull changes


  - I can implmenent a custom cursor that shows in the other user's end where the current user's cursor is atm in the application, just like what figma does -- coool idea, for this, I will just get the cursor position & keep emitting a user cursor position to connected clients, then on connected clients end, i just draw the cursor & it overlays on their board. Cool
*/

/*
  ------- Collaboration with cloud logic -----

  When a user enables cloud features, first thing im to do is upload their local data to the cloud, 
  they'll know that this will happen. So now their local data & cloud is in sync
*/

export const CollaborationPanel = () => {
  const { status, roomId, address, lastError } = useCollaborationStore();

  const statusMeta = useMemo(() => statusConfig[status], [status]);
  const isInRoom = status === "in-room" && !!roomId;

  const handleCopyRoomId = async () => {
    if (!roomId) return;
    try {
      await navigator.clipboard.writeText(roomId);
      toast.success("Board ID copied", { description: roomId });
    } catch (error) {
      console.error("Failed to copy board id", error);
      toast.error("Unable to copy board ID", {
        description: "Copy it manually instead",
      });
    }
  };

  const handleShareRoom = async () => {
    if (!roomId) return;
    const shareMessage = `Collaborate on my Seisami board: ${roomId}`;

    if (navigator.share) {
      try {
        await navigator.share({
          title: "Join my Seisami board",
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

  return null;

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
              Collaborate with your team in real-time on this board.
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
        {isInRoom ? (
          <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border bg-muted/30 p-4">
            <div>
              <p className="text-sm font-semibold">Collaborating on board</p>
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
            </div>
          </div>
        ) : (
          <div className="rounded-xl border border-dashed bg-muted/10 p-4 text-sm text-muted-foreground">
            {status === "error" && lastError ? (
              <span className="text-destructive">{lastError}</span>
            ) : (
              <span>
                Real-time collaboration is ready. Changes will be synced
                automatically.
              </span>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
};
