// WebAuthn / passkey API hooks.
//
// Endpoints (backend Phase 2, see backend-progress.md):
//   GET    /v1/me/passkeys                       → list
//   POST   /v1/auth/webauthn/register/begin      → PublicKeyCredentialCreationOptionsJSON
//   POST   /v1/auth/webauthn/register/finish     → { id, ... }
//   POST   /v1/auth/webauthn/login/begin         → PublicKeyCredentialRequestOptionsJSON
//   POST   /v1/auth/webauthn/login/finish        → { redirect_to } (cookie handoff, see /auth/complete)
//   DELETE /v1/me/passkeys/:id
//
// Register flow:
//   1. backend returns creation options
//   2. browser calls startRegistration → AuthenticatorAttestationResponseJSON
//   3. backend verifies + persists, returns Passkey row
//
// Login flow is identical with startAuthentication; the finish call ends with
// a cookie handoff (matches OAuth) so success here means "navigate the user
// to /auth/complete".
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
  http.get<{ passkeys: Passkey[] }>("/v1/me/passkeys").then((r) => r.data.passkeys ?? []);

export const usePasskeys = () => useQuery({ queryKey: passkeyKeys.list, queryFn: fetchPasskeys });

// Register a passkey for the authenticated user.
//
// `optionsJSON` from the backend is already in the JSON form expected by
// @simplewebauthn/browser — no Base64URL decoding needed on our side.
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

// Login flow — backend returns `{ redirect_to: "/auth/complete?return_to=..." }`
// after writing the one-shot cookie. XHR can't follow cross-origin 302 safely,
// so the contract is an explicit redirect URL instead of HTTP 302. Caller
// uses `window.location.href = redirect_to` to land on the SPA handoff route.
type LoginFinishRes = { redirect_to: string };

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
