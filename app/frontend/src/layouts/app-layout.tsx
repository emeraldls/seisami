import { Outlet } from "react-router-dom";
import { Sidebar } from "~/components/sidebar";
import { CloudSyncProgress } from "~/components/cloud-sync-progress";
import { useState, useEffect } from "react";
import { useSidebar } from "~/contexts/sidebar-context";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import { Info, MessageCircle } from "lucide-react";

export const AppLayout = () => {
  const { collapsed } = useSidebar();

  useEffect(() => {
    const unsubscribeProcessingStart = EventsOn(
      "ai:processing_start",
      (data: any) => {
        toast.info("🤖 AI is analyzing your request...", {
          description: "Processing your voice command",
          duration: 2000,
        });
      }
    );

    const unsubscribeToolComplete = EventsOn(
      "ai:tool_complete",
      (data: any) => {
        toast.success(`✅ Completed: ${data.toolName}`, {
          description:
            data.result?.length > 100
              ? `${data.result.substring(0, 100)}...`
              : data.result,
          duration: 2000,
        });
      }
    );

    const unsubscribeToolError = EventsOn("ai:tool_error", (data: any) => {
      toast.error(`❌ Failed: ${data.toolName}`, {
        description: data.error,
        duration: 3000,
      });
    });

    const unsubscribeProcessingComplete = EventsOn(
      "ai:processing_complete",
      (data: any) => {
        toast.success("🎉 AI finished processing!", {
          description: data.intent
            ? `Intent: ${data.intent}`
            : "Task completed successfully",
          duration: 3000,
        });
      }
    );

    const unsubscribeAiError = EventsOn("ai:error", (data: any) => {
      toast.error("🚨 AI Error", {
        description: data.error,
        duration: 4000,
      });
    });

    return () => {
      unsubscribeProcessingStart();
      unsubscribeToolComplete();
      unsubscribeToolError();
      unsubscribeProcessingComplete();
      unsubscribeAiError();
    };
  }, []);

  return (
    <>
      <div className="flex relative">
        <Sidebar />
        <div
          className="w-full transition-all duration-300 min-h-screen"
          style={{ paddingLeft: collapsed ? 72 : 256 }}
        >
          <Outlet />
        </div>

        <div className="fixed bottom-6 right-6 z-50">
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="rounded-full w-12 h-12 shadow-lg hover:shadow-xl transition-shadow"
              >
                <svg
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <circle cx="12" cy="12" r="10" />
                  <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
                  <path d="M12 17h.01" />
                </svg>
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-48 p-2" align="end" side="top">
              <div>
                <Button
                  variant="ghost"
                  className="w-full justify-start text-sm"
                >
                  <Info /> Help Center
                </Button>
                <Button
                  variant="ghost"
                  className="w-full justify-start text-sm"
                >
                  <MessageCircle /> Report a Bug
                </Button>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>
      <CloudSyncProgress />
    </>
  );
};
