import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Eyebrow, Tabs, TabsList, TabsTrigger } from "@eop/ui";
import { useAdminStats, useAdminTeams, useAdminUsers } from "../../entities/admin";
import { ADMIN_TABS, type AdminProps, type AdminTabKey } from "./model";
import { Overview } from "./ui/overview";
import { TeamsTable } from "./ui/teams-table";
import { UsersTable } from "./ui/users-table";

export function Admin({ tz }: AdminProps) {
  const { t } = useTranslation("app");
  const [tab, setTab] = useState<AdminTabKey>("overview");
  const { data: stats } = useAdminStats();
  const { data: teams } = useAdminTeams();
  const { data: users } = useAdminUsers();

  return (
    <div className="space-y-4">
      <Tabs value={tab} onValueChange={(v) => setTab(v as AdminTabKey)} className="space-y-4">
        <div className="flex items-baseline justify-between">
          <div>
            <Eyebrow>{t("admin.eyebrow", { defaultValue: "Super admin" })}</Eyebrow>
            <h2 className="display-head text-3xl mt-2">
              {t("admin.platform_management", { defaultValue: "Platform management" })}
            </h2>
          </div>
          <TabsList>
            {ADMIN_TABS.map((it) => (
              <TabsTrigger key={it.id} value={it.id}>
                {t(it.i18nKey)}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>

        {tab === "overview" && stats && <Overview stats={stats} />}
        {tab === "teams" && <TeamsTable teams={teams ?? []} users={users ?? []} tz={tz} />}
        {tab === "users" && <UsersTable users={users ?? []} tz={tz} />}
      </Tabs>
    </div>
  );
}
