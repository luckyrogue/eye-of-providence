// Linked accounts API — /v1/me/identities.
//
// Backend lockout guard (DELETE returns 400 with code `last_auth_factor` when
// disconnecting would leave the user without any sign-in method):
//   {"error": "set a password before unlinking...", "code": "last_auth_factor"}
// The shared axios interceptor in shared/api/http.ts already lifts `code`
// onto the thrown Error, so callers can branch on it without touching the
// raw AxiosError.
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "@/shared/api/http";
import type { Identity } from "@/entities/user";

export const identityKeys = {
  list: ["me.identities"] as const,
};

export const fetchIdentities = () =>
  http.get<{ identities: Identity[] }>("/v1/me/identities").then((r) => r.data.identities ?? []);

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
