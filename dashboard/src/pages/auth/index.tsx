import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Auth } from "./auth";
import { useAuth } from "../../entities/session";
import { readRedirect } from "../../shared/lib/redirect";
export function AuthRoute({ mode }: { mode: "login" | "register" }) {
  const { isAuthed, setAuth } = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const postLoginDest = readRedirect(params) ?? "/onboarding";
  useEffect(() => {
    if (isAuthed) void navigate(postLoginDest, { replace: true });
  }, [isAuthed, navigate, postLoginDest]);
  void mode;
  function onAuth(r: { token: string; user_id: string; display_name?: string }) {
    setAuth(r);
    void navigate(postLoginDest, { replace: true });
  }
  return (
    <div className="min-h-screen flex items-center justify-center py-12 px-4">
      <div className="w-full">
        <Auth onAuth={onAuth} />
      </div>
    </div>
  );
}
