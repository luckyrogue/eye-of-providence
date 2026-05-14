import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { Fingerprint } from "lucide-react";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { useLoginPasskey } from "../api";
function mapBrowserError(e: unknown, t: (k: string) => string): string {
  const err = e as {
    name?: string;
    message?: string;
  };
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
