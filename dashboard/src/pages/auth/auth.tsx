import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { Trans, useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Input } from "@eop/ui";
import { Eye, Lock } from "lucide-react";
import {
  register as apiRegister,
  login as apiLogin,
  fetchAuthConfig,
  type AuthResponse,
  type AuthConfig,
} from "../../entities/user";
import { previewInvite, type InvitePreview } from "../../entities/team";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";

type Mode = "login" | "register";

interface FormValues {
  email: string;
  password: string;
  displayName: string;
}

export function Auth({ onAuth }: { onAuth: (r: AuthResponse) => void }) {
  const { t } = useTranslation(["auth", "errors"]);
  const [mode, setMode] = useState<Mode>("login");
  const [authConfig, setAuthConfig] = useState<AuthConfig | null>(null);
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [invitePreview, setInvitePreview] = useState<InvitePreview | null>(null);
  const runToast = useMutationToast();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    defaultValues: { email: "", password: "", displayName: "" },
  });

  useEffect(() => {
    fetchAuthConfig().then((cfg) => {
      setAuthConfig(cfg);
      if (cfg.is_first_user) setMode("register");
    });
  }, []);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get("invite");
    if (code) {
      setInviteCode(code);
      setMode("register");
      previewInvite(code).then(setInvitePreview).catch(() => {
        // Тостим, состояние не блочим — юзер всё равно может попробовать.
      });
    }
  }, []);

  async function onSubmit(values: FormValues) {
    const promise =
      mode === "register"
        ? apiRegister(values.email, values.password, values.displayName, inviteCode || undefined)
        : apiLogin(values.email, values.password);
    const r = await runToast(promise, {
      error: mode === "register" ? t("errors:register_failed") : t("errors:auth_failed"),
    });
    if (r) onAuth(r);
  }

  const registrationBlocked = !!(
    authConfig?.invite_only && !authConfig.is_first_user && !inviteCode
  );

  return (
    <div className="relative">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-xl pt-12 pb-6 text-center reveal">
        <span className="eyebrow">Eye of Providence</span>
        <h1 className="display-head text-5xl md:text-6xl mt-3">
          <em>{mode === "register" ? t("auth:title_register") : t("auth:title_login")}.</em>
          <br />
          {t("auth:subtitle")}
        </h1>
        <p className="mt-4 text-sm text-muted-foreground max-w-sm mx-auto">{t("auth:lead")}</p>
      </div>
      <Card className="max-w-md mx-auto card-hover reveal reveal-delay-2">
        <CardHeader>
          <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
            <Eye className="h-6 w-6 text-primary-foreground" />
          </div>
          <CardTitle className="text-center font-display tracking-tight">
            {mode === "register" ? t("auth:card_register") : t("auth:card_login")}
          </CardTitle>
          <CardDescription className="text-center">
            {invitePreview ? (
              <Trans
                i18nKey="auth:desc_invited"
                values={{ team: invitePreview.team_name }}
                components={{ strong: <strong /> }}
              />
            ) : authConfig?.is_first_user ? (
              t("auth:desc_first_user")
            ) : mode === "register" ? (
              t("auth:desc_register")
            ) : (
              t("auth:desc_login")
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {mode === "register" && registrationBlocked && (
            <div className="rounded-md border border-amber-500/40 bg-amber-500/10 p-3 text-sm">
              <div className="flex items-start gap-2">
                <Lock className="h-4 w-4 mt-0.5 text-amber-600 dark:text-amber-400 shrink-0" />
                <div className="space-y-1">
                  <div className="font-medium text-amber-700 dark:text-amber-300">
                    {t("auth:invite_only_title")}
                  </div>
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

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            {mode === "register" && !registrationBlocked && (
              <Input
                label={t("auth:field_name")}
                placeholder={t("auth:field_name_placeholder")}
                autoComplete="name"
                error={errors.displayName?.message}
                {...register("displayName", {
                  required: t("auth:validation_name_required"),
                  maxLength: { value: 64, message: t("auth:validation_name_max") },
                })}
              />
            )}
            {(!registrationBlocked || mode === "login") && (
              <>
                <Input
                  label={t("auth:field_email")}
                  type="email"
                  autoComplete="email"
                  placeholder="you@example.com"
                  error={errors.email?.message}
                  {...register("email", {
                    required: t("auth:validation_email_required"),
                    pattern: {
                      value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                      message: t("auth:validation_email_invalid"),
                    },
                  })}
                />
                <Input
                  label={
                    mode === "register" ? t("auth:field_password_register") : t("auth:field_password")
                  }
                  type="password"
                  autoComplete={mode === "register" ? "new-password" : "current-password"}
                  error={errors.password?.message}
                  {...register("password", {
                    required: t("auth:validation_password_required"),
                    minLength: {
                      value: mode === "register" ? 8 : 1,
                      message: t("auth:validation_password_min"),
                    },
                  })}
                />
                <Button type="submit" disabled={isSubmitting} className="w-full">
                  {isSubmitting
                    ? "..."
                    : mode === "register"
                      ? t("auth:submit_register")
                      : t("auth:submit_login")}
                </Button>
              </>
            )}
          </form>

          <button
            onClick={() => setMode(mode === "register" ? "login" : "register")}
            className="w-full text-xs text-muted-foreground hover:text-foreground transition-colors"
          >
            {mode === "register" ? t("auth:switch_to_login") : t("auth:switch_to_register")}
          </button>

          {mode === "login" && (
            <a
              href="/forgot-password"
              className="block w-full text-center text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              {t("auth:forgot_password")}
            </a>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
