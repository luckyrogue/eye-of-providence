// Admin overview — 4 KPI tiles в стиле dashboard v2 (eop-card).
//   1. Total users (with super_admin count)
//   2. Total teams (с beta limit indicator)
//   3. Total memberships (отражает density: members_total / teams_total)
//   4. Beta health: free slots remaining

import { useTranslation } from "react-i18next";
import { Users, Building2, Crown, Activity } from "lucide-react";
import type { AdminStats } from "../../../entities/admin";

export function Overview({ stats, teamsCount }: { stats: AdminStats; teamsCount?: number }) {
  const { t } = useTranslation("app");
  const betaSlotsFree =
    stats.beta_limit > 0 ? Math.max(0, stats.beta_limit - stats.teams_total) : null;
  const avgMembersPerTeam =
    stats.teams_total > 0 ? (stats.members_total / stats.teams_total).toFixed(1) : "0.0";

  return (
    <div className="kpi-grid">
      <Tile
        icon={<Users className="h-4 w-4" />}
        label={t("admin.stat_users")}
        value={stats.users_total.toLocaleString()}
        hint={
          teamsCount !== undefined
            ? t("admin.users_per_team_hint", {
                n:
                  stats.users_total > 0
                    ? (stats.users_total / Math.max(1, teamsCount)).toFixed(1)
                    : "—",
              }) || `${(stats.users_total / Math.max(1, teamsCount)).toFixed(1)} per team`
            : undefined
        }
      />
      <Tile
        icon={<Building2 className="h-4 w-4" />}
        label={t("admin.stat_teams")}
        value={stats.teams_total.toLocaleString()}
        hint={
          stats.beta_limit > 0
            ? t("admin.beta_limit_value", { n: stats.beta_limit })
            : t("admin.beta_limit_unlimited")
        }
      />
      <Tile
        icon={<Crown className="h-4 w-4" />}
        label={t("admin.stat_members_total")}
        value={stats.members_total.toLocaleString()}
        hint={
          t("admin.avg_members_per_team", { n: avgMembersPerTeam }) ||
          `~${avgMembersPerTeam} avg / team`
        }
      />
      <Tile
        icon={<Activity className="h-4 w-4" />}
        label={t("admin.stat_beta_capacity")}
        value={betaSlotsFree === null ? "∞" : betaSlotsFree.toString()}
        hint={
          betaSlotsFree === null
            ? t("admin.beta_limit_unlimited")
            : betaSlotsFree === 0
              ? t("admin.beta_full")
              : t("admin.beta_slots_free", { n: betaSlotsFree })
        }
        accent={betaSlotsFree !== null && betaSlotsFree === 0}
      />
    </div>
  );
}

function Tile({
  icon,
  label,
  value,
  hint,
  accent,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  hint?: string;
  accent?: boolean;
}) {
  return (
    <div className="kpi" style={accent ? { borderColor: "hsl(var(--accent))" } : undefined}>
      <span className="kpi-label inline-flex items-center gap-1.5">
        <span style={{ color: accent ? "hsl(var(--accent))" : "hsl(var(--muted-foreground))" }}>
          {icon}
        </span>
        {label}
      </span>
      <span className="kpi-value">{value}</span>
      {hint && (
        <span
          className="kpi-delta flat"
          style={accent ? { color: "hsl(var(--accent))" } : undefined}
        >
          {hint}
        </span>
      )}
    </div>
  );
}
