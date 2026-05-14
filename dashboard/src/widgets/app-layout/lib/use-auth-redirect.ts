import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../../entities/session";
import { useTeams } from "../../../entities/team";
import { useMe } from "../../../entities/user";
import { AUTH_FAILED_EVENT } from "../../../shared/api/http";
import { httpErrorStatus } from "../../../shared/config/query-client";
import { buildLoginURL } from "../../../shared/lib/redirect";
export function useAuthRedirect() {
  const { isAuthed, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const teams = useTeams();
  const me = useMe();
  useEffect(() => {
    function onAuthFailed() {
      const dest = location.pathname + location.search;
      logout();
      void navigate(buildLoginURL(dest), { replace: true });
    }
    window.addEventListener(AUTH_FAILED_EVENT, onAuthFailed);
    return () => window.removeEventListener(AUTH_FAILED_EVENT, onAuthFailed);
  }, [logout, navigate, location.pathname, location.search]);
  useEffect(() => {
    if (!me.isError || me.error == null) return;
    const status = httpErrorStatus(me.error);
    if (status === undefined) return;
    if (status === 401 || status === 403) return;
    if (status < 500 || status >= 600) return;
    const dest = location.pathname + location.search;
    logout();
    void navigate(buildLoginURL(dest), { replace: true });
  }, [me.isError, me.error, logout, navigate, location.pathname, location.search]);
  useEffect(() => {
    if (!isAuthed) {
      const dest = location.pathname + location.search;
      void navigate(buildLoginURL(dest), { replace: true });
    }
  }, [isAuthed, navigate, location.pathname, location.search]);
  useEffect(() => {
    if (!isAuthed) return;
    if (teams.isSuccess && teams.data.length === 0) {
      void navigate("/onboarding", { replace: true });
    }
  }, [isAuthed, teams.isSuccess, teams.data, navigate]);
  function doLogout() {
    logout();
    void navigate("/login", { replace: true });
  }
  return { doLogout };
}
