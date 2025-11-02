import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./ui/dialog";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Users, UserPlus, UserMinus, Mail, Crown, Loader2 } from "lucide-react";
import { ApiClient, BoardMember } from "~/lib/api-client";
import { useBoardStore } from "~/stores/board-store";
import { toast } from "sonner";
import { Avatar, AvatarFallback } from "./ui/avatar";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./ui/alert-dialog";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { format, parseISO } from "date-fns";

export const BoardMembersPanel = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [memberToRemove, setMemberToRemove] = useState<BoardMember | null>(
    null
  );
  const { currentBoard } = useBoardStore();
  const queryClient = useQueryClient();

  const {
    data: members,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["boardMembers", currentBoard?.id],
    queryFn: () => ApiClient.getBoardMembers(currentBoard!.id),
    enabled: !!currentBoard && isOpen,
  });

  const inviteMutation = useMutation({
    mutationFn: (email: string) =>
      ApiClient.inviteUserToBoard(email, currentBoard!.id),
    onSuccess: () => {
      toast.success(`Invitation sent to ${inviteEmail}`);
      setInviteEmail("");

      queryClient.invalidateQueries({
        queryKey: ["boardMembers", currentBoard?.id],
      });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to send invitation");
    },
  });

  const removeMutation = useMutation({
    mutationFn: (userId: string) =>
      ApiClient.removeUserFromBoard(userId, currentBoard!.id),
    onSuccess: () => {
      toast.success(`${memberToRemove?.email} removed from board`);
      setMemberToRemove(null);

      queryClient.invalidateQueries({
        queryKey: ["boardMembers", currentBoard?.id],
      });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to remove member");
    },
  });

  const handleInvite = async () => {
    if (!inviteEmail.trim() || !currentBoard) return;

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(inviteEmail)) {
      toast.error("Please enter a valid email address");
      return;
    }

    inviteMutation.mutate(inviteEmail);
  };

  const handleRemove = async () => {
    if (!memberToRemove || !currentBoard) return;

    removeMutation.mutate(memberToRemove.user_id);
  };

  const getInitials = (email: string) => {
    const name = email?.split("@")[0] ?? "";
    return name
      .split(/[._-]/)
      .map((part) => part[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const formatDate = (dateString: string) => {
    try {
      const date = parseISO(dateString);
      return format(date, "MMM d, yyyy");
    } catch (error) {
      console.error("Error formatting date:", error);
      return dateString;
    }
  };

  if (!currentBoard) return null;

  return (
    <>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start gap-2"
          >
            <Users className="h-4 w-4" />
            <span>Board Members</span>
            <span className="ml-auto text-xs text-muted-foreground">
              {members?.data.length}
            </span>
          </Button>
        </DialogTrigger>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Board Members
            </DialogTitle>
            <DialogDescription>
              Manage who has access to "{currentBoard.name}"
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6">
            {/* Invite Section */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Invite Member</label>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    type="email"
                    placeholder="Enter email address"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") {
                        handleInvite();
                      }
                    }}
                    className="pl-10"
                  />
                </div>
                <Button
                  onClick={handleInvite}
                  disabled={inviteMutation.isPending || !inviteEmail.trim()}
                >
                  {inviteMutation.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <UserPlus className="h-4 w-4" />
                  )}
                  <span>Invite</span>
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                Invited members will have access to all transcriptions and cards
                in this board
              </p>
            </div>

            {/* Members List */}
            <div className="space-y-2">
              <label className="text-sm font-medium">
                Current Members ({members?.data.length})
              </label>
              {isLoading ? (
                <div className="flex items-center justify-center p-8">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : isError ? (
                <div className="text-sm text-destructive p-4 text-center">
                  Failed to load members: {error?.message}
                </div>
              ) : (
                <div className="border rounded-lg divide-y max-h-[400px] overflow-y-auto">
                  {members?.data.length === 0 ? (
                    <div className="p-8 text-center text-sm text-muted-foreground">
                      No members yet. Invite someone to get started!
                    </div>
                  ) : (
                    members?.data.map((member) => (
                      <div
                        key={member.user_id}
                        className="flex items-center gap-3 p-3 hover:bg-muted/50 transition-colors"
                      >
                        <Avatar className="h-10 w-10">
                          <AvatarFallback className="bg-primary/10 text-primary text-sm font-medium">
                            {getInitials(member.email)}
                          </AvatarFallback>
                        </Avatar>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <p className="text-sm font-medium truncate">
                              {member.email}
                            </p>
                            {member.role === "owner" && (
                              <Crown className="h-3.5 w-3.5 text-yellow-600" />
                            )}
                          </div>
                          <p className="text-xs text-muted-foreground">
                            Joined {formatDate(member.joined_at)}
                          </p>
                        </div>
                        {member.role !== "owner" && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setMemberToRemove(member)}
                            className="text-destructive hover:text-destructive hover:bg-destructive/10"
                          >
                            <UserMinus className="h-4 w-4" />
                            <span>Remove</span>
                          </Button>
                        )}
                      </div>
                    ))
                  )}
                </div>
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Remove Confirmation Dialog */}
      <AlertDialog
        open={!!memberToRemove}
        onOpenChange={(open) => !open && setMemberToRemove(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Member</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove{" "}
              <span className="font-medium text-foreground">
                {memberToRemove?.email}
              </span>{" "}
              from this board? They will lose access to all transcriptions and
              cards.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={removeMutation.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRemove}
              disabled={removeMutation.isPending}
              className="bg-destructive hover:bg-destructive/90"
            >
              {removeMutation.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>Removing...</span>
                </>
              ) : (
                "Remove"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
};
