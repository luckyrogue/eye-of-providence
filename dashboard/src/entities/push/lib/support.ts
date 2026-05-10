// Browser-side флаг: можно ли вообще запросить push permissions/subscription.
// Если false — UI должен скрывать push-настройки целиком.
export const PUSH_SUPPORTED = typeof window !== "undefined"
  && "serviceWorker" in navigator
  && "PushManager" in window
  && "Notification" in window;
