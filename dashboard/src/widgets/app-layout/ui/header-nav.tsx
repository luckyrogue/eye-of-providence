import { useTranslation } from "react-i18next";
import { NavLink } from "react-router-dom";
import { Activity, Eye, LogOut, Settings as SettingsIcon, Shield, Users } from "lucide-react";
import { NavItem } from "./nav-item";

export function HeaderNav({ isSuperAdmin, onLogout }: { isSuperAdmin: boolean; onLogout: () => void }) {
  const { t } = useTranslation("common");
  return (
    <header
      className="border-b header-blur sticky top-0 z-10"
      style={{ paddingTop: "env(safe-area-inset-top)" }}
    >
      <div className="mx-auto max-w-6xl px-4 md:px-6 py-3 md:py-4 flex items-center justify-between gap-2">
        <NavLink to="/dashboard" className="flex items-center gap-2 md:gap-3 group min-w-0">
          <div className="h-9 w-9 shrink-0 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center transition-transform duration-300 ease-out-expo group-hover:rotate-[8deg]">
            <Eye className="h-5 w-5 text-primary-foreground" />
          </div>
          <div className="min-w-0">
            <h1 className="font-display text-base md:text-lg font-bold tracking-tightest leading-none truncate">
              {t("app.name")}
            </h1>
            <p className="hidden sm:block text-[11px] uppercase tracking-widest2 text-muted-foreground mt-1 truncate">
              {t("app.tagline")}
            </p>
          </div>
        </NavLink>

        <nav className="hidden md:flex items-center gap-1 text-sm">
          <NavItem to="/dashboard" icon={<Activity className="h-4 w-4" />}>{t("nav.dashboard")}</NavItem>
          <NavItem to="/team" icon={<Users className="h-4 w-4" />}>{t("nav.team")}</NavItem>
          <NavItem to="/settings" icon={<SettingsIcon className="h-4 w-4" />}>{t("nav.settings")}</NavItem>
          {isSuperAdmin && (
            <NavItem to="/admin" icon={<Shield className="h-4 w-4" />} accent>
              {t("nav.admin")}
            </NavItem>
          )}
          <span className="mx-2 h-5 w-px bg-border" />
          <button
            onClick={onLogout}
            className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-muted-foreground hover:bg-secondary"
          >
            <LogOut className="h-4 w-4" /> {t("actions.logout")}
          </button>
        </nav>

        <button
          onClick={onLogout}
          aria-label={t("actions.logout")}
          className="md:hidden flex items-center justify-center h-11 w-11 rounded-md text-muted-foreground hover:bg-secondary active:bg-secondary/70 transition-colors"
        >
          <LogOut className="h-5 w-5" />
        </button>
      </div>
    </header>
  );
}
