import { z } from "zod";

// zod-схема формы подписки. Сообщения — i18n-ключи (resolves в форме через
// translateError или прямой t() при рендере). Cross-field валидация: если
// recordPayment включён и план платный, то amount > 0 и until обязателен.
export const subscriptionFormSchema = z
  .object({
    plan: z.enum(["free", "pro", "team", "enterprise"]),
    until: z.string(),
    note: z.string().max(1000, "subscription.errors.note_too_long"),
    recordPayment: z.boolean(),
    amount: z.string(),
    currency: z
      .string()
      .max(3, "subscription.errors.currency_invalid")
      .regex(/^[A-Za-z]{0,3}$/, "subscription.errors.currency_invalid"),
    method: z.string(),
    paymentNote: z.string().max(500, "subscription.errors.note_too_long"),
  })
  .superRefine((vals, ctx) => {
    if (vals.recordPayment && vals.plan !== "free") {
      const amt = Number(vals.amount);
      if (!Number.isFinite(amt) || amt <= 0) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["amount"],
          message: "subscription.errors.amount_required",
        });
      }
      if (!vals.until) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["until"],
          message: "subscription.errors.until_required",
        });
      }
    }
  });

export type SubscriptionForm = z.infer<typeof subscriptionFormSchema>;

export const DEFAULT_SUBSCRIPTION_FORM: SubscriptionForm = {
  plan: "free",
  until: "",
  note: "",
  recordPayment: true,
  amount: "",
  currency: "USD",
  method: "manual_transfer",
  paymentNote: "",
};
