// useAuthRedirect — hooks для App-Layout:
//   - 401 → logout + redirect на /login c сохранением destination в ?redirect=
//   - !isAuthed → редирект на /login (с destination)
//   - isAuthed && teams.empty → редирект на /onboarding

import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../../entities/session";
import { useTeams } from "../../../entities/team";
import { AUTH_FAILED_EVENT } from "../../../shared/api/http";
import { buildLoginURL } from "../../../shared/lib/redirect";

export function useAuthRedirect() {
  const { isAuthed, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const teams = useTeams();

  useEffect(() => {
    function onAuthFailed() {
      const dest = location.pathname + location.search;
      logout();
      navigate(buildLoginURL(dest), { replace: true });
    }
    window.addEventListener(AUTH_FAILED_EVENT, onAuthFailed);
    return () => window.removeEventListener(AUTH_FAILED_EVENT, onAuthFailed);
  }, [logout, navigate, location.pathname, location.search]);

  useEffect(() => {
    if (!isAuthed) {
      const dest = location.pathname + location.search;
      navigate(buildLoginURL(dest), { replace: true });
    }
  }, [isAuthed, navigate, location.pathname, location.search]);

  // Без команд — на onboarding (но только когда query явно отдала [],
  // не во время initial load).
  useEffect(() => {
    if (!isAuthed) return;
    if (teams.isSuccess && teams.data.length === 0) {
      navigate("/onboarding", { replace: true });
    }
  }, [isAuthed, teams.isSuccess, teams.data, navigate]);

  function doLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  return { doLogout };
}
