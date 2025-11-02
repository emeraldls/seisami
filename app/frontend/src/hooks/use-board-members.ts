import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ApiClient } from "~/lib/api-client";
import { toast } from "sonner";

export const useBoardMembers = (boardId?: string, enabled: boolean = true) => {
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: ["boardMembers", boardId],
    queryFn: () => ApiClient.getBoardMembers(boardId!),
    enabled: !!boardId && enabled,
  });

  const inviteMutation = useMutation({
    mutationFn: ({ email, boardId }: { email: string; boardId: string }) =>
      ApiClient.inviteUserToBoard(email, boardId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["boardMembers", variables.boardId],
      });
    },
  });

  const removeMutation = useMutation({
    mutationFn: ({ userId, boardId }: { userId: string; boardId: string }) =>
      ApiClient.removeUserFromBoard(userId, boardId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["boardMembers", variables.boardId],
      });
    },
  });

  return {
    members: query.data ?? [],
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    inviteMember: inviteMutation.mutate,
    removeMember: removeMutation.mutate,
    isInviting: inviteMutation.isPending,
    isRemoving: removeMutation.isPending,
  };
};
