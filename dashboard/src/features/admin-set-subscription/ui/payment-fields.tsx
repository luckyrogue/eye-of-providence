import { Controller, type Control } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { CheckboxField, Input, SimpleSelect, type SimpleSelectOption } from "@eop/ui";
import type { SubscriptionForm } from "../model";

const METHOD_OPTIONS: SimpleSelectOption[] = [
  { value: "manual_transfer", label: "manual_transfer" },
  { value: "cash", label: "cash" },
  { value: "stripe", label: "stripe" },
  { value: "other", label: "other" },
];

export function SubscriptionPaymentFields({
  control,
  enabled,
}: {
  control: Control<SubscriptionForm>;
  enabled: boolean;
}) {
  const { t } = useTranslation("app");
  return (
    <div className="rounded-lg border bg-muted/20 p-4 space-y-3">
      <CheckboxField
        control={control}
        name="recordPayment"
        label={t("admin.payment_record_label")}
        hideMessage
      />
      {enabled && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
          <Controller
            control={control}
            name="amount"
            render={({ field }) => (
              <Input
                type="number"
                min={0}
                inputMode="numeric"
                value={field.value}
                onChange={(e) => field.onChange(e.target.value)}
                placeholder={t("admin.payment_amount_placeholder")}
                className="font-mono"
              />
            )}
          />
          <Controller
            control={control}
            name="currency"
            render={({ field }) => (
              <Input
                value={field.value}
                onChange={(e) => field.onChange(e.target.value)}
                placeholder={t("admin.payment_currency_placeholder")}
                maxLength={3}
                className="font-mono uppercase"
              />
            )}
          />
          <Controller
            control={control}
            name="method"
            render={({ field }) => (
              <SimpleSelect
                value={field.value}
                onValueChange={field.onChange}
                options={METHOD_OPTIONS}
                triggerClassName="font-mono"
              />
            )}
          />
          <Controller
            control={control}
            name="paymentNote"
            render={({ field }) => (
              <Input
                value={field.value}
                onChange={(e) => field.onChange(e.target.value)}
                placeholder={t("admin.payment_method_placeholder")}
              />
            )}
          />
        </div>
      )}
      <p className="text-[11px] text-muted-foreground font-mono">
        {t("admin.payment_amount_hint")}
      </p>
    </div>
  );
}
