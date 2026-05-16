import type { Webhook } from "./types";

export type ListWebhooksRes = { webhooks: Webhook[] };

export type CreateWebhookRes = {
  secret: string; // plaintext, ровно один раз
  webhook: Webhook;
};
