// import { formatDistanceToNow } from "date-fns";
import { ExternalLink, Info } from "lucide-react";
import { Notification } from "~/lib/api-client";
import { cn } from "~/lib/utils";
import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";
import { useNavigate } from "react-router-dom";

interface NotificationItemProps {
  notification: Notification;
  onClick: () => void;
}

export const NotificationItem = ({
  notification,
  onClick,
}: NotificationItemProps) => {
  const handleClick = () => {
    onClick();

    if (notification.type === "external_url" && notification.target) {
      BrowserOpenURL(notification.target);
    } else if (notification.type === "in_app" && notification.target) {
      console.log(notification.target);
      window.location.href = notification.target;
    }
    // For "info" type, do nothing
  };

  const getIcon = () => {
    switch (notification.type) {
      case "external_url":
        return <ExternalLink className="h-4 w-4" />;
      case "in_app":
        return <Info className="h-4 w-4" />;
      default:
        return <Info className="h-4 w-4" />;
    }
  };

  const isClickable = notification.type !== "info";

  return (
    <div
      className={cn(
        "p-4 transition-colors",
        !notification.read && "bg-muted/50",
        isClickable && "cursor-pointer hover:bg-muted/70"
      )}
      onClick={handleClick}
    >
      <div className="flex gap-3">
        <div className="flex-shrink-0 mt-1">{getIcon()}</div>
        <div className="flex-1 space-y-1">
          <div className="flex items-start justify-between gap-2">
            <p className="text-sm font-medium leading-none">
              {notification.title}
            </p>
            {!notification.read && (
              <div className="h-2 w-2 rounded-full bg-blue-500 flex-shrink-0 mt-1" />
            )}
          </div>
          <p className="text-sm text-muted-foreground">
            {notification.message}
          </p>
          <p className="text-xs text-muted-foreground">
            {/* {formatDistanceToNow(new Date(notification.created_at), {
              addSuffix: true,
            })} */}
          </p>
        </div>
      </div>
    </div>
  );
};
