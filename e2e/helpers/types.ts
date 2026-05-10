// Shared types для tests. Минимальное подмножество backend store.Event.

export interface Event {
  ts?: string;
  user_id?: string;
  app_bundle: string;
  source: string; // "os" | "browser" | "ide" | "cli"
  category: string; // "idle" | "manual" | "ai" | "reading" | "refactor" | "other"
  duration_ms?: number;
  chars_in?: number;
  lines_added?: number;
  lines_removed?: number;
  file_lang?: string;
}

export interface APIToken {
  id: string;
  name: string;
  scope: string;
  prefix: string;
  created_at: string;
  last_used_at?: string | null;
  expires_at?: string | null;
}

export interface Webhook {
  id: string;
  url: string;
  events: string[];
  format: string;
  created_at: string;
}
