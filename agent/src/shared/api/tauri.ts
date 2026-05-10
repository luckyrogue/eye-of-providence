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
