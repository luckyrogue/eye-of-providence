export interface SubscriptionForm {
  plan: "free" | "pro" | "team" | "enterprise";
  until: string;
  note: string;
  recordPayment: boolean;
  amount: string;
  currency: string;
  method: string;
  paymentNote: string;
}

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
