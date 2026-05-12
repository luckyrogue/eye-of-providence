// Показываем ТОЛЬКО реальные routes. Раньше были 7 "tab"-пунктов вида
// /dashboard?tab=foo — NavLink матчит по pathname (игнорируя query) и
// подсвечивал все 7 одновременно; tab-routing внутри /dashboard не реализован.
// Когда появится — вернуть пункты с явной active-проверкой по location.search.

import { useTranslation } from "react-i18next";
import { NavLink, useLocation } from "react-router-dom";
import { Home, Plug, Settings as SettingsIcon, Shield, Users } from "lucide-react";
import type { ComponentType } from "react";

type NavItem = {
  to: string;
  label: string;
  icon: ComponentType<{ className?: string }>;
  badge?: string;
  badgeKind?: "new";
  // forceActive — кастомная проверка (для query-param tab'ов). Если undefined
  // → используется стандартная NavLink isActive (path match).
  forceActive?: boolean;
  // matchExact — NavLink end={true}. Нужно когда у двух пунктов общий префикс,
  // чтобы оба не подсвечивались (e.g. /settings vs /settings?tab=devices).
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
  const location = useLocation();

  const workspace: NavItem[] = [
    { to: "/dashboard", label: t("nav.dashboard"), icon: Home, matchExact: true },
    { to: "/team", label: t("nav.team"), icon: Users, matchExact: true },
  ];

  // /settings — два sidebar-пункта (Интеграции + Настройки) ведут в одну
  // страницу с tab'ом. NavLink по pathname матчит оба → нужна ручная
  // проверка location.search для разделения.
  const isIntegrations =
    location.pathname.startsWith("/settings") && location.search.includes("tab=devices");
  const isSettingsRoot =
    location.pathname.startsWith("/settings") && !location.search.includes("tab=devices");

  const insights: NavItem[] = [
    {
      to: "/settings?tab=devices",
      label: t("sidebar.integrations"),
      icon: Plug,
      forceActive: isIntegrations,
    },
    {
      to: "/settings",
      label: t("nav.settings"),
      icon: SettingsIcon,
      forceActive: isSettingsRoot,
    },
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

// forceActive: если задан — active-class приклеивается по нему вместо
// стандартного pathname-match'а NavLink.
function NavItemLink({ item, onNavigate }: { item: NavItem; onNavigate?: () => void }) {
  const Icon = item.icon;
  if (item.forceActive !== undefined) {
    return (
      <NavLink
        to={item.to}
        onClick={onNavigate}
        end={item.matchExact}
        className={`nav-item ${item.forceActive ? "active" : ""}`}
      >
        <Icon className="nav-icon" />
        <span className="truncate">{item.label}</span>
        {item.badge && <span className={`badge ${item.badgeKind ?? ""}`}>{item.badge}</span>}
      </NavLink>
    );
  }
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
