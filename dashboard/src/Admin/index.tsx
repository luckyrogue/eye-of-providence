import { useState } from "react";
import { Eyebrow, Tab, TabBar } from "@eop/ui";
import { useAdminStats, useAdminTeams, useAdminUsers } from "../api/admin";
import { Overview } from "./Overview";
import { TeamsTable } from "./TeamsTable";
import { UsersTable } from "./UsersTable";

type TabKey = "overview" | "teams" | "users";

export function Admin({ tz }: { tz: string }) {
  const [tab, setTab] = useState<TabKey>("overview");
  const stats = useAdminStats();
  const teams = useAdminTeams();
  const users = useAdminUsers();

  return (
    <div className="space-y-4">
      <div className="flex items-baseline justify-between">
        <div>
          <Eyebrow>Super admin</Eyebrow>
          <h2 className="display-head text-3xl mt-2">Управление платформой</h2>
        </div>
        <TabBar>
          <Tab active={tab === "overview"} onClick={() => setTab("overview")}>Обзор</Tab>
          <Tab active={tab === "teams"} onClick={() => setTab("teams")}>Компании</Tab>
          <Tab active={tab === "users"} onClick={() => setTab("users")}>Пользователи</Tab>
        </TabBar>
      </div>

      {tab === "overview" && stats.data && <Overview stats={stats.data} />}
      {tab === "teams" && <TeamsTable teams={teams.data ?? []} users={users.data ?? []} tz={tz} />}
      {tab === "users" && <UsersTable users={users.data ?? []} tz={tz} />}
    </div>
  );
}
