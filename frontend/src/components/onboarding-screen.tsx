import { useState } from "react";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { useBoardStore } from "~/stores/board-store";
import { Label } from "~/components/ui/label";
import Logo from "~/assets/images/logo.png";

export function OnboardingScreen() {
  const [boardName, setBoardName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const { createBoard, isLoading } = useBoardStore();

  const handleCreateBoard = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!boardName.trim()) return;

    setIsCreating(true);
    const board = await createBoard(boardName.trim());

    if (board) {
      setBoardName("");
    }
    setIsCreating(false);
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-8">
        <Card className="shadow-xl border-0">
          <CardHeader className="text-center">
            <CardTitle className="text-2xl">Create Your First Board</CardTitle>
            <CardDescription className="">
              Start organizing your tasks and ideas with a personalized board
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreateBoard} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="boardName">Board Name</Label>
                <Input
                  id="boardName"
                  type="text"
                  placeholder="e.g., Personal Tasks, Work Projects..."
                  value={boardName}
                  onChange={(e) => setBoardName(e.target.value)}
                  className="w-full"
                  disabled={isCreating || isLoading}
                  autoFocus
                />
              </div>

              <Button
                type="submit"
                className="w-full"
                disabled={!boardName.trim() || isCreating || isLoading}
              >
                {isCreating || isLoading ? "Creating Board..." : "Create Board"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
