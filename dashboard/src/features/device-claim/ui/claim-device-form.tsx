import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Input, toast } from "@eop/ui";
import { Plug } from "lucide-react";
import { useClaimDevice } from "../../../entities/device";

const CODE_LEN = 6;

// Compact inline form: 6-char code + optional name. Дублирует «магазинный»
// UX устройств — pairing wizard'а в самой апке (extension/agent) показывает
// код, юзер вбивает его в dashboard.
export function ClaimDeviceForm() {
  const { t } = useTranslation("developer");
  const claim = useClaimDevice();
  const [code, setCode] = useState("");
  const [name, setName] = useState("");

  const trimmed = code.trim().toUpperCase();
  const canSubmit = trimmed.length === CODE_LEN && !claim.isPending;

  async function submit() {
    if (!canSubmit) return;
    try {
      await claim.mutateAsync({ code: trimmed, name });
      toast.success(t("devices_claim_success"));
      setCode("");
      setName("");
    } catch (e) {
      const err = e as Error & { code?: string };
      if (err.code === "code_already_claimed") {
        toast.error(t("devices_claim_error_used"));
      } else {
        toast.error(t("devices_claim_error_invalid"));
      }
    }
  }

  return (
    <form
      className="flex flex-col sm:flex-row gap-2 items-stretch sm:items-end"
      onSubmit={(e) => {
        e.preventDefault();
        void submit();
      }}
    >
      <div className="flex-1">
        <label htmlFor="device-code" className="text-xs text-muted-foreground block mb-1">
          {t("devices_claim_code")}
        </label>
        <Input
          id="device-code"
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\s/g, "").toUpperCase())}
          placeholder={t("devices_claim_code_placeholder")}
          maxLength={CODE_LEN}
          autoComplete="off"
          autoCapitalize="characters"
          spellCheck={false}
          className="font-mono uppercase tracking-widest text-center"
        />
      </div>
      <div className="flex-[1.5]">
        <label htmlFor="device-name" className="text-xs text-muted-foreground block mb-1">
          {t("devices_claim_name")}
        </label>
        <Input
          id="device-name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={t("devices_claim_name_placeholder")}
          maxLength={64}
        />
      </div>
      <Button type="submit" disabled={!canSubmit} className="sm:self-stretch">
        <Plug className="h-3.5 w-3.5 mr-1.5" />
        {t("devices_claim_submit")}
      </Button>
    </form>
  );
}
