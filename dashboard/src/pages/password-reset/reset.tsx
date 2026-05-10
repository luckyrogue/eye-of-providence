import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { CheckCircle2, Lock } from "lucide-react";
import { ResetPasswordForm } from "../../features/auth-reset-password";

export function ResetPasswordRoute() {
  const { t } = useTranslation("auth");
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const token = params.get("token") ?? "";
  const [done, setDone] = useState(false);

  function onDone() {
    setDone(true);
    // Все старые сессии инвалидированы (token_version bumped) —
    // отправляем юзера логиниться заново.
    setTimeout(() => navigate("/login", { replace: true }), 1500);
  }

  if (!token) {
    return (
      <div className="mx-auto max-w-md px-4 pt-12 sm:pt-16">
        <Card>
          <CardHeader>
            <CardTitle className="text-center">{t("reset_invalid_link")}</CardTitle>
            <CardDescription className="text-center">{t("reset_invalid_link_lead")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Link to="/forgot-password" className="block text-center text-sm underline">
              {t("reset_request_new")}
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="relative">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-md px-4 pt-12 sm:pt-16">
        <Card className="card-hover reveal">
          <CardHeader>
            <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
              {done ? (
                <CheckCircle2 className="h-6 w-6 text-primary-foreground" />
              ) : (
                <Lock className="h-6 w-6 text-primary-foreground" />
              )}
            </div>
            <CardTitle className="text-center font-display tracking-tight">
              {done ? t("reset_done_title") : t("reset_title")}
            </CardTitle>
            <CardDescription className="text-center">
              {done ? t("reset_done_lead") : t("reset_lead")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {!done && <ResetPasswordForm token={token} onDone={onDone} />}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
