import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "~/components/ui/alert-dialog";
import { Checkbox } from "~/components/ui/checkbox";
import { Label } from "~/components/ui/label";
import { useState } from "react";

interface LogoutConfirmationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (clearLocalData: boolean) => void;
}

export const LogoutConfirmationDialog = ({
  open,
  onOpenChange,
  onConfirm,
}: LogoutConfirmationDialogProps) => {
  const [clearLocalData, setClearLocalData] = useState(false);

  const handleConfirm = () => {
    onConfirm(clearLocalData);
    setClearLocalData(false);
    onOpenChange(false);
  };

  const handleCancel = () => {
    setClearLocalData(false);
    onOpenChange(false);
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Confirm Logout</AlertDialogTitle>
          <AlertDialogDescription className="space-y-4">
            <span>Are you sure you want to log out?</span>
            <div className="flex items-start space-x-2 pt-2">
              <Checkbox
                id="clear-data"
                checked={clearLocalData}
                onCheckedChange={(checked: boolean) =>
                  setClearLocalData(checked === true)
                }
              />
              <div className="grid gap-1.5 leading-none">
                <Label
                  htmlFor="clear-data"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
                >
                  Clear all local data
                </Label>
                <p className="text-sm text-muted-foreground">
                  This will remove all boards, settings, and cached data from
                  this device. Your data will still be available in the cloud.
                </p>
              </div>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={handleCancel}>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm}>Logout</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};
