import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  startAuthentication,
  startRegistration,
  type AuthenticationResponseJSON,
  type PublicKeyCredentialCreationOptionsJSON,
  type PublicKeyCredentialRequestOptionsJSON,
  type RegistrationResponseJSON,
} from "@simplewebauthn/browser";
import { http } from "../../shared/api/http";
import type { Passkey } from "../../entities/user";
export const passkeyKeys = {
  list: ["me.passkeys"] as const,
};
export const fetchPasskeys = () =>
  http
    .get<{
      passkeys: Passkey[];
    }>("/v1/me/passkeys")
    .then((r) => r.data.passkeys ?? []);
export const usePasskeys = () => useQuery({ queryKey: passkeyKeys.list, queryFn: fetchPasskeys });
async function registerPasskey(nickname: string): Promise<Passkey> {
  const beginRes = await http.post<PublicKeyCredentialCreationOptionsJSON>(
    "/v1/auth/webauthn/register/begin",
    { nickname },
  );
  const attResp: RegistrationResponseJSON = await startRegistration({
    optionsJSON: beginRes.data,
  });
  const finishRes = await http.post<Passkey>("/v1/auth/webauthn/register/finish", {
    nickname,
    credential: attResp,
  });
  return finishRes.data;
}
export function useRegisterPasskey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: registerPasskey,
    onSuccess: () => qc.invalidateQueries({ queryKey: passkeyKeys.list }),
  });
}
type LoginFinishRes = {
  redirect_to: string;
};
async function loginPasskey(email: string | undefined): Promise<LoginFinishRes> {
  const beginRes = await http.post<PublicKeyCredentialRequestOptionsJSON>(
    "/v1/auth/webauthn/login/begin",
    email ? { email } : {},
  );
  const authResp: AuthenticationResponseJSON = await startAuthentication({
    optionsJSON: beginRes.data,
  });
  const finishRes = await http.post<LoginFinishRes>(
    "/v1/auth/webauthn/login/finish",
    { email, credential: authResp },
    { withCredentials: true },
  );
  return finishRes.data;
}
export function useLoginPasskey() {
  return useMutation({
    mutationFn: (email?: string) => loginPasskey(email),
  });
}
export const revokePasskey = (id: string) =>
  http.delete(`/v1/me/passkeys/${encodeURIComponent(id)}`).then(() => undefined);
export function useRevokePasskey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: revokePasskey,
    onSuccess: () => qc.invalidateQueries({ queryKey: passkeyKeys.list }),
  });
}
