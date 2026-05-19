// Push notifications setup widget — Settings page section.
//
// 4 состояния:
//   - browser не поддерживает: hide widget
//   - server не настроил VAPID (503): hide
//   - permission denied: показ message
//   - subscribed: список endpoint'ов + кнопка добавить устройство

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Bell } from "lucide-react";
import { PUSH_SUPPORTED, useVAPIDKey, usePushSubscriptions } from "@/entities/push";
import { EnablePushButton } from "./enable-push-button";
import { SubscriptionRow } from "./subscription-row";

export function PushNotificationsWidget() {
  const { t } = useTranslation("pwa");
  const vapid = useVAPIDKey();
  const subs = usePushSubscriptions();
  const [permission] = useState<NotificationPermission>(
    typeof Notification !== "undefined" ? Notification.permission : "default",
  );

  if (!PUSH_SUPPORTED) return null;
  if (vapid.isError) return null;

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
          <EnablePushButton />
        ) : (
          <ul className="divide-y">
            {subList.map((s) => (
              <SubscriptionRow key={s.id} sub={s} />
            ))}
            <li className="pt-3">
              <EnablePushButton variant="ghost" />
            </li>
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
