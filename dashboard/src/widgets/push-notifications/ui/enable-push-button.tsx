import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { Bell } from "lucide-react";
import { useSubscribePush, useVAPIDKey } from "@/entities/push";
import { useMutationToast } from "@/shared/hooks";
import { urlBase64ToUint8Array } from "@/shared/lib/push";

// variant: "primary" — крупная кнопка для пустого state, "ghost" — мелкая
// для добавления второго устройства.
export function EnablePushButton({ variant = "primary" }: { variant?: "primary" | "ghost" }) {
  const { t } = useTranslation("pwa");
  const vapid = useVAPIDKey();
  const subscribe = useSubscribePush();
  const runToast = useMutationToast();
  const [busy, setBusy] = useState(false);

  async function enable() {
    if (!vapid.data?.key) return;
    setBusy(true);
    try {
      const perm = await Notification.requestPermission();
      if (perm !== "granted") return;
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

  if (variant === "ghost") {
    return (
      <Button size="sm" variant="ghost" onClick={enable} disabled={busy}>
        <Bell className="h-3.5 w-3.5 mr-1" />
        {t("add_device")}
      </Button>
    );
  }
  return (
    <Button size="sm" onClick={enable} disabled={busy || !vapid.data?.key}>
      <Bell className="h-3.5 w-3.5 mr-1" />
      {t("enable")}
    </Button>
  );
}
