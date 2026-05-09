import { useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Input } from "@eop/ui";
import { ArrowRight, Copy, Loader2, Mail, SkipForward } from "lucide-react";

// InviteStep — два режима:
//   1) link-only (по умолчанию): сгенерить ссылку, скопировать руками
//   2) email-driven: ввести email, backend сам отправит письмо
//
// max_uses=10 для link-only, max_uses=1 для email-driven (см. handler).
export function InviteStep({
  busy,
  inviteUrl,
  inviteEmailSent,
  onGenerate,
  onCopy,
  onSkip,
  onContinue,
}: {
  busy: boolean;
  inviteUrl: string;
  inviteEmailSent: boolean;
  onGenerate: (email?: string) => void;
  onCopy: () => void;
  onSkip: () => void;
  onContinue: () => void;
}) {
  const [mode, setMode] = useState<"link" | "email">("link");
  const [email, setEmail] = useState("");
  const [copied, setCopied] = useState(false);

  function copy() {
    onCopy();
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Mail className="h-5 w-5" />
          Пригласите команду
        </CardTitle>
        <CardDescription>
          Вышлите приглашение на email или сгенерируйте ссылку, чтобы поделиться ей напрямую.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex gap-2 text-xs">
          <button
            type="button"
            onClick={() => setMode("link")}
            className={`px-3 py-1.5 rounded-md border transition-colors ${
              mode === "link" ? "bg-secondary" : "hover:bg-secondary/50"
            }`}
          >
            Ссылка (10 мест)
          </button>
          <button
            type="button"
            onClick={() => setMode("email")}
            className={`px-3 py-1.5 rounded-md border transition-colors ${
              mode === "email" ? "bg-secondary" : "hover:bg-secondary/50"
            }`}
          >
            Email (1 место)
          </button>
        </div>

        {mode === "email" ? (
          <>
            <Input
              type="email"
              placeholder="teammate@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={busy || inviteEmailSent}
            />
            {!inviteEmailSent ? (
              <Button
                onClick={() => onGenerate(email.trim())}
                disabled={busy || !email.includes("@")}
                className="w-full"
              >
                {busy ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
                Отправить приглашение
              </Button>
            ) : (
              <div className="rounded-md border bg-green-500/10 border-green-500/40 p-3 text-sm">
                Письмо отправлено. Если в течение 5 минут не пришло — проверь папку «Спам».
              </div>
            )}
          </>
        ) : !inviteUrl ? (
          <Button onClick={() => onGenerate(undefined)} disabled={busy} className="w-full">
            {busy ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
            Сгенерировать invite-ссылку
          </Button>
        ) : (
          <>
            <div className="flex items-center gap-2">
              <input
                readOnly
                value={inviteUrl}
                className="flex-1 rounded-md border bg-background px-3 py-2 text-xs font-mono"
              />
              <Button size="sm" variant="outline" onClick={copy}>
                <Copy className="h-3.5 w-3.5 mr-1" />
                {copied ? "Скопировано!" : "Скопировать"}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground font-mono">
              Действительна 7 дней. До 10 регистраций. Отозвать в Команда → Участники.
            </p>
          </>
        )}

        <div className="flex justify-between pt-2">
          <Button variant="ghost" onClick={onSkip} className="text-muted-foreground">
            <SkipForward className="h-3.5 w-3.5 mr-1.5" />
            Пропустить
          </Button>
          {(inviteUrl || inviteEmailSent) && (
            <Button onClick={onContinue}>
              Дальше <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
