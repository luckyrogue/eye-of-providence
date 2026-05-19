import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "@/shared/api/http";
import type { Device } from "./types";
import type { ListDevicesRes } from "./res";

export const deviceKeys = {
  list: ["me.devices"] as const,
};

export const fetchDevices = () =>
  http.get<ListDevicesRes>("/v1/me/devices/").then((r) => r.data.devices ?? []);

export const claimDevice = (code: string, name: string) =>
  http
    .post<Device>("/v1/me/devices/claim", { code: code.trim().toUpperCase(), name: name.trim() })
    .then((r) => r.data);

export const revokeDevice = (id: string) =>
  http.delete(`/v1/me/devices/${encodeURIComponent(id)}`).then(() => undefined);

export const useDevices = () => useQuery({ queryKey: deviceKeys.list, queryFn: fetchDevices });

export function useClaimDevice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ code, name }: { code: string; name: string }) => claimDevice(code, name),
    onSuccess: () => qc.invalidateQueries({ queryKey: deviceKeys.list }),
  });
}

export function useRevokeDevice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: revokeDevice,
    onSuccess: () => qc.invalidateQueries({ queryKey: deviceKeys.list }),
  });
}
