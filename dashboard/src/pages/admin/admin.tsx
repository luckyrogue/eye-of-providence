import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Eyebrow, Tab, TabBar } from "@eop/ui";
import { useAdminStats, useAdminTeams, useAdminUsers } from "../../entities/admin";
import { Overview } from "./ui/overview";
import { TeamsTable } from "./ui/teams-table";
import { UsersTable } from "./ui/users-table";

type TabKey = "overview" | "teams" | "users";

export function Admin({ tz }: { tz: string }) {
  const { t } = useTranslation("app");
  const [tab, setTab] = useState<TabKey>("overview");
  const { data: stats } = useAdminStats();
  const { data: teams } = useAdminTeams();
  const { data: users } = useAdminUsers();

  return (
    <div className="space-y-4">
      <div className="flex items-baseline justify-between">
        <div>
          <Eyebrow>{t("admin.eyebrow", { defaultValue: "Super admin" })}</Eyebrow>
          <h2 className="display-head text-3xl mt-2">
            {t("admin.platform_management", { defaultValue: "Platform management" })}
          </h2>
        </div>
        <TabBar>
          <Tab active={tab === "overview"} onClick={() => setTab("overview")}>
            {t("admin.overview")}
          </Tab>
          <Tab active={tab === "teams"} onClick={() => setTab("teams")}>
            {t("admin.teams")}
          </Tab>
          <Tab active={tab === "users"} onClick={() => setTab("users")}>
            {t("admin.users")}
          </Tab>
        </TabBar>
      </div>

      {tab === "overview" && stats && <Overview stats={stats} />}
      {tab === "teams" && <TeamsTable teams={teams ?? []} users={users ?? []} tz={tz} />}
      {tab === "users" && <UsersTable users={users ?? []} tz={tz} />}
    </div>
  );
}
