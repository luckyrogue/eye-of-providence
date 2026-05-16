import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";
import type { WebhookEvent, WebhookFormat } from "./types";
import type { CreateWebhookRes, ListWebhooksRes } from "./res";

export const webhookKeys = {
  list: ["me.webhooks"] as const,
};

export const fetchWebhooks = () =>
  http.get<ListWebhooksRes>("/v1/me/webhooks/").then((r) => r.data.webhooks ?? []);

export const createWebhook = (url: string, events: WebhookEvent[], format: WebhookFormat) =>
  http.post<CreateWebhookRes>("/v1/me/webhooks/", { url, events, format }).then((r) => r.data);

export const deleteWebhook = (id: string) =>
  http.delete(`/v1/me/webhooks/${encodeURIComponent(id)}`).then(() => undefined);

export const useWebhooks = () => useQuery({ queryKey: webhookKeys.list, queryFn: fetchWebhooks });

export function useCreateWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      url,
      events,
      format,
    }: {
      url: string;
      events: WebhookEvent[];
      format: WebhookFormat;
    }) => createWebhook(url, events, format),
    onSuccess: () => qc.invalidateQueries({ queryKey: webhookKeys.list }),
  });
}

export function useDeleteWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteWebhook,
    onSuccess: () => qc.invalidateQueries({ queryKey: webhookKeys.list }),
  });
}
