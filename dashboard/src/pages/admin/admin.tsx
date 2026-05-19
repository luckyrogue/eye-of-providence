import { useState } from "react";
import { useTranslation } from "react-i18next";
import { AlertTriangle } from "lucide-react";
import { useAdminStats, useAdminTeams, useAdminUsers } from "@/entities/admin";
import { AdminRefreshButton } from "@/features/admin-refresh";
import { APITokensCrossUser } from "@/widgets/admin-api-tokens-cross-user";
import { AuditLog } from "@/widgets/admin-audit";
import { ContentEditorPage } from "@/widgets/admin-content-editor";
import { EmailTemplatesPage } from "@/widgets/admin-email-templates";
import { Overview } from "@/widgets/admin-overview";
import { Revenue } from "@/widgets/admin-revenue";
import { AdminSkeleton } from "@/widgets/admin-skeleton";
import { SSOConfigs } from "@/widgets/admin-sso";
import { TeamsTable } from "@/widgets/admin-teams-table";
import { UsersTable } from "@/widgets/admin-users-table";
import { WebhooksCrossTeam } from "@/widgets/admin-webhooks-cross-team";
import { ADMIN_TABS, type AdminProps, type AdminTabKey } from "./model";

export function Admin({ tz }: AdminProps) {
  const { t } = useTranslation("app");
  const [tab, setTab] = useState<AdminTabKey>("overview");
  const stats = useAdminStats();
  const teams = useAdminTeams();
  const users = useAdminUsers();

  const isLoading = stats.isPending || teams.isPending || users.isPending;
  const errored = stats.isError || teams.isError || users.isError;
  const isFetching = stats.isFetching || teams.isFetching || users.isFetching;

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
          <AdminRefreshButton isFetching={isFetching} />
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
          {tab === "email_templates" && <EmailTemplatesPage />}
          {tab === "content" && <ContentEditorPage />}
          {tab === "webhooks" && <WebhooksCrossTeam tz={tz} />}
          {tab === "tokens" && <APITokensCrossUser tz={tz} />}
          {tab === "audit" && <AuditLog tz={tz} />}
        </>
      )}
    </>
  );
}
