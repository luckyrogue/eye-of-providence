export type Subscription = {
  id: string;
  endpoint: string;
  user_agent?: string;
  created_at: string;
  last_used_at?: string | null;
};
export type VAPIDInfo = {
  key: string;
  subject: string;
};
