import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Button, Eyebrow, Input, Modal, Select, useConfirm } from "@eop/ui";
import {
  useAdminPayments,
  useAdminSetSubscription,
  type AdminTeam,
  type SetSubscriptionReq,
} from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/useMutationToast";
import { SubscriptionPaymentFields } from "./SubscriptionPaymentFields";
import { SubscriptionPaymentsList } from "./SubscriptionPaymentsList";

export interface SubscriptionForm {
  plan: "free" | "pro" | "team" | "enterprise";
  until: string;
  note: string;
  recordPayment: boolean;
  amount: string;
  currency: string;
  method: string;
  paymentNote: string;
}

const DEFAULT_FORM: SubscriptionForm = {
  plan: "free",
  until: "",
  note: "",
  recordPayment: true,
  amount: "",
  currency: "USD",
  method: "manual_transfer",
  paymentNote: "",
};

export function SubscriptionModal({
  team,
  tz,
  onClose,
}: {
  team: AdminTeam | null;
  tz: string;
  onClose: () => void;
}) {
  const setSub = useAdminSetSubscription();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const payments = useAdminPayments(team?.id ?? null);

  const { register, handleSubmit, watch, setValue, reset } = useForm<SubscriptionForm>({
    defaultValues: DEFAULT_FORM,
  });

  // При смене team — заполнить форму актуальными значениями.
  useEffect(() => {
    if (team) {
      reset({
        ...DEFAULT_FORM,
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
      success: "Подписка обновлена",
      error: "Не удалось обновить подписку",
    });
    if (ok !== null) onClose();
  }

  async function revoke() {
    if (!team) return;
    const proceed = await confirm({
      title: `Отозвать подписку у «${team.name}»?`,
      description: "Команда уйдёт на free-тариф. Это можно обратить, выдав подписку заново.",
      destructive: true,
      confirmText: "Отозвать",
    });
    if (!proceed) return;
    const ok = await runToast(
      setSub.mutateAsync({ teamID: team.id, payload: { plan: "free", until: "" } }),
      { success: "Подписка отозвана", error: "Не удалось отозвать" },
    );
    if (ok !== null) onClose();
  }

  if (!team) return null;
  const plan = watch("plan");
  const recordPayment = watch("recordPayment");

  return (
    <Modal open={!!team} onClose={onClose}>
      <form onSubmit={handleSubmit(onSave)} className="p-6 space-y-5">
        <div>
          <Eyebrow>Subscription</Eyebrow>
          <h3 className="display-head text-2xl mt-2">{team.name}</h3>
          <p className="text-xs text-muted-foreground mt-1">{team.owner_email ?? "—"}</p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <Select label="План" mono {...register("plan")} className="w-full px-3 py-2">
            <option value="free">free</option>
            <option value="pro">pro</option>
            <option value="team">team</option>
            <option value="enterprise">enterprise</option>
          </Select>
          <div className="space-y-1">
            <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
              Активна до
            </label>
            <input
              type="date"
              {...register("until")}
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            />
            <div className="flex gap-1.5 pt-1">
              {[1, 3, 6, 12].map((m) => (
                <button
                  key={m}
                  type="button"
                  onClick={() => quickExtend(m)}
                  className="text-[11px] font-mono text-muted-foreground hover:text-foreground"
                >
                  +{m === 12 ? "1y" : `${m}m`}
                </button>
              ))}
            </div>
          </div>
        </div>

        <Input label="Заметка (видна owner'у)" placeholder="напр. «Custom deal до конца года»" {...register("note")} />

        {plan !== "free" && (
          <SubscriptionPaymentFields register={register} enabled={recordPayment} />
        )}

        <div className="flex items-center justify-between pt-2 border-t">
          <Button
            type="button"
            variant="ghost"
            onClick={revoke}
            disabled={setSub.isPending}
            className="text-destructive hover:bg-destructive/10"
          >
            Отозвать (вернуть на free)
          </Button>
          <div className="flex gap-2">
            <Button type="button" variant="outline" onClick={onClose}>Отмена</Button>
            <Button type="submit" disabled={setSub.isPending}>
              {setSub.isPending ? "..." : "Сохранить"}
            </Button>
          </div>
        </div>

        <SubscriptionPaymentsList payments={payments.data ?? []} tz={tz} />
      </form>
    </Modal>
  );
}
