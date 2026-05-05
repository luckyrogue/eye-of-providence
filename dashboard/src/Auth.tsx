import { useEffect, useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Eye, Mail, User as UserIcon } from "lucide-react";
import { register, login, previewInvite, type AuthResponse, type InvitePreview } from "./api";

type Mode = "login" | "register";

export function Auth({ onAuth }: { onAuth: (r: AuthResponse) => void }) {
  const [mode, setMode] = useState<Mode>("register");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [invitePreview, setInvitePreviewData] = useState<InvitePreview | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
            <>
              Тебя пригласили в команду <strong>{invitePreview.team_name}</strong>.
              Создай аккаунт, чтобы присоединиться.
            </>
          ) : mode === "register" ? (
            "Создай аккаунт, чтобы отслеживать AI-активность"
          ) : (
            "Войди в свой аккаунт"
          )}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {error && <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-sm text-destructive">{error}</div>}

        {mode === "register" && (
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
          <label className="text-xs text-muted-foreground">Пароль (≥ 8 символов)</label>
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
