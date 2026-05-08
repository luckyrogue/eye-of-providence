import { useNavigate } from "react-router-dom";
import { Settings } from "./Settings";
import { useAuth } from "../../shared/hooks/useAuth";

export function SettingsRoute() {
  const { logout } = useAuth();
  const navigate = useNavigate();
  // Изменение tz — пока no-op, getTz читает локально каждый раз.
  return (
    <Settings
      onWiped={() => {
        logout();
        navigate("/login", { replace: true });
      }}
      onTzChange={() => { /* tz обновляется через query refetch при смене */ }}
    />
  );
}
