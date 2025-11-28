import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./ui/alert-dialog";

export const DownloadAlert = ({
  onOpen,
  open,
  onClose,
}: {
  onOpen: () => void;
  open: boolean;
  onClose: () => void;
}) => {
  return (
    <AlertDialog open={open} onOpenChange={onOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Opening Seisami</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-2">
              <p>
                Because the app isn't signed yet, you'll need to follow these
                steps to open it:
              </p>
              <ol className="list-decimal list-inside space-y-2 ml-2">
                <li>
                  Open the app (you'll see a security warning saying it can't be
                  opened).
                </li>
                <li>
                  Go to <span className="font-bold">System Settings</span> &gt;{" "}
                  <span className="font-bold">Privacy & Security</span>.
                </li>
                <li>
                  Scroll down and click{" "}
                  <span className="font-bold">Open Anyway</span> for Seisami.
                </li>
              </ol>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogAction
            onClick={() => {
              onClose();
            }}
          >
            Done
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};
