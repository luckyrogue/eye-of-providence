// Sidebar (artifact: Eye of Providence (1) — dashboard.jsx).
// Brand + user-card + 2 nav groups (Workspace / Insights) + privacy pill.

import { useTranslation } from "react-i18next";
import { NavLink } from "react-router-dom";
import {
  Activity,
  Brain,
  Code2,
  FolderOpen,
  Home,
  Languages,
  Plug,
  Settings as SettingsIcon,
  Shield,
  Users,
  FileText,
} from "lucide-react";
import type { ComponentType } from "react";

type NavItem = {
  to: string;
  label: string;
  icon: ComponentType<{ className?: string }>;
  badge?: string;
  badgeKind?: "new";
};

export function Sidebar({
  user,
  isSuperAdmin,
  open,
  onNavigate,
}: {
  user: { name: string; handle: string; avatarLabel: string };
  isSuperAdmin: boolean;
  open?: boolean;
  // onNavigate — закрываем mobile drawer после клика на nav-item, иначе
  // юзер должен сам закрыть drawer вручную. На desktop no-op.
  onNavigate?: () => void;
}) {
  const { t } = useTranslation("common");
  const workspace: NavItem[] = [
    { to: "/dashboard", label: t("nav.dashboard"), icon: Home },
    { to: "/dashboard?tab=activity", label: t("dashboard.stat_active"), icon: Activity },
    { to: "/dashboard?tab=ai", label: t("dashboard.stat_ai_share"), icon: Brain, badge: "live" },
    { to: "/dashboard?tab=provenance", label: "Code provenance", icon: Code2 },
    { to: "/dashboard?tab=projects", label: t("nav.projects") || "Projects", icon: FolderOpen },
    { to: "/dashboard?tab=languages", label: "Languages", icon: Languages },
  ];
  const insights: NavItem[] = [
    {
      to: "/dashboard?tab=reports",
      label: t("nav.reports") || "Reports",
      icon: FileText,
      badge: "new",
      badgeKind: "new",
    },
    { to: "/team", label: t("nav.team"), icon: Users },
    { to: "/settings?tab=devices", label: "Integrations", icon: Plug },
    { to: "/settings", label: t("nav.settings"), icon: SettingsIcon },
  ];

  return (
    <aside className={`eop-sidebar ${open ? "open" : ""}`}>
      <div className="brand">
        <span className="brand-mark grid place-items-center">
          <svg viewBox="0 0 28 28" width="26" height="26">
            <polygon
              points="14,3 26,24 2,24"
              fill="none"
              stroke="hsl(var(--accent))"
              strokeWidth="1.4"
            />
            <circle
              cx="14"
              cy="17"
              r="3"
              fill="none"
              stroke="hsl(var(--accent))"
              strokeWidth="1.4"
            />
            <circle cx="14" cy="17" r="1" fill="hsl(var(--accent))" />
          </svg>
        </span>
        <div className="brand-name">
          Eye of Providence
          <small>v0.4 · beta</small>
        </div>
      </div>

      <div className="user-card">
        <div className="user-avatar">{user.avatarLabel}</div>
        <div className="user-info flex-1 min-w-0">
          <div className="user-name truncate">{user.name}</div>
          <div className="user-meta truncate">{user.handle}</div>
        </div>
        <span className="user-live" title="Tracking active" />
      </div>

      <div className="nav-group">
        <h6>Workspace</h6>
        {workspace.map((it) => (
          <NavLink
            key={it.to + it.label}
            to={it.to}
            end={it.to === "/dashboard"}
            onClick={onNavigate}
            className={({ isActive }) => `nav-item ${isActive ? "active" : ""}`}
          >
            <it.icon className="nav-icon" />
            <span className="truncate">{it.label}</span>
            {it.badge && <span className={`badge ${it.badgeKind ?? ""}`}>{it.badge}</span>}
          </NavLink>
        ))}
      </div>

      <div className="nav-group">
        <h6>Insights</h6>
        {insights.map((it) => (
          <NavLink
            key={it.to + it.label}
            to={it.to}
            onClick={onNavigate}
            className={({ isActive }) => `nav-item ${isActive ? "active" : ""}`}
          >
            <it.icon className="nav-icon" />
            <span className="truncate">{it.label}</span>
            {it.badge && <span className={`badge ${it.badgeKind ?? ""}`}>{it.badge}</span>}
          </NavLink>
        ))}
        {isSuperAdmin && (
          <NavLink
            to="/admin"
            onClick={onNavigate}
            className={({ isActive }) => `nav-item ${isActive ? "active" : ""}`}
          >
            <Shield className="nav-icon" />
            <span>{t("nav.admin")}</span>
          </NavLink>
        )}
      </div>

      <div className="sidebar-footer">
        <div className="privacy-pill">
          <Shield style={{ width: 14, height: 14 }} />
          <span>Local-first · 0 sync pending</span>
        </div>
      </div>
    </aside>
  );
}
