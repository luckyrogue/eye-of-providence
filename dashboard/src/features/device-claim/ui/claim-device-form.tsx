import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { Button, Form, InputField, toast } from "@eop/ui";
import { useClaimDevice } from "../../../entities/device";

const CODE_LEN = 6;

const claimDeviceSchema = z.object({
  code: z
    .string()
    .length(CODE_LEN, "developer:devices_claim_code_invalid")
    .regex(/^[A-Z0-9]+$/, "developer:devices_claim_code_invalid"),
  name: z.string().max(64, "developer:devices_claim_name_max"),
});

type ClaimDeviceValues = z.infer<typeof claimDeviceSchema>;

export function ClaimDeviceForm() {
  const { t } = useTranslation("developer");
  const claim = useClaimDevice();
  const form = useForm<ClaimDeviceValues>({
    resolver: zodResolver(claimDeviceSchema),
    defaultValues: { code: "", name: "" },
    mode: "onChange",
  });
  const tr = (msg?: string) => (msg ? t(msg as never, { defaultValue: msg }) : msg);

  async function onSubmit(values: ClaimDeviceValues) {
    try {
      await claim.mutateAsync({ code: values.code, name: values.name });
      toast.success(t("devices_claim_success"));
      form.reset({ code: "", name: "" });
    } catch (e) {
      const err = e as Error & { code?: string };
      if (err.code === "code_already_claimed") {
        toast.error(t("devices_claim_error_used"));
      } else {
        toast.error(t("devices_claim_error_invalid"));
      }
    }
  }

  const submitting = form.formState.isSubmitting || claim.isPending;

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-3 pb-1 pt-0.5 sm:flex-row sm:flex-nowrap sm:items-end sm:gap-2 sm:overflow-x-auto"
      >
        <div className="w-full min-w-0 sm:flex-1 sm:basis-0 sm:min-w-[9rem]">
          <InputField
            control={form.control}
            name="code"
            label={t("devices_claim_code")}
            placeholder={t("devices_claim_code_placeholder")}
            maxLength={CODE_LEN}
            autoComplete="off"
            autoCapitalize="characters"
            spellCheck={false}
            className="font-mono uppercase tracking-widest text-center"
            transform={(raw: string) => raw.replace(/\s/g, "").toUpperCase()}
            translateError={tr}
            hideMessage
          />
        </div>
        <div className="w-full min-w-0 sm:flex-[1.5] sm:basis-0 sm:min-w-[10rem]">
          <InputField
            control={form.control}
            name="name"
            label={t("devices_claim_name")}
            placeholder={t("devices_claim_name_placeholder")}
            maxLength={64}
            translateError={tr}
            hideMessage
          />
        </div>
        <Button
          type="submit"
          disabled={submitting || !form.formState.isValid}
          className="w-full shrink-0 sm:w-auto"
        >
          {t("devices_claim_submit")}
        </Button>
      </form>
    </Form>
  );
}
