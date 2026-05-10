// Outbound webhook subscriptions.

export type WebhookEvent = "commit.ingested" | "report.generated";

// raw  — наш канонический {event, data, sent_at} с HMAC-подписью.
// slack — Slack Block Kit-совместимый payload (для Slack incoming-webhook).
export type WebhookFormat = "raw" | "slack";

type Webhook = {
  id: string;
  url: string;
  events: WebhookEvent[];
  format: WebhookFormat;
  active: boolean;
  last_delivery_at?: string | null;
  last_status?: number | null;
  created_at: string;
};

type CreateWebhookRes = {
  secret: string; // plaintext, ровно один раз
  webhook: Webhook;
};

export type { Webhook, CreateWebhookRes };
