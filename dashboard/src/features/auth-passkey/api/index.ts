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
import { http } from "@/shared/api";
import type { Passkey } from "@/entities/user";

export const passkeyKeys = {
  list: ["me.passkeys"] as const,
};

export const fetchPasskeys = () =>
  http.get<{ passkeys: Passkey[] }>("/v1/me/passkeys").then((r) => r.data.passkeys ?? []);

export const usePasskeys = () => useQuery({ queryKey: passkeyKeys.list, queryFn: fetchPasskeys });

// Register a passkey for the authenticated user.
//
// Backend wraps the WebAuthn options in `{ options, session_id }` envelope —
// session_id is required on finish to bind challenge to the right ceremony.
// Field name is `attestation` (not `credential`) per backend contract,
// see handler_webauthn.go webauthnFinishRegisterReq.
type RegisterBeginRes = {
  options: PublicKeyCredentialCreationOptionsJSON;
  session_id: string;
};

async function registerPasskey(nickname: string): Promise<Passkey> {
  const beginRes = await http.post<RegisterBeginRes>("/v1/auth/webauthn/register/begin", {
    nickname,
  });
  const attResp: RegistrationResponseJSON = await startRegistration({
    optionsJSON: beginRes.data.options,
  });
  const finishRes = await http.post<Passkey>("/v1/auth/webauthn/register/finish", {
    session_id: beginRes.data.session_id,
    attestation: attResp,
    nickname,
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

// Login flow — backend returns `{ token, user_id }` after writing the one-shot
// handoff cookie. Caller redirects to `/auth/complete?return_to=...` so the
// SPA's auth-handoff page consumes the cookie and lands the user.
type LoginBeginRes = {
  options: PublicKeyCredentialRequestOptionsJSON;
  session_id: string;
};
type LoginFinishRes = { token: string; user_id: string; redirect_to: string };

async function loginPasskey(email: string | undefined): Promise<LoginFinishRes> {
  const beginRes = await http.post<LoginBeginRes>(
    "/v1/auth/webauthn/login/begin",
    email ? { email } : {},
  );
  const authResp: AuthenticationResponseJSON = await startAuthentication({
    optionsJSON: beginRes.data.options,
  });
  const finishRes = await http.post<LoginFinishRes>(
    "/v1/auth/webauthn/login/finish",
    {
      session_id: beginRes.data.session_id,
      assertion: authResp,
      return_to: "/",
    },
    { withCredentials: true },
  );
  return {
    ...finishRes.data,
    redirect_to: finishRes.data.redirect_to ?? "/auth/complete?return_to=/",
  };
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
