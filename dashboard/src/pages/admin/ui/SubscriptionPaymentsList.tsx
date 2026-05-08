import { Eyebrow } from "@eop/ui";
import type { Payment } from "../../../entities/admin";
import { formatDate } from "../../../shared/lib/tz";

export function SubscriptionPaymentsList({ payments, tz }: { payments: Payment[]; tz: string }) {
  if (payments.length === 0) return null;
  return (
    <div className="pt-4 border-t">
      <Eyebrow>История платежей · {payments.length}</Eyebrow>
      <ul className="mt-3 space-y-1.5">
        {payments.map((p) => (
          <li key={p.id} className="flex items-baseline justify-between text-xs font-mono">
            <span className="text-muted-foreground">{formatDate(p.paid_at, tz)}</span>
            <span>
              {(p.amount_cents / 100).toFixed(2)} {p.currency}
              <span className="text-muted-foreground"> · {p.method}</span>
              {p.note && <span className="text-muted-foreground"> · {p.note}</span>}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
