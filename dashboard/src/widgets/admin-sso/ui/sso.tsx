// SSO tab — global view всех OIDC configs across teams + force-disable.
// Wired к GET /v1/admin/sso + POST /v1/admin/sso/:id/disable.

import { useTranslation } from "react-i18next";
import { toast, useConfirm } from "@eop/ui";
import { Power, ShieldCheck } from "lucide-react";
import { useAdminSSODisable, useAdminSSOList, type AdminSSOConfig } from "@/entities/admin";
import { useMutationToast } from "@/shared/hooks";
import { formatDate } from "@/shared/lib/tz";

export function SSOConfigs({ tz }: { tz: string }) {
  const { t } = useTranslation("app");
  const { data, isPending, isError, error } = useAdminSSOList();
  const disable = useAdminSSODisable();
  const confirm = useConfirm();
  const runToast = useMutationToast();

  const handleDisable = async (cfg: AdminSSOConfig) => {
    const ok = await confirm({
      title: t("admin.sso_disable_confirm_title", { team: cfg.team_name }),
      description: t("admin.sso_disable_confirm_lead"),
      destructive: true,
      confirmText: t("admin.sso_disable_action"),
    });
    if (!ok) return;
    await runToast(disable.mutateAsync(cfg.team_id), {
      success: t("admin.sso_disabled"),
      error: t("admin.sso_disable_failed"),
    });
  };

  if (isPending) {
    return (
      <div className="eop-card" style={{ minHeight: 200 }}>
        <div className="h-5 w-40 rounded mb-3" style={{ background: "hsl(var(--muted))" }} />
      </div>
    );
  }
  if (isError) {
    return (
      <div
        className="eop-card"
        style={{
          border: "1px solid hsl(var(--destructive) / 0.4)",
          background: "hsl(var(--destructive) / 0.06)",
        }}
      >
        <div className="font-medium">{t("admin.error_title")}</div>
        <div className="text-[13px] text-muted-foreground mt-1">
          {error?.message ?? t("admin.error_lead")}
        </div>
      </div>
    );
  }

  const configs = data ?? [];

  return (
    <div className="eop-card">
      <div className="card-head">
        <div>
          <div className="card-title">{t("admin.sso_title")}</div>
          <div className="card-sub">{t("admin.sso_sub", { n: configs.length })}</div>
        </div>
      </div>
      {configs.length === 0 ? (
        <div className="text-[13px] text-muted-foreground py-3">{t("admin.sso_empty")}</div>
      ) : (
        <div className="flex flex-col gap-3">
          {configs.map((cfg) => (
            <div
              key={cfg.team_id}
              className="grid gap-4 items-start p-4 rounded-lg border"
              style={{
                gridTemplateColumns: "auto 1fr auto",
                borderColor: cfg.enabled ? "hsl(var(--success) / 0.3)" : "hsl(var(--border))",
                background: cfg.enabled ? "hsl(var(--success) / 0.04)" : "rgba(255,255,255,0.02)",
              }}
            >
              <div
                className="h-10 w-10 rounded-md grid place-items-center shrink-0"
                style={{
                  background: cfg.enabled ? "hsl(var(--success) / 0.12)" : "rgba(255,255,255,0.05)",
                  color: cfg.enabled ? "hsl(var(--success))" : "hsl(var(--muted-foreground))",
                }}
              >
                <ShieldCheck className="h-4 w-4" />
              </div>
              <div className="min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="font-medium">{cfg.team_name}</span>
                  <span
                    className="font-mono text-[10px] uppercase tracking-widest2 px-2 py-0.5 rounded"
                    style={{
                      background: cfg.enabled
                        ? "hsl(var(--success) / 0.1)"
                        : "rgba(255,255,255,0.05)",
                      color: cfg.enabled ? "hsl(var(--success))" : "hsl(var(--muted-foreground))",
                    }}
                  >
                    {cfg.enabled ? t("admin.sso_enabled") : t("admin.sso_disabled_label")}
                  </span>
                  <span className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
                    {cfg.provider}
                  </span>
                </div>
                <div className="font-mono text-[11px] text-muted-foreground truncate">
                  {cfg.oidc_issuer}
                </div>
                <div className="flex flex-wrap gap-3 mt-2 text-[11px] text-muted-foreground font-mono">
                  {cfg.allowed_domains.length > 0 && (
                    <span>
                      {t("admin.sso_domains")}: {cfg.allowed_domains.join(", ")}
                    </span>
                  )}
                  <span>JIT: {cfg.jit_provision ? `on (${cfg.jit_role})` : "off"}</span>
                  <span>
                    {t("admin.sso_updated")} {formatDate(cfg.updated_at, tz)}
                  </span>
                </div>
              </div>
              {cfg.enabled && (
                // eslint-disable-next-line no-restricted-syntax -- custom destructive action with icon, в строке
                <button
                  type="button"
                  onClick={() => void handleDisable(cfg)}
                  disabled={disable.isPending}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border text-[12px] disabled:opacity-60"
                  style={{
                    borderColor: "hsl(var(--destructive) / 0.4)",
                    color: "hsl(var(--destructive))",
                  }}
                >
                  <Power className="h-3.5 w-3.5" />
                  {t("admin.sso_force_disable")}
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// Silence unused import warnings for re-exported toast utility (we use it through
// useMutationToast helper, but having direct import keeps surface for future
// inline error toasts).
void toast;
