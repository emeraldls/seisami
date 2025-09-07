import { Outlet } from "react-router-dom";
import { Sidebar } from "~/components/sidebar";
import { useState } from "react";
import { useSidebar } from "~/contexts/sidebar-context";

export const AppLayout = () => {
  const { collapsed } = useSidebar();

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
