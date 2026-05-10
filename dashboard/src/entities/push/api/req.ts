// Web Push subscription управление через backend.
//
// Flow:
//   1. fetchVAPIDKey — получить public ECDSA P-256 key из /v1/me/push/vapid-key
//   2. browser asks permission через Notification.requestPermission()
//   3. registration.pushManager.subscribe({applicationServerKey: vapid})
//   4. POST /v1/me/push/subscribe с PushSubscriptionJSON
//   5. при unsubscribe: subscription.unsubscribe() + POST /unsubscribe

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";

type Subscription = {
  id: string;
  endpoint: string;
  user_agent?: string;
  created_at: string;
  last_used_at?: string | null;
};

type VAPIDInfo = { key: string; subject: string };

const pushKeys = {
  vapid: ["push.vapid"] as const,
  subscriptions: ["push.subscriptions"] as const,
};

export const fetchVAPIDKey = () =>
  http.get<VAPIDInfo>("/v1/me/push/vapid-key").then((r) => r.data);

export const fetchPushSubscriptions = () =>
  http
    .get<{ subscriptions: Subscription[] }>("/v1/me/push/subscriptions")
    .then((r) => r.data.subscriptions ?? []);

export const subscribePush = (sub: PushSubscriptionJSON) =>
  http.post("/v1/me/push/subscribe", sub).then(() => undefined);

export const unsubscribePush = (endpoint: string) =>
  http.post("/v1/me/push/unsubscribe", { endpoint }).then(() => undefined);

export const useVAPIDKey = () =>
  useQuery({
    queryKey: pushKeys.vapid,
    queryFn: fetchVAPIDKey,
    staleTime: Infinity, // VAPID key никогда не меняется в lifecycle деплоя
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

export type { Subscription, VAPIDInfo };
