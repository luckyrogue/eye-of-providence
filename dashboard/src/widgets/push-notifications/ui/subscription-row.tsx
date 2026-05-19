import { Bell } from "lucide-react";
import type { Subscription } from "@/entities/push";
import { DisablePushButton } from "@/features/push-unsubscribe";
import { parseUserAgent } from "@/shared/lib/push";

export function SubscriptionRow({ sub }: { sub: Subscription }) {
  const label = parseUserAgent(sub.user_agent ?? "");
  return (
    <li className="flex items-center justify-between py-2 gap-4">
      <div className="min-w-0 flex-1">
        <div className="text-sm flex items-center gap-2">
          <Bell className="h-3.5 w-3.5 text-muted-foreground" />
          <span>{label || sub.endpoint.split("/")[2] || "Unknown device"}</span>
        </div>
      </div>
      <DisablePushButton sub={sub} />
    </li>
  );
}
