// Webhooks management — Settings page section.
//
// Симметрия с api-tokens widget: list / create-form / show-secret modal.

import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button, Card, CardContent, CardDescription, CardHeader, CardTitle,
  Input, Modal, SimpleSelect, useConfirm,
} from "@eop/ui";
import { Copy, Plus, Trash2, Webhook as WebhookIcon } from "lucide-react";
import {
  useWebhooks, useCreateWebhook, useDeleteWebhook,
  type Webhook, type WebhookEvent, type WebhookFormat,
} from "../../entities/webhook";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";

const ALL_EVENTS: WebhookEvent[] = ["commit.ingested", "report.generated"];

export function WebhooksWidget() {
  const { t } = useTranslation("developer");
  const list = useWebhooks();
  const create = useCreateWebhook();
  const del = useDeleteWebhook();
  const confirm = useConfirm();
  const runToast = useMutationToast();

  const [showCreate, setShowCreate] = useState(false);
  const [showSecret, setShowSecret] = useState<string | null>(null);
  const [url, setUrl] = useState("");
  const [events, setEvents] = useState<WebhookEvent[]>(["commit.ingested"]);
  const [format, setFormat] = useState<WebhookFormat>("raw");

  function toggleEvent(e: WebhookEvent) {
    setEvents((prev) => (prev.includes(e) ? prev.filter((x) => x !== e) : [...prev, e]));
  }

  async function submitCreate() {
    if (!url.trim() || events.length === 0) return;
    const r = await runToast(create.mutateAsync({ url: url.trim(), events, format }), {});
    if (r) {
      setShowSecret(r.secret);
      setShowCreate(false);
      setUrl("");
      setEvents(["commit.ingested"]);
      setFormat("raw");
    }
  }

  async function doDelete(hook: Webhook) {
    const ok = await confirm({
      title: t("webhooks_delete_confirm"),
      description: t("webhooks_delete_confirm_lead"),
      destructive: true,
    });
    if (!ok) return;
    await runToast(del.mutateAsync(hook.id), {});
  }

  return (
    <Card>
      <CardHeader className="flex-col sm:flex-row sm:items-start sm:justify-between gap-3 sm:gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <WebhookIcon className="h-4 w-4 text-muted-foreground shrink-0" />
            <CardTitle>{t("webhooks_title")}</CardTitle>
          </div>
          <CardDescription className="mt-1">{t("webhooks_lead")}</CardDescription>
        </div>
        <Button size="sm" onClick={() => setShowCreate(true)} className="w-full sm:w-auto shrink-0">
          <Plus className="h-3.5 w-3.5 mr-1" />
          {t("webhooks_create")}
        </Button>
      </CardHeader>
      <CardContent>
        {list.isPending ? (
          <p className="text-sm text-muted-foreground">…</p>
        ) : !list.data || list.data.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("webhooks_empty")}</p>
        ) : (
          <ul className="divide-y">
            {list.data.map((hook) => <WebhookRow key={hook.id} hook={hook} onDelete={() => doDelete(hook)} />)}
          </ul>
        )}
      </CardContent>

      {/* Create */}
      <Modal open={showCreate} onClose={() => setShowCreate(false)} className="max-w-md">
        <div className="p-6 space-y-4">
          <h3 className="font-display text-lg">{t("webhooks_create")}</h3>
          <div className="space-y-3">
            <div>
              <label className="text-xs text-muted-foreground">{t("webhooks_url")}</label>
              <Input
                type="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder={t("webhooks_url_placeholder")}
              />
            </div>
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
                {format === "slack" ? t("webhooks_format_slack_hint") : t("webhooks_format_raw_hint")}
              </p>
            </div>
            <div>
              <label className="text-xs text-muted-foreground">{t("webhooks_events")}</label>
              <div className="space-y-2 pt-1">
                {ALL_EVENTS.map((e) => (
                  <label key={e} className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={events.includes(e)}
                      onChange={() => toggleEvent(e)}
                    />
                    {t(`webhooks_event_${e.replace(".", "_")}` as const)}
                  </label>
                ))}
              </div>
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="ghost" size="sm" onClick={() => setShowCreate(false)}>
              {t("tokens_close")}
            </Button>
            <Button
              size="sm"
              onClick={submitCreate}
              disabled={!url.trim() || events.length === 0 || create.isPending}
            >
              {t("webhooks_create")}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Show secret */}
      <Modal open={!!showSecret} onClose={() => setShowSecret(null)} className="max-w-lg">
        <div className="p-6 space-y-4">
          <h3 className="font-display text-lg">{t("webhooks_secret_title")}</h3>
          <p className="text-sm text-muted-foreground">{t("webhooks_secret_lead")}</p>
          {showSecret && <SecretField value={showSecret} />}
          <div className="flex justify-end pt-2">
            <Button size="sm" onClick={() => setShowSecret(null)}>{t("tokens_close")}</Button>
          </div>
        </div>
      </Modal>
    </Card>
  );
}

function WebhookRow({ hook, onDelete }: { hook: Webhook; onDelete: () => void }) {
  const { t } = useTranslation("developer");
  const last =
    hook.last_delivery_at && hook.last_status != null
      ? t("webhooks_last_delivery", {
          at: new Date(hook.last_delivery_at).toLocaleString(),
          status: hook.last_status === -1 ? t("webhooks_status_network") : hook.last_status,
        })
      : t("webhooks_no_deliveries");

  return (
    <li className="flex items-center justify-between py-3 gap-4">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs truncate">{hook.url}</span>
          {hook.format === "slack" && (
            <span className="text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded bg-purple-500/10 text-purple-600 dark:text-purple-400">
              slack
            </span>
          )}
        </div>
        <div className="flex items-center gap-2 mt-1 flex-wrap">
          {hook.events.map((e) => (
            <span key={e} className="text-[11px] uppercase tracking-wider px-1.5 py-0.5 rounded bg-muted">
              {e}
            </span>
          ))}
        </div>
        <div className="text-xs text-muted-foreground mt-1">{last}</div>
      </div>
      <Button variant="ghost" size="sm" onClick={onDelete}>
        <Trash2 className="h-3.5 w-3.5 mr-1" />
        {t("tokens_revoke")}
      </Button>
    </li>
  );
}

function SecretField({ value }: { value: string }) {
  const { t } = useTranslation("developer");
  const [copied, setCopied] = useState(false);
  function copy() {
    void navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }
  return (
    <div className="flex items-center gap-2">
      <input
        readOnly
        value={value}
        className="flex-1 rounded-md border bg-background px-3 py-2 text-xs font-mono"
      />
      <Button size="sm" variant="outline" onClick={copy}>
        <Copy className="h-3.5 w-3.5 mr-1" />
        {copied ? t("tokens_copied") : t("tokens_copy")}
      </Button>
    </div>
  );
}
