import { useTranslation } from "react-i18next";
import { StatTile } from "@eop/ui";
import { Building2, Crown, Users } from "lucide-react";
import type { AdminStats } from "../../../entities/admin";

export function Overview({ stats }: { stats: AdminStats }) {
  const { t } = useTranslation("app");
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile label={t("admin.stat_users", { defaultValue: "Users" })} value={stats.users_total} icon={<Users className="h-4 w-4" />} />
      <StatTile
        label={t("admin.stat_teams", { defaultValue: "Teams" })}
        value={stats.teams_total}
        hint={
          stats.beta_limit > 0
            ? t("admin.beta_limit_value", { n: stats.beta_limit, defaultValue: `beta limit · ${stats.beta_limit}` })
            : t("admin.beta_limit_unlimited", { defaultValue: "no limit" })
        }
        icon={<Building2 className="h-4 w-4" />}
      />
      <StatTile
        label={t("admin.stat_members_total", { defaultValue: "Memberships" })}
        value={stats.members_total}
        icon={<Crown className="h-4 w-4" />}
      />
    </div>
  );
}
