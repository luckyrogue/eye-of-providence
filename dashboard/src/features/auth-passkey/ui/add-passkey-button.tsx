// Settings → "Add passkey" button.
//
// Two-step:
//   1. PromptDialog collects an optional nickname (e.g. "MacBook Touch ID")
//   2. useRegisterPasskey orchestrates begin → startRegistration → finish
//
// Browser errors (`NotAllowedError`, "user cancelled", lack of platform
// authenticator) are mapped to i18n keys so the toast surfaces something
// actionable. Anything else falls through to the raw message — useful for
// debugging while the integration is fresh.
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, PromptDialog } from "@eop/ui";
import { Fingerprint } from "lucide-react";
import { useMutationToast } from "@/shared/hooks";
import { useRegisterPasskey } from "../api";

function mapBrowserError(e: unknown, t: (k: string) => string): string {
  const err = e as { name?: string; message?: string };
  if (err?.name === "NotAllowedError") return t("auth:passkey.error_not_allowed");
  if (err?.name === "NotSupportedError") return t("auth:passkey.error_not_supported");
  return err?.message || t("auth:passkey.error_generic");
}

export function AddPasskeyButton() {
  const { t } = useTranslation(["common", "auth"]);
  const register = useRegisterPasskey();
  const runToast = useMutationToast();
  const [open, setOpen] = useState(false);

  const browserSupported = typeof window !== "undefined" && !!window.PublicKeyCredential;

  async function submit(nickname: string) {
    try {
      await register.mutateAsync(nickname);
      await runToast(Promise.resolve(true), { success: t("auth:passkey.add_success") });
      setOpen(false);
    } catch (e) {
      const msg = mapBrowserError(e, (k) => t(k as never));
      await runToast(Promise.reject(new Error(msg)), {});
    }
  }

  if (!browserSupported) {
    return <p className="text-xs text-muted-foreground">{t("auth:passkey.error_not_supported")}</p>;
  }

  return (
    <>
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={() => setOpen(true)}
        disabled={register.isPending}
      >
        <Fingerprint className="h-4 w-4" aria-hidden />
        <span>{t("common:settings.passkey_add")}</span>
      </Button>

      <PromptDialog
        open={open}
        title={t("auth:passkey.register_dialog_title")}
        description={t("auth:passkey.register_dialog_lead")}
        label={t("auth:passkey.register_nickname_label")}
        placeholder={t("auth:passkey.register_nickname_placeholder")}
        confirmText={t("common:settings.passkey_add")}
        cancelText={t("common:actions.cancel")}
        busy={register.isPending}
        onClose={() => setOpen(false)}
        onConfirm={submit}
      />
    </>
  );
}
