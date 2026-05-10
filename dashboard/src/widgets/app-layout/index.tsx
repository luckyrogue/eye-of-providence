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

  // Bottom-nav на mobile (PWA-style) + top-nav на desktop. Содержимое
  // заворачиваем в pb-16 чтобы bottom-nav не накрывал последний row.
  return (
    <main
      className="min-h-screen bg-gradient-to-br from-background via-background to-primary/5"
      style={{
        paddingBottom: "calc(4rem + env(safe-area-inset-bottom))",
      }}
    >
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
              <h1 className="font-display text-base md:text-lg font-bold tracking-tightest leading-none truncate">{t("app.name")}</h1>
              <p className="hidden sm:block text-[11px] uppercase tracking-widest2 text-muted-foreground mt-1 truncate">
                {t("app.tagline")}
              </p>
            </div>
          </NavLink>

          {/* Desktop nav: full set with labels */}
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
              onClick={doLogout}
              className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-muted-foreground hover:bg-secondary"
            >
              <LogOut className="h-4 w-4" /> {t("actions.logout")}
            </button>
          </nav>

          {/* Mobile: только logout-кнопка в header (44×44 tap-target) */}
          <button
            onClick={doLogout}
            aria-label={t("actions.logout")}
            className="md:hidden flex items-center justify-center h-11 w-11 rounded-md text-muted-foreground hover:bg-secondary active:bg-secondary/70 transition-colors"
          >
            <LogOut className="h-5 w-5" />
          </button>
        </div>
      </header>

      <div className="mx-auto max-w-6xl px-4 md:px-6 py-4 md:py-8 space-y-4 md:space-y-6">
        <Outlet />
      </div>

      {/* Mobile bottom-nav — fixed, native-app pattern. Hidden на desktop. */}
      <nav
        className="md:hidden fixed bottom-0 inset-x-0 border-t bg-background/95 backdrop-blur-md z-20"
        style={{ paddingBottom: "env(safe-area-inset-bottom)" }}
        aria-label={t("nav.bottom")}
      >
        <div className="flex items-stretch">
          <BottomNavItem to="/dashboard" icon={<Activity className="h-5 w-5" />} label={t("nav.dashboard")} />
          <BottomNavItem to="/team" icon={<Users className="h-5 w-5" />} label={t("nav.team")} />
          <BottomNavItem to="/settings" icon={<SettingsIcon className="h-5 w-5" />} label={t("nav.settings")} />
          {isSuperAdmin && (
            <BottomNavItem to="/admin" icon={<Shield className="h-5 w-5" />} label={t("nav.admin")} accent />
          )}
        </div>
      </nav>
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

// BottomNavItem — mobile-only tab. Min-height 56px (Material Design /
// iOS-стиль), tap-target ≥ 44px. Active state — primary-color подсветка
// + top accent bar для visual feedback.
function BottomNavItem({
  to,
  icon,
  label,
  accent,
}: {
  to: string;
  icon: React.ReactNode;
  label: string;
  accent?: boolean;
}) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        cn(
          "relative flex-1 flex flex-col items-center justify-center gap-0.5 py-2 min-h-[56px] transition-colors",
          isActive
            ? accent
              ? "text-amber-600 dark:text-amber-400"
              : "text-primary"
            : "text-muted-foreground hover:text-foreground active:bg-secondary/30",
        )
      }
    >
      {({ isActive }) => (
        <>
          {/* Top accent bar — заметно показывает active state */}
          <span
            className={cn(
              "absolute top-0 left-1/2 -translate-x-1/2 h-0.5 w-8 rounded-full transition-all",
              isActive
                ? accent
                  ? "bg-amber-500"
                  : "bg-primary"
                : "bg-transparent",
            )}
          />
          {icon}
          <span className="text-[10px] font-medium leading-none">{label}</span>
        </>
      )}
    </NavLink>
  );
}
