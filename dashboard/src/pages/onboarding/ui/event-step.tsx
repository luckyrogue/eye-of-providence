import { useEffect } from "react";
import { Trans, useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { ArrowRight, CheckCircle2, Loader2, Radio, SkipForward } from "lucide-react";
import { useOnboardingStatus } from "../../../entities/user";

// Polling каждые 5s до первого события от агента; при has_event=true —
// auto-advance в onFinish (короткая задержка задаётся в эффекте ниже).
export function EventStep({ onFinish, onSkip }: { onFinish: () => void; onSkip: () => void }) {
  const { t } = useTranslation("onboarding");
  const status = useOnboardingStatus({ refetchInterval: 5000 });
  const hasEvent = status.data?.has_event ?? false;

  // Auto-advance: как только событие пришло, через 1.5s закрываем wizard.
  useEffect(() => {
    if (hasEvent) {
      const timer = setTimeout(onFinish, 1500);
      return () => clearTimeout(timer);
    }
  }, [hasEvent, onFinish]);

  if (hasEvent) {
    return (
      <Card className="card-hover reveal">
        <CardHeader>
          <CardTitle className="font-display tracking-tight flex items-center gap-2">
            <CheckCircle2 className="h-5 w-5 text-green-500" />
            {t("event.title_received")}
          </CardTitle>
          <CardDescription>{t("event.lead_received")}</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Radio className="h-5 w-5 text-primary animate-pulse" />
          {t("event.title_waiting")}
        </CardTitle>
        <CardDescription>{t("event.lead_waiting")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-3 rounded-md border bg-muted/30 p-3">
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          <span className="text-sm text-muted-foreground font-mono">{t("event.polling_note")}</span>
        </div>
        <div className="rounded-md border bg-card p-3 text-xs space-y-2">
          <p className="font-medium">{t("event.troubleshoot_title")}</p>
          <ul className="list-disc pl-4 space-y-1 text-muted-foreground">
            <li>{t("event.ts_tauri")}</li>
            <li>
              <Trans
                i18nKey="onboarding:event.ts_vscode"
                components={{ code: <code className="font-mono" /> }}
              />
            </li>
            <li>{t("event.ts_browser")}</li>
          </ul>
        </div>
        <div className="flex justify-end pt-2">
          <Button variant="ghost" onClick={onSkip} className="text-muted-foreground">
            <SkipForward className="h-3.5 w-3.5 mr-1.5" />
            {t("event.open_empty")}
            <ArrowRight className="h-4 w-4 ml-2" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
