import { useState } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation(["onboarding", "common"]);
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
          {t("invite.title")}
        </CardTitle>
        <CardDescription>{t("invite.lead")}</CardDescription>
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
            {t("invite.tab_link")}
          </button>
          <button
            type="button"
            onClick={() => setMode("email")}
            className={`px-3 py-1.5 rounded-md border transition-colors ${
              mode === "email" ? "bg-secondary" : "hover:bg-secondary/50"
            }`}
          >
            {t("invite.tab_email")}
          </button>
        </div>

        {mode === "email" ? (
          <>
            <Input
              type="email"
              placeholder={t("invite.email_placeholder")}
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
                {t("invite.send")}
              </Button>
            ) : (
              <div className="rounded-md border bg-green-500/10 border-green-500/40 p-3 text-sm">
                {t("invite.sent_ok")}
              </div>
            )}
          </>
        ) : !inviteUrl ? (
          <Button onClick={() => onGenerate(undefined)} disabled={busy} className="w-full">
            {busy ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
            {t("invite.generate")}
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
                {copied ? t("invite.copied") : t("invite.copy")}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground font-mono">{t("invite.url_note")}</p>
          </>
        )}

        <div className="flex justify-between pt-2">
          <Button variant="ghost" onClick={onSkip} className="text-muted-foreground">
            <SkipForward className="h-3.5 w-3.5 mr-1.5" />
            {t("common:actions.skip")}
          </Button>
          {(inviteUrl || inviteEmailSent) && (
            <Button onClick={onContinue}>
              {t("common:actions.continue")} <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
