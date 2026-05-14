import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../shared/api/http";
import type { Identity } from "../../entities/user";
export const identityKeys = {
  list: ["me.identities"] as const,
};
export const fetchIdentities = () =>
  http
    .get<{
      identities: Identity[];
    }>("/v1/me/identities")
    .then((r) => r.data.identities ?? []);
export const useIdentities = () =>
  useQuery({ queryKey: identityKeys.list, queryFn: fetchIdentities });
export const unlinkIdentity = (id: string) =>
  http.delete(`/v1/me/identities/${encodeURIComponent(id)}`).then(() => undefined);
export function useUnlinkIdentity() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: unlinkIdentity,
    onSuccess: () => qc.invalidateQueries({ queryKey: identityKeys.list }),
  });
}
