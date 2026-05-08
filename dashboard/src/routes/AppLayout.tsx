import { useEffect } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { Activity, Eye, LogOut, Settings as SettingsIcon, Shield, Users } from "lucide-react";
import { cn } from "@eop/ui";
import { useAuth } from "../hooks/useAuth";
import { useMe, useTeams } from "../hooks/queries";
import { AUTH_FAILED_EVENT } from "../api";

export function AppLayout() {
  const { isAuthed, logout } = useAuth();
  const navigate = useNavigate();
  const me = useMe();
  const teams = useTeams();

  // 401 → logout + редирект на /login
  useEffect(() => {
    function onAuthFailed() {
      logout();
      navigate("/login", { replace: true });
    }
    window.addEventListener(AUTH_FAILED_EVENT, onAuthFailed);
    return () => window.removeEventListener(AUTH_FAILED_EVENT, onAuthFailed);
  }, [logout, navigate]);

  // Не аутентифицирован — отправляем на login.
  useEffect(() => {
    if (!isAuthed) navigate("/login", { replace: true });
  }, [isAuthed, navigate]);

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
              <h1 className="font-display text-lg font-bold tracking-tightest leading-none">Eye of Providence</h1>
              <p className="text-[11px] uppercase tracking-widest2 text-muted-foreground mt-1">Manual vs AI · live</p>
            </div>
          </NavLink>

          <nav className="flex items-center gap-1 text-sm">
            <NavItem to="/dashboard" icon={<Activity className="h-4 w-4" />}>Дашборд</NavItem>
            <NavItem to="/team" icon={<Users className="h-4 w-4" />}>Команда</NavItem>
            <NavItem to="/settings" icon={<SettingsIcon className="h-4 w-4" />}>Настройки</NavItem>
            {isSuperAdmin && (
              <NavItem to="/admin" icon={<Shield className="h-4 w-4" />} accent>
                Admin
              </NavItem>
            )}
            <span className="mx-2 h-5 w-px bg-border" />
            <button
              onClick={doLogout}
              className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-muted-foreground hover:bg-secondary"
            >
              <LogOut className="h-4 w-4" /> Выход
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
