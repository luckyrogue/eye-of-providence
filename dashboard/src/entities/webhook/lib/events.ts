import type { WebhookEvent } from "../api/types";

// Все доступные события, на которые можно подписать webhook. UI рендерит
// чекбоксы из этого списка.
export const ALL_WEBHOOK_EVENTS: WebhookEvent[] = ["commit.ingested", "report.generated"];
