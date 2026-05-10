import { useNavigate } from "react-router-dom";
import { Settings } from "./settings";
import { useAuth } from "../../entities/session";

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
    />
  );
}
