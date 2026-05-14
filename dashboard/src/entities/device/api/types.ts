export type DeviceKind = "ext" | "agent" | "ide";
export type Device = {
  id: string;
  kind: DeviceKind | string;
  name: string;
  prefix: string;
  created_at: string;
  last_used_at?: string | null;
};
