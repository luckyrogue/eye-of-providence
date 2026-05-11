import { invoke } from "@tauri-apps/api/core";

export type PreflightStatus = "ok" | "warn" | "error";

export type PreflightCheck = {
  id: string;
  label: string;
  status: PreflightStatus;
  message: string;
  action_url?: string;
  action_label?: string;
};

export function pendingCount(): Promise<number> {
  return invoke<number>("pending_count");
}

export function preflightRun(): Promise<PreflightCheck[]> {
  return invoke<PreflightCheck[]>("preflight_run");
}

export type AccountInfo = {
  paired: boolean;
  user_id: string | null;
  backend_url: string | null;
};

export type PairBeginResponse = {
  pair_id: string;
  secret: string;
  code: string;
  expires_in: number;
};

export type PollResponse = {
  status: "pending" | "claimed" | "expired";
  token: string | null;
  user_id: string | null;
};

export function accountInfo(): Promise<AccountInfo> {
  return invoke<AccountInfo>("account_info");
}

export function pairBegin(): Promise<PairBeginResponse> {
  return invoke<PairBeginResponse>("pair_begin");
}

export function pairPoll(pairId: string, secret: string): Promise<PollResponse> {
  return invoke<PollResponse>("pair_poll", { pairId, secret });
}

export function logout(): Promise<void> {
  return invoke<void>("logout");
}

export function setBackendUrl(url: string): Promise<void> {
  return invoke<void>("set_backend_url", { url });
}

export function openLogsFolder(): Promise<string> {
  return invoke<string>("open_logs_folder");
}

export function openUrl(url: string): Promise<void> {
  return invoke<void>("open_url", { url });
}

export function setPaused(paused: boolean): Promise<void> {
  return invoke<void>("set_paused", { paused });
}

export function isPaused(): Promise<boolean> {
  return invoke<boolean>("is_paused");
}

export function getAutostart(): Promise<boolean> {
  return invoke<boolean>("get_autostart");
}

export function setAutostart(enabled: boolean): Promise<void> {
  return invoke<void>("set_autostart", { enabled });
}
