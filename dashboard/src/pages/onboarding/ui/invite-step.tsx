import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Input,
  SecretField,
  Tabs,
  TabsList,
  TabsTrigger,
} from "@eop/ui";
import { ArrowRight, Loader2, Mail, SkipForward } from "lucide-react";

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
        <Tabs value={mode} onValueChange={(v) => setMode(v as "link" | "email")}>
          <TabsList>
            <TabsTrigger value="link">{t("invite.tab_link")}</TabsTrigger>
            <TabsTrigger value="email">{t("invite.tab_email")}</TabsTrigger>
          </TabsList>
        </Tabs>

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
            <SecretField
              value={inviteUrl}
              copyLabel={t("invite.copy")}
              copiedLabel={t("invite.copied")}
              onCopy={onCopy}
            />
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
