import { useEffect } from "react";
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Activity, Eye, LogOut, Settings as SettingsIcon, Shield, Users } from "lucide-react";
import { cn } from "@eop/ui";
import { useAuth } from "../../shared/hooks/use-auth";
import { useMe } from "../../entities/user";
import { useTeams } from "../../entities/team";
import { AUTH_FAILED_EVENT } from "../../shared/api/http";
import { buildLoginURL } from "../../shared/lib/redirect";

export function AppLayout() {
  const { t } = useTranslation("common");
  const { isAuthed, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const me = useMe();
  const teams = useTeams();

  // 401 → logout + редирект на /login с сохранённым ?redirect=, чтобы
  // после re-login юзер вернулся туда же.
  useEffect(() => {
    function onAuthFailed() {
      const dest = location.pathname + location.search;
      logout();
      navigate(buildLoginURL(dest), { replace: true });
    }
    window.addEventListener(AUTH_FAILED_EVENT, onAuthFailed);
    return () => window.removeEventListener(AUTH_FAILED_EVENT, onAuthFailed);
  }, [logout, navigate, location.pathname, location.search]);

  // Не аутентифицирован — отправляем на login c сохранением destination.
  useEffect(() => {
    if (!isAuthed) {
      const dest = location.pathname + location.search;
      navigate(buildLoginURL(dest), { replace: true });
    }
  }, [isAuthed, navigate, location.pathname, location.search]);

  // Без команд — на onboarding (но только если не super_admin без команд:
  // он мог только что разлогиниться-залогиниться и попасть сюда).
  useEffect(() => {
    if (!isAuthed) return;
    if (teams.isSuccess && teams.data.length === 0) {
      navigate("/onboarding", { replace: true });
    }
  }, [isAuthed, teams.isSuccess, teams.data, navigate]);

  const isSuperAdmin = me.data?.global_role === "super_admin";

  function doLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  return (
    <main className="min-h-screen bg-gradient-to-br from-background via-background to-primary/5">
      <div className="border-b header-blur sticky top-0 z-10">
        <div className="mx-auto max-w-6xl px-6 py-4 flex items-center justify-between">
          <NavLink to="/dashboard" className="flex items-center gap-3 group">
            <div className="h-9 w-9 rounded-lg bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center transition-transform duration-300 ease-out-expo group-hover:rotate-[8deg]">
              <Eye className="h-5 w-5 text-primary-foreground" />
            </div>
            <div>
              <h1 className="font-display text-lg font-bold tracking-tightest leading-none">{t("app.name")}</h1>
              <p className="text-[11px] uppercase tracking-widest2 text-muted-foreground mt-1">{t("app.tagline")}</p>
            </div>
          </NavLink>

          <nav className="flex items-center gap-1 text-sm">
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
              onClick={doLogout}
              className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-muted-foreground hover:bg-secondary"
            >
              <LogOut className="h-4 w-4" /> {t("actions.logout")}
            </button>
          </nav>
        </div>
      </div>

      <div className="mx-auto max-w-6xl px-6 py-8 space-y-6">
        <Outlet />
      </div>
    </main>
  );
}

function NavItem({
  to,
  icon,
  children,
  accent,
}: {
  to: string;
  icon: React.ReactNode;
  children: React.ReactNode;
  accent?: boolean;
}) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        cn(
          "flex items-center gap-1.5 rounded-md px-3 py-1.5 transition-colors",
          isActive
            ? "bg-primary text-primary-foreground"
            : accent
              ? "text-amber-600 dark:text-amber-400 hover:bg-amber-500/10"
              : "text-muted-foreground hover:bg-secondary",
        )
      }
    >
      {icon} {children}
    </NavLink>
  );
}
