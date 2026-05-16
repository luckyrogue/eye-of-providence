import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Auth } from "./auth";
import { useAuth } from "../../entities/session";
import { readRedirect } from "../../shared/lib/redirect";

export function AuthRoute({ mode }: { mode: "login" | "register" }) {
  const { isAuthed, setAuth } = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();

  // postLoginDest: если в URL `?redirect=/path` (валидный) — туда; иначе
  // через /onboarding (он сам редиректнет на /dashboard если онбординг
  // уже dismissed/completed).
  const postLoginDest = readRedirect(params) ?? "/onboarding";

  // Уже залогинен — сразу туда же.
  useEffect(() => {
    if (isAuthed) void navigate(postLoginDest, { replace: true });
  }, [isAuthed, navigate, postLoginDest]);

  // Если в URL ?invite=CODE — Auth компонент сам подхватит, мы просто пропускаем mode-prop.
  // Существующий Auth ловит invite через URL params, нам остаётся только дать onAuth.
  void mode; // mode-аргумент — для будущего расщепления на отдельные routes login/signup.

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
