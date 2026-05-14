import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";
import type { ListPushSubscriptionsRes, VAPIDInfoRes } from "./res";
export const pushKeys = {
  vapid: ["push.vapid"] as const,
  subscriptions: ["push.subscriptions"] as const,
};
export const fetchVAPIDKey = () =>
  http.get<VAPIDInfoRes>("/v1/me/push/vapid-key").then((r) => r.data);
export const fetchPushSubscriptions = () =>
  http
    .get<ListPushSubscriptionsRes>("/v1/me/push/subscriptions")
    .then((r) => r.data.subscriptions ?? []);
export const subscribePush = (sub: PushSubscriptionJSON) =>
  http.post("/v1/me/push/subscribe", sub).then(() => undefined);
export const unsubscribePush = (endpoint: string) =>
  http.post("/v1/me/push/unsubscribe", { endpoint }).then(() => undefined);
export const useVAPIDKey = () =>
  useQuery({
    queryKey: pushKeys.vapid,
    queryFn: fetchVAPIDKey,
    staleTime: Infinity,
    retry: false,
  });
export const usePushSubscriptions = () =>
  useQuery({ queryKey: pushKeys.subscriptions, queryFn: fetchPushSubscriptions });
export function useSubscribePush() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: subscribePush,
    onSuccess: () => qc.invalidateQueries({ queryKey: pushKeys.subscriptions }),
  });
}
export function useUnsubscribePush() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: unsubscribePush,
    onSuccess: () => qc.invalidateQueries({ queryKey: pushKeys.subscriptions }),
  });
}
