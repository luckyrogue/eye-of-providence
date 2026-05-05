import { useEffect, useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Eye, Lock, Mail, User as UserIcon } from "lucide-react";
import { register, login, previewInvite, fetchAuthConfig, type AuthResponse, type AuthConfig, type InvitePreview } from "./api";

type Mode = "login" | "register";

export function Auth({ onAuth }: { onAuth: (r: AuthResponse) => void }) {
  const [mode, setMode] = useState<Mode>("login");
  const [authConfig, setAuthConfig] = useState<AuthConfig | null>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [invitePreview, setInvitePreviewData] = useState<InvitePreview | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Конфиг auth (invite-only режим, первый ли пользователь)
  useEffect(() => {
    fetchAuthConfig().then((cfg) => {
      setAuthConfig(cfg);
      // Если invite-only нет и нет приглашения — показываем login по умолчанию.
      // Если первый user — показываем register (bootstrap).
      if (cfg.is_first_user) setMode("register");
    });
  }, []);

  // Поддержка ?invite=CODE в URL
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get("invite");
    if (code) {
      setInviteCode(code);
      setMode("register");
      previewInvite(code).then(setInvitePreviewData).catch(() => setError("Приглашение невалидно или устарело"));
    }
  }, []);

  async function submit() {
    setBusy(true);
    setError(null);
    try {
      const r = mode === "register"
        ? await register(email, password, displayName, inviteCode || undefined)
        : await login(email, password);
      onAuth(r);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  // Регистрация заблокирована? (invite-only, не первый user, нет invite кода)
  const registrationBlocked = !!(
    authConfig?.invite_only && !authConfig?.is_first_user && !inviteCode
  );

  return (
    <Card className="max-w-md mx-auto mt-12">
      <CardHeader>
        <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
          <Eye className="h-6 w-6 text-primary-foreground" />
        </div>
        <CardTitle className="text-center">
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
        {error && <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-sm text-destructive">{error}</div>}

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

        {mode === "register" && !registrationBlocked && (
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground flex items-center gap-1.5"><UserIcon className="h-3 w-3" />Имя</label>
            <input
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Темирлан"
              className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>
        )}

        {(!registrationBlocked || mode === "login") && (
          <>
            <div className="space-y-1">
              <label className="text-xs text-muted-foreground flex items-center gap-1.5"><Mail className="h-3 w-3" />Email</label>
              <input
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <div className="space-y-1">
              <label className="text-xs text-muted-foreground">Пароль{mode === "register" ? " (≥ 8 символов)" : ""}</label>
              <input
                type="password"
                autoComplete={mode === "register" ? "new-password" : "current-password"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && submit()}
                className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <Button onClick={submit} disabled={busy} className="w-full">
              {busy ? "..." : mode === "register" ? "Создать аккаунт" : "Войти"}
            </Button>
          </>
        )}

        <button
          onClick={() => setMode(mode === "register" ? "login" : "register")}
          className="w-full text-xs text-muted-foreground hover:text-foreground"
        >
          {mode === "register" ? "Уже есть аккаунт? Войти" : "Нет аккаунта? Зарегистрироваться"}
        </button>
      </CardContent>
    </Card>
  );
}
