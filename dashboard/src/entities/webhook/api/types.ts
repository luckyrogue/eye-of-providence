export type WebhookEvent = "commit.ingested" | "report.generated";
export type WebhookFormat = "raw" | "slack";
export type Webhook = {
  id: string;
  url: string;
  events: WebhookEvent[];
  format: WebhookFormat;
  active: boolean;
  last_delivery_at?: string | null;
  last_status?: number | null;
  created_at: string;
};
