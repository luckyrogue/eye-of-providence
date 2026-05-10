import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { Button, Dialog, DialogContent, Eyebrow, Form, Input, SimpleSelect, useConfirm } from "@eop/ui";
import {
  useAdminPayments,
  useAdminSetSubscription,
  type AdminTeam,
  type SetSubscriptionReq,
} from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { DEFAULT_SUBSCRIPTION_FORM, type SubscriptionForm } from "../model";
import { SubscriptionPaymentFields } from "./payment-fields";
import { SubscriptionPaymentsList } from "./payments-list";

export function SubscriptionModal({
  team,
  tz,
  onClose,
}: {
  team: AdminTeam | null;
  tz: string;
  onClose: () => void;
}) {
  const { t } = useTranslation(["app", "common"]);
  const setSub = useAdminSetSubscription();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const payments = useAdminPayments(team?.id ?? null);

  const form = useForm<SubscriptionForm>({
    defaultValues: DEFAULT_SUBSCRIPTION_FORM,
  });
  const { handleSubmit, watch, setValue, reset, register, control } = form;

  useEffect(() => {
    if (team) {
      reset({
        ...DEFAULT_SUBSCRIPTION_FORM,
        plan: (team.subscription_plan as SubscriptionForm["plan"]) || "free",
        until: team.subscription_until ? team.subscription_until.slice(0, 10) : "",
        note: team.subscription_note ?? "",
      });
    }
  }, [team, reset]);

  function quickExtend(months: number) {
    const current = watch("until");
    const base = current ? new Date(current) : new Date();
    if (!current || base < new Date()) base.setTime(Date.now());
    base.setMonth(base.getMonth() + months);
    setValue("until", base.toISOString().slice(0, 10), { shouldDirty: true });
  }

  async function onSave(values: SubscriptionForm) {
    if (!team) return;
    const payload: SetSubscriptionReq = {
      plan: values.plan,
      until: values.until ? new Date(values.until + "T23:59:59Z").toISOString() : "",
      note: values.note,
    };
    const amt = parseInt(values.amount, 10);
    if (values.recordPayment && values.plan !== "free" && amt > 0 && values.until) {
      payload.payment = {
        amount_cents: amt,
        currency: values.currency,
        method: values.method,
        note: values.paymentNote,
        covers_until: new Date(values.until + "T23:59:59Z").toISOString(),
      };
    }
    const ok = await runToast(setSub.mutateAsync({ teamID: team.id, payload }), {
      success: t("app:admin.subscription_save_success"),
      error: t("app:admin.subscription_update_failed"),
    });
    if (ok !== null) onClose();
  }

  async function revoke() {
    if (!team) return;
    const proceed = await confirm({
      title: t("app:admin.subscription_revoke_confirm_title", { name: team.name }),
      description: t("app:admin.subscription_revoke_confirm_lead"),
      destructive: true,
      confirmText: t("app:admin.subscription_revoke_confirm"),
    });
    if (!proceed) return;
    const ok = await runToast(
      setSub.mutateAsync({ teamID: team.id, payload: { plan: "free", until: "" } }),
      {
        success: t("app:admin.subscription_revoked_success"),
        error: t("app:admin.subscription_revoke_failed"),
      },
    );
    if (ok !== null) onClose();
  }

  if (!team) return null;
  const plan = watch("plan");
  const recordPayment = watch("recordPayment");

  return (
    <Dialog open={!!team} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <Form {...form}>
          <form onSubmit={handleSubmit(onSave)} className="p-6 space-y-5">
            <div>
              <Eyebrow>{t("app:admin.subscription_eyebrow")}</Eyebrow>
              <h3 className="display-head text-2xl mt-2">{team.name}</h3>
              <p className="text-xs text-muted-foreground mt-1">{team.owner_email ?? "—"}</p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1">
                <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
                  {t("app:admin.subscription_plan_label")}
                </label>
                <SimpleSelect
                  value={plan}
                  onValueChange={(v) => setValue("plan", v as SubscriptionForm["plan"], { shouldDirty: true })}
                  triggerClassName="w-full font-mono"
                  options={[
                    { value: "free", label: "free" },
                    { value: "pro", label: "pro" },
                    { value: "team", label: "team" },
                    { value: "enterprise", label: "enterprise" },
                  ]}
                />
              </div>
              <div className="space-y-1">
                <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
                  {t("app:admin.subscription_until_label")}
                </label>
                <Input type="date" {...register("until")} className="w-full" />
                <div className="flex gap-1.5 pt-1">
                  {[1, 3, 6, 12].map((m) => (
                    <Button
                      key={m}
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => quickExtend(m)}
                      className="text-[11px] font-mono text-muted-foreground hover:text-foreground h-auto px-1.5 py-0.5"
                    >
                      +{m === 12 ? "1y" : `${m}m`}
                    </Button>
                  ))}
                </div>
              </div>
            </div>

            <div className="space-y-1">
              <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
                {t("app:admin.subscription_note_label")}
              </label>
              <Input placeholder={t("app:admin.subscription_note_placeholder")} {...register("note")} />
            </div>

            {plan !== "free" && <SubscriptionPaymentFields control={control} enabled={recordPayment} />}

            <div className="flex items-center justify-between pt-2 border-t">
              <Button
                type="button"
                variant="ghost"
                onClick={revoke}
                disabled={setSub.isPending}
                className="text-destructive hover:bg-destructive/10"
              >
                {t("app:admin.subscription_revoke")}
              </Button>
              <div className="flex gap-2">
                <Button type="button" variant="outline" onClick={onClose}>
                  {t("common:actions.cancel")}
                </Button>
                <Button type="submit" disabled={setSub.isPending}>
                  {setSub.isPending ? "..." : t("common:actions.save")}
                </Button>
              </div>
            </div>

            <SubscriptionPaymentsList payments={payments.data ?? []} tz={tz} />
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
