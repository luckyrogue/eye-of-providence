import { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Auth } from "../Auth";
import { useAuth } from "../hooks/useAuth";

export function AuthRoute({ mode }: { mode: "login" | "register" }) {
  const { isAuthed, setAuth } = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();

  // Уже залогинен — редирект на dashboard.
  useEffect(() => {
    if (isAuthed) navigate("/dashboard", { replace: true });
  }, [isAuthed, navigate]);

  // Если в URL ?invite=CODE — Auth компонент сам подхватит, мы просто пропускаем mode-prop.
  // Существующий Auth ловит invite через URL params, нам остаётся только дать onAuth.
  void mode; // mode-аргумент — для будущего расщепления на отдельные routes login/signup.
  void params;

  function onAuth(r: { token: string; user_id: string; display_name?: string }) {
    setAuth(r);
    navigate("/dashboard", { replace: true });
  }

  return <Auth onAuth={onAuth} />;
}
