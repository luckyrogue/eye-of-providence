import { Navigate, useNavigate, useSearchParams } from "react-router-dom";
import { Settings } from "./settings";
import { useAuth } from "../../entities/session";

export function SettingsRoute() {
  const [search] = useSearchParams();
  const { logout } = useAuth();
  const navigate = useNavigate();

  if (search.get("tab") === "devices") {
    return <Navigate to="/integrations" replace />;
  }

  return (
    <Settings
      onWiped={() => {
        void logout();
        void navigate("/login", { replace: true });
      }}
    />
  );
}
