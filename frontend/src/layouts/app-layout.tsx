import { Outlet } from "react-router-dom";
import { Sidebar } from "~/components/sidebar";
import { useState, useEffect } from "react";
import { useSidebar } from "~/contexts/sidebar-context";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import { toast } from "sonner";

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
    <div className="flex">
      <Sidebar />
      <div
        className="w-full transition-all duration-300"
        style={{ paddingLeft: collapsed ? 72 : 256 }}
      >
        <Outlet />
      </div>
    </div>
  );
};
