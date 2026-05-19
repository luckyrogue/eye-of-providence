// Login page → "Sign in with passkey" inline button.
//
// Usernameless flow: we don't ask for an email. The backend issues an empty
// `allowCredentials` set so the authenticator picks the right resident key
// for our RPID; non-resident credentials are unsupported in v0.5.
//
// On success: backend writes the one-shot cookie and returns
// `{ redirect_to }`. The button navigates the browser there, which lands on
// `/auth/complete` where `useEffect` picks up the cookie. Errors are
// surfaced via toast — `NotAllowedError` = user cancelled / no matching key.
import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { Fingerprint } from "lucide-react";
import { useMutationToast } from "@/shared/hooks";
import { useLoginPasskey } from "../api";

function mapBrowserError(e: unknown, t: (k: string) => string): string {
  const err = e as { name?: string; message?: string };
  if (err?.name === "NotAllowedError") return t("auth:passkey.error_not_allowed");
  if (err?.name === "NotSupportedError") return t("auth:passkey.error_not_supported");
  return err?.message || t("auth:passkey.error_generic");
}

export function PasskeyLoginButton({ className }: { className?: string }) {
  const { t } = useTranslation(["auth"]);
  const login = useLoginPasskey();
  const runToast = useMutationToast();

  const browserSupported = typeof window !== "undefined" && !!window.PublicKeyCredential;

  async function onClick() {
    try {
      const r = await login.mutateAsync(undefined);
      window.location.href = r.redirect_to;
    } catch (e) {
      const msg = mapBrowserError(e, (k) => t(k as never));
      await runToast(Promise.reject(new Error(msg)), {});
    }
  }

  if (!browserSupported) return null;

  return (
    <Button
      type="button"
      variant="outline"
      onClick={onClick}
      disabled={login.isPending}
      className={className ?? "w-full"}
    >
      <Fingerprint className="h-4 w-4" aria-hidden />
      <span>{login.isPending ? t("auth:passkey.login_waiting") : t("auth:passkey.login")}</span>
    </Button>
  );
}
