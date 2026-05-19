// Forma для редактирования team flags. RHF + zod.
// Allowlist из 6 ключей. Boolean → switch, number → input с min.
// Empty / clear → null (back to plan default).

import { useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button, Input } from "@eop/ui";
import { Loader2 } from "lucide-react";
import { useMutationToast } from "@/shared/hooks";
import { useTeamFlags, useUpdateTeamFlags, type TeamFlagsPayload } from "../api/team-flags";

type BoolFlagKey =
  | "enable_insights"
  | "enable_team_reports"
  | "enable_anomaly_detection"
  | "enable_webhooks";

type NumFlagKey = "k_anon_threshold" | "audit_log_retention_days";

const BOOL_FLAGS: BoolFlagKey[] = [
  "enable_insights",
  "enable_team_reports",
  "enable_anomaly_detection",
  "enable_webhooks",
];

const NUM_FLAGS: { key: NumFlagKey; min: number }[] = [
  { key: "k_anon_threshold", min: 1 },
  { key: "audit_log_retention_days", min: 30 },
];

const schema = z.object({
  enable_insights: z.union([z.boolean(), z.null()]),
  enable_team_reports: z.union([z.boolean(), z.null()]),
  enable_anomaly_detection: z.union([z.boolean(), z.null()]),
  enable_webhooks: z.union([z.boolean(), z.null()]),
  k_anon_threshold: z.union([z.number().int().min(1), z.null()]),
  audit_log_retention_days: z.union([z.number().int().min(30), z.null()]),
});

type FormValues = z.infer<typeof schema>;

const EMPTY: FormValues = {
  enable_insights: null,
  enable_team_reports: null,
  enable_anomaly_detection: null,
  enable_webhooks: null,
  k_anon_threshold: null,
  audit_log_retention_days: null,
};

