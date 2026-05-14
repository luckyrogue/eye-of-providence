import { useEffect, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Eye, Lock } from "lucide-react";
import { fetchAuthConfig, type AuthConfig, type AuthResponse } from "../../entities/user";
import { LoginForm } from "../../features/auth-login";
import { RegisterForm } from "../../features/auth-register";
import { useInvitePreview } from "./model/use-invite-preview";
type Mode = "login" | "register";
export function Auth({ onAuth }: { onAuth: (r: AuthResponse) => void }) {
  const { t } = useTranslation("auth");
  const [mode, setMode] = useState<Mode>("login");
  const [authConfig, setAuthConfig] = useState<AuthConfig | null>(null);
  const { inviteCode, invitePreview } = useInvitePreview();
  useEffect(() => {
    void fetchAuthConfig().then((cfg) => {
      setAuthConfig(cfg);
      if (cfg.is_first_user) setMode("register");
    });
  }, []);
  useEffect(() => {
    if (inviteCode) setMode("register");
  }, [inviteCode]);
  const registrationBlocked = !!(
    authConfig?.invite_only &&
    !authConfig.is_first_user &&
    !inviteCode
  );
  return (
    <div className="relative px-4 sm:px-6">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-xl pt-8 sm:pt-12 pb-6 text-center reveal">
        <span className="eyebrow">Eye of Providence</span>
        <h1 className="display-head text-3xl sm:text-4xl md:text-5xl lg:text-6xl mt-3">
          <em>{mode === "register" ? t("title_register") : t("title_login")}.</em>
          <br />
          {t("subtitle")}
        </h1>
        <p className="mt-4 text-sm text-muted-foreground max-w-sm mx-auto">{t("lead")}</p>
      </div>
      <Card className="max-w-md mx-auto card-hover reveal reveal-delay-2">
        <CardHeader>
          <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
            <Eye className="h-6 w-6 text-primary-foreground" />
          </div>
          <CardTitle className="text-center font-display tracking-tight">
            {mode === "register" ? t("card_register") : t("card_login")}
          </CardTitle>
          <CardDescription className="text-center">
            {invitePreview ? (
              <Trans
                i18nKey="auth:desc_invited"
                values={{ team: invitePreview.team_name }}
                components={{ strong: <strong /> }}
              />
            ) : authConfig?.is_first_user ? (
              t("desc_first_user")
            ) : mode === "register" ? (
              t("desc_register")
            ) : (
              t("desc_login")
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {mode === "register" && registrationBlocked && (
            <div className="rounded-md border border-warning/40 bg-warning/10 p-3 text-sm">
              <div className="flex items-start gap-2">
                <Lock className="h-4 w-4 mt-0.5 text-warning shrink-0" />
                <div className="space-y-1">
                  <div className="font-medium text-warning">{t("invite_only_title")}</div>
                  <p className="text-xs text-muted-foreground">
                    <Trans
                      i18nKey="auth:invite_only_text"
                      components={{
                        code: <code className="mx-1 rounded bg-secondary px-1 text-[11px]" />,
                      }}
                    />
                  </p>
                </div>
              </div>
            </div>
          )}

          {!registrationBlocked && mode === "register" && (
            <RegisterForm inviteCode={inviteCode} onSuccess={onAuth} />
          )}
          {mode === "login" && <LoginForm onSuccess={onAuth} />}

          <Button
            type="button"
            variant="ghost"
            onClick={() => setMode(mode === "register" ? "login" : "register")}
            className="w-full text-xs text-muted-foreground hover:text-foreground"
          >
            {mode === "register" ? t("switch_to_login") : t("switch_to_register")}
          </Button>

          {mode === "login" && (
            <a
              href="/forgot-password"
              className="block w-full text-center text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              {t("forgot_password")}
            </a>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
