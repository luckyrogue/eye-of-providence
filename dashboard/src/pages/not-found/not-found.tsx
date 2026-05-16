import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { Button } from "@eop/ui";

// 404 страница — отдаётся для любого pathпо catch-all маршрута.
// Минимальная по UX-весу: один CTA на главную + ссылка на dashboard,
// если юзер залогинен (выбор делается через localStorage без хука,
// чтобы страница работала и без AuthProvider).
export function NotFound() {
  const { t } = useTranslation("common");
  const hasSession = typeof window !== "undefined" && !!localStorage.getItem("eop_token");

  return (
    <main className="min-h-screen flex items-center justify-center bg-background text-foreground px-6">
      <div className="max-w-md w-full text-center space-y-6">
        <div className="text-7xl font-semibold tracking-tighter text-muted-foreground">404</div>
        <div className="space-y-2">
          <h1 className="text-2xl font-medium">{t("not_found.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("not_found.lead")}</p>
        </div>
        <div className="flex items-center justify-center gap-2">
          <Link to="/">
            <Button variant="outline">{t("not_found.home")}</Button>
          </Link>
          {hasSession && (
            <Link to="/dashboard">
              <Button>{t("not_found.dashboard")}</Button>
            </Link>
          )}
        </div>
      </div>
    </main>
  );
}
