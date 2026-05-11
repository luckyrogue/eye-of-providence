// AppShellV2 — новый layout с Sidebar + Topbar. Используется AppLayout
// (widgets/app-layout) если флаг включён, либо как явная альтернатива.

import { useState } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../../../entities/session";
import { useMe } from "../../../entities/user";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

export function AppShellV2() {
  const { logout } = useAuth();
  const me = useMe();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();
  const section = sectionLabel(location.pathname);

  const userName = me.data?.display_name || me.data?.email?.split("@")[0] || "User";
  const avatarLabel = userName.slice(0, 2).toUpperCase();

  return (
    <div className="grid grid-cols-1 md:grid-cols-[248px_1fr] min-h-screen">
      <Sidebar
        user={{
          name: userName,
          handle: me.data?.email ? "@" + me.data.email.split("@")[0] : "@you",
          avatarLabel,
        }}
        isSuperAdmin={me.data?.global_role === "super_admin"}
        open={sidebarOpen}
      />
      <main className="flex flex-col min-w-0">
        <Topbar
          crumb={{ section, now: "" }}
          onMenuClick={() => setSidebarOpen((s) => !s)}
          onLogout={() => {
            logout();
            window.location.href = "/login";
          }}
        />
        <div className="eop-dash-content mx-auto w-full">
          <Outlet />
        </div>
      </main>
    </div>
  );
}

function sectionLabel(path: string): string {
  if (path.startsWith("/dashboard")) return "Workspace";
  if (path.startsWith("/team")) return "Workspace";
  if (path.startsWith("/settings")) return "Settings";
  if (path.startsWith("/admin")) return "Admin";
  return "Workspace";
}
