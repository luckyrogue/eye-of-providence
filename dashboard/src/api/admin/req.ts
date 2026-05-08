type UpdateUserReq = { global_role?: string; display_name?: string };

type AddMemberReq = { email: string; role: string };

type SubscriptionPaymentReq = {
  amount_cents: number;
  currency: string;
  method: string;
  note: string;
  covers_until: string;
};

type SetSubscriptionReq = {
  plan?: string;
  until?: string | null;
  note?: string | null;
  payment?: SubscriptionPaymentReq;
};

export type { UpdateUserReq, AddMemberReq, SubscriptionPaymentReq, SetSubscriptionReq };
