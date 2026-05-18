import { useTranslation } from "react-i18next";
import { NavLink } from "react-router-dom";
import { EopMark } from "@eop/ui";
import { Home, Plug, Settings as SettingsIcon, Shield, Users } from "lucide-react";
import type { ComponentType } from "react";

type NavItem = {
  to: string;
  label: string;
  icon: ComponentType<{ className?: string }>;
  badge?: string;
  badgeKind?: "new";
  matchExact?: boolean;
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
  onNavigate?: () => void;
}) {
  const { t } = useTranslation(["common", "app"]);

  const workspace: NavItem[] = [
    { to: "/dashboard", label: t("nav.dashboard"), icon: Home, matchExact: true },
    { to: "/team", label: t("nav.team"), icon: Users, matchExact: true },
  ];

  const insights: NavItem[] = [
    {
      to: "/integrations",
      label: t("sidebar.integrations"),
      icon: Plug,
      matchExact: true,
    },
    {
      to: "/settings",
      label: t("nav.settings"),
      icon: SettingsIcon,
      matchExact: true,
    },
  ];

  return (
    <aside className={`eop-sidebar ${open ? "open" : ""}`}>
      <div className="brand">
        <span className="brand-mark grid place-items-center text-accent">
          <EopMark size={26} />
        </span>
        <div className="brand-name">
          Eye of Providence
          <small>{t("sidebar.version_label")}</small>
        </div>
      </div>

      <div className="user-card">
        <div className="user-avatar">{user.avatarLabel}</div>
        <div className="user-info flex-1 min-w-0">
          <div className="user-name truncate">{user.name}</div>
          <div className="user-meta truncate">{user.handle}</div>
        </div>
        <span className="user-live" title={t("sidebar.tracking_active")} />
      </div>

      <div className="nav-group">
        <h6>{t("sidebar.workspace")}</h6>
        {workspace.map((it) => (
          <NavItemLink key={it.to + it.label} item={it} onNavigate={onNavigate} />
        ))}
      </div>

      <div className="nav-group">
        <h6>{t("sidebar.insights")}</h6>
        {insights.map((it) => (
          <NavItemLink key={it.to + it.label} item={it} onNavigate={onNavigate} />
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
          <span>{t("sidebar.local_first")}</span>
        </div>
      </div>
    </aside>
  );
}

function NavItemLink({ item, onNavigate }: { item: NavItem; onNavigate?: () => void }) {
  const Icon = item.icon;
  return (
    <NavLink
      to={item.to}
      onClick={onNavigate}
      end={item.matchExact}
      className={({ isActive }) => `nav-item ${isActive ? "active" : ""}`}
    >
      <Icon className="nav-icon" />
      <span className="truncate">{item.label}</span>
      {item.badge && <span className={`badge ${item.badgeKind ?? ""}`}>{item.badge}</span>}
    </NavLink>
  );
}
