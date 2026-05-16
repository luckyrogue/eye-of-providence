// Paired native client device — browser extension, Tauri agent, VS Code.
//
// Источник правды на бэкенде — таблица api_tokens с kind IS NOT NULL.
// Pairing-code flow живёт в pairing_codes (эфемерно).

export type DeviceKind = "ext" | "agent" | "ide";

export type Device = {
  id: string;
  kind: DeviceKind | string;
  name: string;
  prefix: string;
  created_at: string;
  last_used_at?: string | null;
};