export function TeamFlagsForm({ teamID }: { teamID: string }) {
  const { t } = useTranslation("app");
  const flags = useTeamFlags(teamID);
  const update = useUpdateTeamFlags();
  const runToast = useMutationToast();

  const initial = useMemo<FormValues>(() => coerceFlags(flags.data ?? {}), [flags.data]);

  const { register, handleSubmit, reset, watch, setValue, formState } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: EMPTY,
  });

  useEffect(() => {
    reset(initial);
  }, [initial, reset]);

  async function onSubmit(values: FormValues) {
    const payload: TeamFlagsPayload = {};
    for (const k of BOOL_FLAGS) payload[k] = values[k];
    for (const { key } of NUM_FLAGS) payload[key] = values[key];
    await runToast(update.mutateAsync({ teamID, payload }), {
      success: t("admin.team_flags.toast.saved"),
      error: t("admin.team_flags.toast.save_failed"),
    });
  }

  async function onResetAll() {
    const cleared: TeamFlagsPayload = {};
    for (const k of BOOL_FLAGS) cleared[k] = null;
    for (const { key } of NUM_FLAGS) cleared[key] = null;
    await runToast(update.mutateAsync({ teamID, payload: cleared }), {
      success: t("admin.team_flags.toast.reset"),
      error: t("admin.team_flags.toast.save_failed"),
    });
  }

  if (flags.isPending) {
    return (
      <div className="text-sm text-muted-foreground py-6 flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin" />
        {t("admin.loading")}
      </div>
    );
  }
  if (flags.isError) {
    return (
      <div className="text-sm py-6" style={{ color: "hsl(var(--destructive))" }}>
        {flags.error?.message ?? t("admin.error_lead")}
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>
      <p className="text-xs text-muted-foreground">{t("admin.team_flags.intro")}</p>

      <div className="space-y-3">
        {BOOL_FLAGS.map((key) => {
          const value = watch(key);
          const checked = value === true;
          const isOverride = value !== null;
          return (
            <div
              key={key}
              className="flex items-start justify-between gap-3 rounded-md border p-3"
              style={{ borderColor: "hsl(var(--border))" }}
            >
              <div className="min-w-0">
                <div className="text-sm font-medium">{t(`admin.team_flags.${key}.label`)}</div>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {t(`admin.team_flags.${key}.desc`)}
                </p>
                <div className="text-[11px] text-muted-foreground mt-1 font-mono">
                  {isOverride
                    ? t("admin.team_flags.source.override")
                    : t("admin.team_flags.source.plan_default")}
                </div>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                {/* eslint-disable-next-line no-restricted-syntax -- switch UI primitive */}
                <button
                  type="button"
                  role="switch"
                  aria-checked={checked}
                  aria-label={t(`admin.team_flags.${key}.label`)}
                  onClick={() =>
                    setValue(key, value === true ? false : value === false ? null : true, {
                      shouldDirty: true,
                    })
                  }
                  className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
                    checked ? "bg-foreground" : "bg-muted"
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-background shadow transition-transform ${
                      checked ? "translate-x-4" : "translate-x-0.5"
                    }`}
                  />
                </button>
                {isOverride && (
                  // eslint-disable-next-line no-restricted-syntax
                  <button
                    type="button"
                    className="text-[11px] text-muted-foreground underline"
                    onClick={() => setValue(key, null, { shouldDirty: true })}
                  >
                    {t("admin.team_flags.action.clear")}
                  </button>
                )}
              </div>
            </div>
          );
        })}

        {NUM_FLAGS.map(({ key, min }) => {
          const value = watch(key);
          const isOverride = value !== null;
          return (
            <div
              key={key}
              className="rounded-md border p-3 space-y-1.5"
              style={{ borderColor: "hsl(var(--border))" }}
            >
              <div className="flex items-center justify-between gap-2">
                <label className="text-sm font-medium" htmlFor={`flag-${key}`}>
                  {t(`admin.team_flags.${key}.label`)}
                </label>
                {isOverride && (
                  // eslint-disable-next-line no-restricted-syntax
                  <button
                    type="button"
                    className="text-[11px] text-muted-foreground underline"
                    onClick={() => setValue(key, null, { shouldDirty: true })}
                  >
                    {t("admin.team_flags.action.clear")}
                  </button>
                )}
              </div>
              <p className="text-xs text-muted-foreground">{t(`admin.team_flags.${key}.desc`)}</p>
              <Input
                id={`flag-${key}`}
                type="number"
                min={min}
                {...register(key, {
                  setValueAs: (v) => (v === "" || v === null || v === undefined ? null : Number(v)),
                })}
                placeholder={t("admin.team_flags.placeholder.use_plan_default")}
              />
            </div>
          );
        })}
      </div>

      {formState.errors.k_anon_threshold && (
        <div className="text-xs" style={{ color: "hsl(var(--destructive))" }}>
          {t("admin.team_flags.error.below_minimum", { key: "k_anon_threshold", min: 1 })}
        </div>
      )}
      {formState.errors.audit_log_retention_days && (
        <div className="text-xs" style={{ color: "hsl(var(--destructive))" }}>
          {t("admin.team_flags.error.below_minimum", {
            key: "audit_log_retention_days",
            min: 30,
          })}
        </div>
      )}

      <div className="flex items-center justify-end gap-2">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => void onResetAll()}
          disabled={update.isPending}
        >
          {t("admin.team_flags.cta.reset_all")}
        </Button>
        <Button type="submit" size="sm" disabled={update.isPending || !formState.isDirty}>
          {update.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />}
          {t("admin.team_flags.cta.save")}
        </Button>
      </div>
    </form>
  );
}

function coerceFlags(raw: Record<string, unknown>): FormValues {
  const out: FormValues = { ...EMPTY };
  for (const k of BOOL_FLAGS) {
    const v = raw[k];
    out[k] = typeof v === "boolean" ? v : null;
  }
  for (const { key } of NUM_FLAGS) {
    const v = raw[key];
    out[key] = typeof v === "number" ? v : null;
  }
  return out;
}
