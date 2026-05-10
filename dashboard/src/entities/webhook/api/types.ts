// Outbound webhook subscriptions.

export type WebhookEvent = "commit.ingested" | "report.generated";

type Webhook = {
  id: string;
  url: string;
  events: WebhookEvent[];
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
