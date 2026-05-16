import { useTranslation } from "react-i18next";
import { Link, useRouteError } from "react-router-dom";
import { Button } from "@eop/ui";

// Soft retry вместо `window.location.reload()`: повторный mount перезапускает
// loaders и react-query queries без полного перезагрузки страницы.
export function RouteError() {
  const error = useRouteError();
  const { t } = useTranslation("common");
  const detail = error instanceof Error ? error.message : String(error ?? "");

  return (
    <main className="min-h-screen flex items-center justify-center bg-background text-foreground px-6">
      <div className="max-w-md w-full text-center space-y-6">
        <div className="text-6xl font-semibold tracking-tighter text-muted-foreground">!</div>
        <div className="space-y-2">
          <h1 className="text-2xl font-medium">{t("route_error.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("route_error.lead")}</p>
          {detail && (
            <pre className="mt-3 whitespace-pre-wrap break-words rounded border bg-muted/30 px-3 py-2 text-left text-xs text-muted-foreground">
              {detail}
            </pre>
          )}
        </div>
        <div className="flex items-center justify-center gap-2">
          <Button variant="outline" onClick={() => window.history.back()}>
            {t("route_error.back")}
          </Button>
          <Link to="/">
            <Button>{t("route_error.home")}</Button>
          </Link>
        </div>
      </div>
    </main>
  );
}
