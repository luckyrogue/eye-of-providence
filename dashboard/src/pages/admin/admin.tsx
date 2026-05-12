// Admin panel — production-grade super_admin view. Под eop-theme:
// page-head + KPI grid + tab-segmented control в стиле topbar date-pick.
//
// Все хуки уже подключены к backend (useAdminStats / useAdminTeams /
// useAdminUsers + payments + delete + role + subscription mutations).
// Здесь добавляем loading skeletons + error states + refresh UX.

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { toast } from "@eop/ui";
import { Loader2, RefreshCw, AlertTriangle } from "lucide-react";
import { adminKeys, useAdminStats, useAdminTeams, useAdminUsers } from "../../entities/admin";
import { ADMIN_TABS, type AdminProps, type AdminTabKey } from "./model";
import { Overview } from "./ui/overview";
import { TeamsTable } from "./ui/teams-table";
import { UsersTable } from "./ui/users-table";
import { Revenue } from "./ui/revenue";
import { SSOConfigs } from "./ui/sso";
import { AuditLog } from "./ui/audit";
import { AdminSkeleton } from "./ui/admin-skeleton";

export function Admin({ tz }: AdminProps) {
  const { t } = useTranslation("app");
  const qc = useQueryClient();
  const [tab, setTab] = useState<AdminTabKey>("overview");
  const stats = useAdminStats();
  const teams = useAdminTeams();
  const users = useAdminUsers();

  const isLoading = stats.isPending || teams.isPending || users.isPending;
  const errored = stats.isError || teams.isError || users.isError;
  const isFetching = stats.isFetching || teams.isFetching || users.isFetching;

  const refresh = async () => {
    try {
      await qc.invalidateQueries({ queryKey: adminKeys.all });
      toast.success(t("admin.refreshed") || "Refreshed");
    } catch {
      toast.error(t("admin.refresh_failed") || "Refresh failed");
    }
  };

  return (
    <>
      <div className="page-head">
        <div>
          <h1>{t("admin.platform_management")}</h1>
          <div className="sub font-mono" style={{ color: "hsl(var(--muted-foreground))" }}>
            {t("admin.eyebrow")} ·{" "}
            {stats.data ? (
              <>
                {stats.data.users_total} {t("admin.stat_users").toLowerCase()} ·{" "}
                {stats.data.teams_total} {t("admin.stat_teams").toLowerCase()}
              </>
            ) : (
              "—"
            )}
          </div>
        </div>
        <div className="page-head-actions flex items-center gap-2">
          <div className="eop-tab-pills">
            {ADMIN_TABS.map((it) => (
              // eslint-disable-next-line no-restricted-syntax -- segmented control (artifact pattern)
              <button
                key={it.id}
                type="button"
                onClick={() => setTab(it.id)}
                className={tab === it.id ? "on" : ""}
              >
                {t(it.i18nKey)}
              </button>
            ))}
          </div>
          {/* eslint-disable-next-line no-restricted-syntax -- icon-btn refresh */}
          <button
            type="button"
            onClick={() => void refresh()}
            disabled={isFetching}
            aria-label={t("admin.refresh") || "Refresh"}
            className="inline-flex items-center justify-center w-9 h-9 rounded-lg border transition-colors hover:bg-foreground/5 disabled:opacity-60"
            style={{ borderColor: "hsl(var(--border))" }}
          >
            {isFetching ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4" />
            )}
          </button>
        </div>
      </div>

      {errored && (
        <div
          className="eop-card flex items-start gap-3 mb-4"
          style={{
            border: "1px solid hsl(var(--destructive) / 0.4)",
            background: "hsl(var(--destructive) / 0.06)",
          }}
        >
          <AlertTriangle
            className="h-5 w-5 shrink-0 mt-0.5"
            style={{ color: "hsl(var(--destructive))" }}
          />
          <div className="flex-1">
            <div className="font-medium text-[14px]">
              {t("admin.error_title") || "Failed to load some admin data"}
            </div>
            <div className="text-[13px] text-muted-foreground mt-1">
              {(stats.error?.message ?? teams.error?.message ?? users.error?.message) ||
                t("admin.error_lead") ||
                "Check backend connectivity and try refresh."}
            </div>
          </div>
        </div>
      )}

      {isLoading && !stats.data && !teams.data && !users.data ? (
        <AdminSkeleton />
      ) : (
        <>
          {tab === "overview" && stats.data && (
            <Overview stats={stats.data} teamsCount={teams.data?.length} />
          )}
          {tab === "teams" && (
            <TeamsTable teams={teams.data ?? []} users={users.data ?? []} tz={tz} />
          )}
          {tab === "users" && <UsersTable users={users.data ?? []} tz={tz} />}
          {tab === "revenue" && <Revenue tz={tz} />}
          {tab === "sso" && <SSOConfigs tz={tz} />}
          {tab === "audit" && <AuditLog tz={tz} />}
        </>
      )}
    </>
  );
}
