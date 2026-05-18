import { invoke } from "@tauri-apps/api/core";

export type BackendState = "online" | "offline";
export type LocalApiState = "online" | "offline";

export type ConnectionStatus = {
  backend: BackendState;
  local_api: LocalApiState;
  local_api_port: number;
  paired: boolean;
  backend_url?: string;
};

export type AccountInfo = {
  paired: boolean;
  user_id: string | null;
  backend_url: string;
  token_from_env: boolean;
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

export function pendingCount(): Promise<number> {
  return invoke<number>("pending_count");
}

export function connectionStatus(): Promise<ConnectionStatus> {
  return invoke<ConnectionStatus>("connection_status");
}

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

export type PermissionsStatus = {
  accessibility: boolean;
  app_name: string;
  optional: boolean;
};

export function permissionsStatus(): Promise<PermissionsStatus> {
  return invoke<PermissionsStatus>("permissions_status");
}

export function openPermissionsSettings(): Promise<void> {
  return invoke<void>("open_permissions_settings");
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
