import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  SecretField,
  SimpleSelect,
} from "@eop/ui";
import {
  ALL_WEBHOOK_EVENTS,
  useCreateWebhook,
  type WebhookEvent,
  type WebhookFormat,
} from "@/entities/webhook";
import { useMutationToast } from "@/shared/hooks";

export function CreateWebhookDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useTranslation("developer");
  const create = useCreateWebhook();
  const runToast = useMutationToast();

  const [url, setUrl] = useState("");
  const [events, setEvents] = useState<WebhookEvent[]>(["commit.ingested"]);
  const [format, setFormat] = useState<WebhookFormat>("raw");
  const [secret, setSecret] = useState<string | null>(null);

  function toggleEvent(e: WebhookEvent) {
    setEvents((prev) => (prev.includes(e) ? prev.filter((x) => x !== e) : [...prev, e]));
  }

  function reset() {
    setUrl("");
    setEvents(["commit.ingested"]);
    setFormat("raw");
    setSecret(null);
  }

  function close() {
    reset();
    onClose();
  }

  async function submit() {
    if (!url.trim() || events.length === 0) return;
    const r = await runToast(create.mutateAsync({ url: url.trim(), events, format }), {});
    if (r) setSecret(r.secret);
  }

  return (
    <>
      <Dialog open={open && !secret} onOpenChange={(o) => !o && close()}>
        <DialogContent className="max-w-md p-6">
          <DialogHeader>
            <DialogTitle>{t("webhooks_create")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <Input
              label={t("webhooks_url")}
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder={t("webhooks_url_placeholder")}
            />
            <div>
              <label className="text-xs text-muted-foreground">{t("webhooks_format")}</label>
              <SimpleSelect
                value={format}
                onValueChange={(v) => setFormat(v as WebhookFormat)}
                triggerClassName="w-full"
                options={[
                  { value: "raw", label: t("webhooks_format_raw") },
                  { value: "slack", label: t("webhooks_format_slack") },
                ]}
              />
              <p className="mt-1 text-[11px] text-muted-foreground">
                {format === "slack"
                  ? t("webhooks_format_slack_hint")
                  : t("webhooks_format_raw_hint")}
              </p>
            </div>
            <div>
              <label className="text-xs text-muted-foreground">{t("webhooks_events")}</label>
              <div className="space-y-2 pt-1">
                {ALL_WEBHOOK_EVENTS.map((e) => (
                  <label key={e} className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.includes(e)}
                      onCheckedChange={() => toggleEvent(e)}
                      aria-label={t(`webhooks_event_${e.replace(".", "_")}` as const)}
                    />
                    {t(`webhooks_event_${e.replace(".", "_")}` as const)}
                  </label>
                ))}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" size="sm" onClick={close}>
              {t("tokens_close")}
            </Button>
            <Button
              size="sm"
              onClick={submit}
              disabled={!url.trim() || events.length === 0 || create.isPending}
            >
              {t("webhooks_create")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!secret} onOpenChange={(o) => !o && close()}>
        <DialogContent className="max-w-lg p-6">
          <DialogHeader>
            <DialogTitle>{t("webhooks_secret_title")}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">{t("webhooks_secret_lead")}</p>
          {secret && (
            <SecretField
              value={secret}
              copyLabel={t("tokens_copy")}
              copiedLabel={t("tokens_copied")}
            />
          )}
          <DialogFooter>
            <Button size="sm" onClick={close}>
              {t("tokens_close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
