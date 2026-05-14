import { useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button, Input } from "@eop/ui";
import { Loader2 } from "lucide-react";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { usePlanOverrides, useUpdateTeamPlanLimits, type PlanOverridesPayload } from "../api";
type OverrideKey =
  | "max_users_per_team"
  | "max_webhooks"
  | "max_api_tokens"
  | "event_history_days"
  | "max_teams_per_user"
  | "audit_log_retention_days";
const OVERRIDE_KEYS: {
  key: OverrideKey;
  min: number;
  max: number;
}[] = [
  { key: "max_users_per_team", min: 1, max: 10000 },
  { key: "max_webhooks", min: 0, max: 1000 },
  { key: "max_api_tokens", min: 0, max: 500 },
  { key: "event_history_days", min: 7, max: 3650 },
  { key: "max_teams_per_user", min: 1, max: 100 },
  { key: "audit_log_retention_days", min: 30, max: 3650 },
];
const schema = z.object({
  max_users_per_team: z.union([z.number().int().min(1).max(10000), z.null()]),
  max_webhooks: z.union([z.number().int().min(0).max(1000), z.null()]),
  max_api_tokens: z.union([z.number().int().min(0).max(500), z.null()]),
  event_history_days: z.union([z.number().int().min(7).max(3650), z.null()]),
  max_teams_per_user: z.union([z.number().int().min(1).max(100), z.null()]),
  audit_log_retention_days: z.union([z.number().int().min(30).max(3650), z.null()]),
});
type FormValues = z.infer<typeof schema>;
const EMPTY: FormValues = {
  max_users_per_team: null,
  max_webhooks: null,
  max_api_tokens: null,
  event_history_days: null,
  max_teams_per_user: null,
  audit_log_retention_days: null,
};
export function PlanLimitsForm({ teamID }: { teamID: string }) {
  const { t } = useTranslation("app");
  const overrides = usePlanOverrides(teamID);
  const update = useUpdateTeamPlanLimits();
  const runToast = useMutationToast();
  const initial = useMemo<FormValues>(
    () => coerceOverrides(overrides.data?.overrides ?? {}),
    [overrides.data?.overrides],
  );
  const planDefaults = overrides.data?.plan_defaults ?? {};
  const { register, handleSubmit, reset, formState, watch, setValue } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: EMPTY,
  });
  useEffect(() => {
    reset(initial);
  }, [initial, reset]);
  async function onSubmit(values: FormValues) {
    const payload: PlanOverridesPayload = {};
    for (const { key } of OVERRIDE_KEYS) payload[key] = values[key];
    await runToast(update.mutateAsync({ teamID, payload }), {
      success: t("admin.plan_overrides.toast.saved"),
      error: t("admin.plan_overrides.toast.save_failed"),
    });
  }
  async function onResetAll() {
    const cleared: PlanOverridesPayload = {};
    for (const { key } of OVERRIDE_KEYS) cleared[key] = null;
    await runToast(update.mutateAsync({ teamID, payload: cleared }), {
      success: t("admin.plan_overrides.toast.reset"),
      error: t("admin.plan_overrides.toast.save_failed"),
    });
  }
  if (overrides.isPending) {
    return (
      <div className="text-sm text-muted-foreground py-6 flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin" />
        {t("admin.loading")}
      </div>
    );
  }
  if (overrides.isError) {
    return (
      <div className="text-sm py-6" style={{ color: "hsl(var(--destructive))" }}>
        {overrides.error?.message ?? t("admin.error_lead")}
      </div>
    );
  }
  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>
      <p className="text-xs text-muted-foreground">{t("admin.plan_overrides.intro")}</p>
      <div className="space-y-3">
        {OVERRIDE_KEYS.map(({ key, min, max }) => {
          const def = planDefaults[key];
          const defLabel =
            def === null || def === undefined
              ? t("admin.plan_overrides.placeholder.unlimited")
              : t("admin.plan_overrides.placeholder.use_default", { default: def });
          const value = watch(key);
          const isOverride = value !== null;
          return (
            <div
              key={key}
              className="rounded-md border p-3 space-y-1.5"
              style={{ borderColor: "hsl(var(--border))" }}
            >
              <div className="flex items-center justify-between gap-2">
                <label className="text-sm font-medium" htmlFor={`ovr-${key}`}>
                  {t(`admin.plan_overrides.${key}.label`)}
                </label>
                {isOverride && (
                  <Button
                    type="button"
                    variant="link"
                    className="h-auto p-0 text-[11px] text-muted-foreground"
                    onClick={() => setValue(key, null, { shouldDirty: true })}
                  >
                    {t("admin.plan_overrides.action.clear")}
                  </Button>
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                {t(`admin.plan_overrides.${key}.helper`, {
                  default:
                    def === null || def === undefined
                      ? t("admin.plan_overrides.placeholder.unlimited")
                      : def,
                })}
              </p>
              <Input
                id={`ovr-${key}`}
                type="number"
                min={min}
                max={max}
                {...register(key, {
                  setValueAs: (v) => (v === "" || v === null || v === undefined ? null : Number(v)),
                })}
                placeholder={defLabel}
              />
              {formState.errors[key] && (
                <div className="text-xs" style={{ color: "hsl(var(--destructive))" }}>
                  {t("admin.plan_overrides.error.out_of_range", { field: key, min, max })}
                </div>
              )}
            </div>
          );
        })}
      </div>

      <div className="flex items-center justify-end gap-2">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => void onResetAll()}
          disabled={update.isPending}
        >
          {t("admin.plan_overrides.cta.reset_all")}
        </Button>
        <Button type="submit" size="sm" disabled={update.isPending || !formState.isDirty}>
          {update.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />}
          {t("admin.plan_overrides.cta.save")}
        </Button>
      </div>
    </form>
  );
}
function coerceOverrides(raw: Record<string, unknown>): FormValues {
  const out: FormValues = { ...EMPTY };
  for (const { key } of OVERRIDE_KEYS) {
    const v = raw[key];
    out[key] = typeof v === "number" ? v : null;
  }
  return out;
}
