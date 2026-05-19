// /auth/complete — landing route after OAuth/passkey backend redirect.
//
// Backend issues a one-shot HttpOnly cookie `eop_session_handoff` and bounces
// the browser here with `?return_to=/somewhere`. We read the cookie, save the
// JWT to localStorage, prime react-query, and navigate to `return_to` (or
// `/dashboard`). Errors are rendered inline (cookie expired or never set →
// 401) with a back-to-login CTA.
import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { AlertTriangle, Loader2 } from "lucide-react";
import { completeSessionHandoff, fetchMe } from "@/entities/user";
import { useAuth } from "@/entities/session";
import { safeRedirect } from "@/shared/lib/redirect";

type Phase = "loading" | "error";

export function AuthCompleteRoute() {
  const { t } = useTranslation("auth");
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const { setAuth, logout } = useAuth();
  const qc = useQueryClient();
  const [phase, setPhase] = useState<Phase>("loading");

  useEffect(() => {
    let cancelled = false;
    const returnTo = safeRedirect(params.get("return_to")) ?? "/dashboard";

    void (async () => {
      try {
        const token = await completeSessionHandoff();
        if (cancelled) return;
        // Token is in localStorage; load /me so we can hydrate useAuth's
        // user_id (route guards depend on it) before navigating away.
        const me = await fetchMe();
        if (cancelled) return;
        setAuth({ user_id: me.user_id, token, display_name: me.display_name });
        await qc.invalidateQueries();
        if (cancelled) return;
        void navigate(returnTo, { replace: true });
      } catch {
        if (!cancelled) {
          logout();
          setPhase("error");
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [navigate, params, qc, setAuth, logout]);

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-12">
      <Card className="w-full max-w-md card-hover">
        <CardHeader>
          <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
            {phase === "loading" ? (
              <Loader2 className="h-6 w-6 text-primary-foreground animate-spin" aria-hidden />
            ) : (
              <AlertTriangle className="h-6 w-6 text-primary-foreground" aria-hidden />
            )}
          </div>
          <CardTitle className="text-center font-display tracking-tight">
            {phase === "loading" ? t("complete.title") : t("complete.error_title")}
          </CardTitle>
          <CardDescription className="text-center">
            {phase === "loading" ? t("complete.lead") : t("complete.error_lead")}
          </CardDescription>
        </CardHeader>
        {phase === "error" && (
          <CardContent>
            <Button asChild className="w-full">
              <Link to="/login">{t("complete.back_to_login")}</Link>
            </Button>
          </CardContent>
        )}
      </Card>
    </div>
  );
}
