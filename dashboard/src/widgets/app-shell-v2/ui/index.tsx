// AppShellV2 — новый layout с Sidebar + Topbar. Используется AppLayout
// (widgets/app-layout) если флаг включён, либо как явная альтернатива.

import { useState } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuth } from "../../../entities/session";
import { useMe } from "../../../entities/user";
import { useAuthRedirect } from "../../app-layout/lib/use-auth-redirect";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

export function AppShellV2() {
  const { t } = useTranslation("common");
  const { isAuthed, logout } = useAuth();
  const me = useMe();
  useAuthRedirect();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();
  const section = sectionLabel(location.pathname, t);

  if (!isAuthed) return null;

  const userName =
    me.data?.display_name || me.data?.email?.split("@")[0] || t("topbar.user_default");
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
        onNavigate={() => setSidebarOpen(false)}
      />
      {/* Backdrop overlay для mobile drawer — click outside закрывает sidebar.
          Скрыт на md+ где sidebar — sticky column а не off-canvas. */}
      {sidebarOpen && (
        // eslint-disable-next-line no-restricted-syntax -- backdrop overlay; IconButton не подходит (full-screen click target)
        <button
          type="button"
          aria-label={t("topbar.menu_toggle")}
          onClick={() => setSidebarOpen(false)}
          className="md:hidden fixed inset-0 z-[90] bg-black/60 backdrop-blur-sm"
        />
      )}
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

function sectionLabel(path: string, t: (k: string) => string): string {
  if (path.startsWith("/dashboard") || path.startsWith("/team"))
    return t("topbar.section_workspace");
  if (path.startsWith("/settings")) return t("topbar.section_settings");
  if (path.startsWith("/admin")) return t("topbar.section_admin");
  return t("topbar.section_workspace");
}
