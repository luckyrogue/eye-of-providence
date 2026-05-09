import type { UseFormRegister } from "react-hook-form";
import { useTranslation } from "react-i18next";
import type { SubscriptionForm } from "./subscription-modal";

export function SubscriptionPaymentFields({
  register,
  enabled,
}: {
  register: UseFormRegister<SubscriptionForm>;
  enabled: boolean;
}) {
  const { t } = useTranslation("app");
  return (
    <div className="rounded-lg border bg-muted/20 p-4 space-y-3">
      <label className="flex items-center gap-2 text-sm font-medium">
        <input type="checkbox" {...register("recordPayment")} />
        {t("admin.payment_record_label")}
      </label>
      {enabled && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          <input
            type="number"
            min="0"
            {...register("amount")}
            placeholder={t("admin.payment_amount_placeholder")}
            className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono"
          />
          <input
            {...register("currency")}
            placeholder={t("admin.payment_currency_placeholder")}
            maxLength={3}
            className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono uppercase"
          />
          <select
            {...register("method")}
            className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono"
          >
            <option value="manual_transfer">manual_transfer</option>
            <option value="cash">cash</option>
            <option value="stripe">stripe</option>
            <option value="other">other</option>
          </select>
          <input
            {...register("paymentNote")}
            placeholder={t("admin.payment_method_placeholder")}
            className="rounded-md border bg-background px-2 py-1.5 text-sm"
          />
        </div>
      )}
      <p className="text-[11px] text-muted-foreground font-mono">{t("admin.payment_amount_hint")}</p>
    </div>
  );
}
