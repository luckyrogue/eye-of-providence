// Push notifications setup widget — Settings page section.
//
// 4 состояния:
//   - browser не поддерживает: hide widget
//   - не subscribed: button "Включить"
//   - subscribed: список endpoint'ов + unsubscribe button per row
//   - VAPID не настроен (server returns 503): hide или show "coming soon"

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, useConfirm } from "@eop/ui";
import { Bell, Trash2 } from "lucide-react";
import {
  useVAPIDKey, usePushSubscriptions, useSubscribePush, useUnsubscribePush,
  type Subscription,
} from "../../entities/push";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";

const PUSH_SUPPORTED = typeof window !== "undefined"
  && "serviceWorker" in navigator
  && "PushManager" in window
  && "Notification" in window;

export function PushNotificationsWidget() {
  const { t } = useTranslation("pwa");
  const vapid = useVAPIDKey();
  const subs = usePushSubscriptions();
  const subscribe = useSubscribePush();
  const unsubscribe = useUnsubscribePush();
  const confirm = useConfirm();
  const runToast = useMutationToast();

  const [permission, setPermission] = useState<NotificationPermission>(
    typeof Notification !== "undefined" ? Notification.permission : "default",
  );
  const [busy, setBusy] = useState(false);

  // Скрываем widget если нет browser support или server не настроен
  // (vapid endpoint вернул 503).
  if (!PUSH_SUPPORTED) return null;
  if (vapid.isError) return null;

  async function handleEnable() {
    if (!vapid.data?.key) return;
    setBusy(true);
    try {
      const perm = await Notification.requestPermission();
      setPermission(perm);
      if (perm !== "granted") {
        return;
      }
      const reg = await navigator.serviceWorker.ready;
      const sub = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        // PushManager type требует BufferSource. .buffer гарантированно
        // ArrayBuffer (мы создаём new Uint8Array(rawData.length) — не SAB).
        applicationServerKey: urlBase64ToUint8Array(vapid.data.key).buffer as ArrayBuffer,
      });
      await runToast(subscribe.mutateAsync(sub.toJSON() as PushSubscriptionJSON), {});
    } catch (err) {
      console.warn("[push] subscribe failed", err);
    } finally {
      setBusy(false);
    }
  }

  async function handleDisable(s: Subscription) {
    const ok = await confirm({
      title: t("disable_confirm"),
      destructive: true,
    });
    if (!ok) return;
    // Unsubscribe браузер локально, чтобы permission entry не висел.
    try {
      const reg = await navigator.serviceWorker.ready;
      const browserSub = await reg.pushManager.getSubscription();
      if (browserSub && browserSub.endpoint === s.endpoint) {
        await browserSub.unsubscribe();
      }
    } catch {
      // best-effort
    }
    await runToast(unsubscribe.mutateAsync(s.endpoint), {});
  }

  const subList = subs.data ?? [];

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Bell className="h-4 w-4 text-muted-foreground" />
          <CardTitle>{t("push_title")}</CardTitle>
        </div>
        <CardDescription>{t("push_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        {permission === "denied" ? (
          <p className="text-sm text-muted-foreground">{t("permission_denied")}</p>
        ) : subList.length === 0 ? (
          <Button size="sm" onClick={handleEnable} disabled={busy || !vapid.data?.key}>
            <Bell className="h-3.5 w-3.5 mr-1" />
            {t("enable")}
          </Button>
        ) : (
          <ul className="divide-y">
            {subList.map((s) => <SubscriptionRow key={s.id} sub={s} onDisable={() => handleDisable(s)} />)}
            <li className="pt-3">
              <Button size="sm" variant="ghost" onClick={handleEnable} disabled={busy}>
                <Bell className="h-3.5 w-3.5 mr-1" />
                {t("add_device")}
              </Button>
            </li>
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function SubscriptionRow({ sub, onDisable }: { sub: Subscription; onDisable: () => void }) {
  const { t } = useTranslation("pwa");
  // user_agent — длинный, отображаем только browser+platform идею.
  const label = parseUserAgent(sub.user_agent ?? "");
  return (
    <li className="flex items-center justify-between py-2 gap-4">
      <div className="min-w-0 flex-1">
        <div className="text-sm flex items-center gap-2">
          <Bell className="h-3.5 w-3.5 text-muted-foreground" />
          <span>{label || sub.endpoint.split("/")[2] || "Unknown device"}</span>
        </div>
      </div>
      <Button size="sm" variant="ghost" onClick={onDisable}>
        <Trash2 className="h-3.5 w-3.5 mr-1" />
        {t("disable")}
      </Button>
    </li>
  );
}

// urlBase64ToUint8Array — VAPID public key приходит base64url-encoded
// (RFC 4648 §5). PushManager.subscribe требует Uint8Array.
function urlBase64ToUint8Array(b64: string): Uint8Array {
  const padding = "=".repeat((4 - (b64.length % 4)) % 4);
  const base64 = (b64 + padding).replace(/-/g, "+").replace(/_/g, "/");
  const rawData = window.atob(base64);
  const out = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; i++) out[i] = rawData.charCodeAt(i);
  return out;
}

// parseUserAgent — короткий browser+platform label из UA-string.
// Heuristic, не идеально, но хватает для UI.
function parseUserAgent(ua: string): string {
  if (!ua) return "";
  const platform = /iPhone|iPad/.test(ua) ? "iOS"
    : /Android/.test(ua) ? "Android"
      : /Macintosh/.test(ua) ? "Mac"
        : /Windows/.test(ua) ? "Windows"
          : /Linux/.test(ua) ? "Linux"
            : "";
  const browser = /Edg\//.test(ua) ? "Edge"
    : /Chrome\//.test(ua) ? "Chrome"
      : /Firefox\//.test(ua) ? "Firefox"
        : /Safari\//.test(ua) ? "Safari"
          : "";
  return [browser, platform].filter(Boolean).join(" • ");
}
