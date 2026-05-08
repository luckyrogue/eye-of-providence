import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
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
      error: mode === "register" ? "Регистрация не удалась" : "Не удалось войти",
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
          <em>{mode === "register" ? "Создай" : "Войди"}.</em>
          <br />
          Отслеживай.
        </h1>
        <p className="mt-4 text-sm text-muted-foreground max-w-sm mx-auto">
          Сколько ты пишешь сам, а сколько — с AI. Privacy-by-design.
        </p>
      </div>
      <Card className="max-w-md mx-auto card-hover reveal reveal-delay-2">
        <CardHeader>
          <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
            <Eye className="h-6 w-6 text-primary-foreground" />
          </div>
          <CardTitle className="text-center font-display tracking-tight">
            {mode === "register" ? "Регистрация" : "Вход"}
          </CardTitle>
          <CardDescription className="text-center">
            {invitePreview ? (
              <>Тебя пригласили в команду <strong>{invitePreview.team_name}</strong>. Создай аккаунт, чтобы присоединиться.</>
            ) : authConfig?.is_first_user ? (
              "Ты первый пользователь — станешь super_admin."
            ) : mode === "register" ? (
              "Создай аккаунт, чтобы отслеживать AI-активность"
            ) : (
              "Войди в свой аккаунт"
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {mode === "register" && registrationBlocked && (
            <div className="rounded-md border border-amber-500/40 bg-amber-500/10 p-3 text-sm">
              <div className="flex items-start gap-2">
                <Lock className="h-4 w-4 mt-0.5 text-amber-600 dark:text-amber-400 shrink-0" />
                <div className="space-y-1">
                  <div className="font-medium text-amber-700 dark:text-amber-300">Регистрация по приглашению</div>
                  <p className="text-xs text-muted-foreground">
                    Открытая регистрация выключена. Попроси у участника команды invite-ссылку вида
                    <code className="mx-1 rounded bg-secondary px-1 text-[11px]">?invite=...</code>.
                  </p>
                </div>
              </div>
            </div>
          )}

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            {mode === "register" && !registrationBlocked && (
              <Input
                label="Имя"
                placeholder="Темирлан"
                autoComplete="name"
                error={errors.displayName?.message}
                {...register("displayName", {
                  required: "имя обязательно",
                  maxLength: { value: 64, message: "максимум 64 символа" },
                })}
              />
            )}
            {(!registrationBlocked || mode === "login") && (
              <>
                <Input
                  label="Email"
                  type="email"
                  autoComplete="email"
                  placeholder="you@example.com"
                  error={errors.email?.message}
                  {...register("email", {
                    required: "email обязателен",
                    pattern: { value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: "невалидный email" },
                  })}
                />
                <Input
                  label={mode === "register" ? "Пароль (≥ 8 символов)" : "Пароль"}
                  type="password"
                  autoComplete={mode === "register" ? "new-password" : "current-password"}
                  error={errors.password?.message}
                  {...register("password", {
                    required: "пароль обязателен",
                    minLength: { value: mode === "register" ? 8 : 1, message: "минимум 8 символов" },
                  })}
                />
                <Button type="submit" disabled={isSubmitting} className="w-full">
                  {isSubmitting ? "..." : mode === "register" ? "Создать аккаунт" : "Войти"}
                </Button>
              </>
            )}
          </form>

          <button
            onClick={() => setMode(mode === "register" ? "login" : "register")}
            className="w-full text-xs text-muted-foreground hover:text-foreground transition-colors"
          >
            {mode === "register" ? "Уже есть аккаунт? Войти" : "Нет аккаунта? Зарегистрироваться"}
          </button>
        </CardContent>
      </Card>
    </div>
  );
}
