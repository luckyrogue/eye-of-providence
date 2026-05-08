import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, PlanBadge, Tab, TabBar } from "@eop/ui";
import { FolderGit2, GitCommit, Settings, Users } from "lucide-react";
import { useMembers, useTeamSummary, type Team } from "../../../entities/team";
import { MembersTab } from "./MembersTab";
import { ProjectsTab } from "./ProjectsTab";
import { CommitsTab } from "./CommitsTab";
import { SettingsTab } from "./SettingsTab";
import { plural } from "../utils";

type TabKey = "members" | "projects" | "commits" | "settings";

export function TeamDetail({ teamID, team, tz }: { teamID: string; team: Team; tz: string }) {
  const role = team.role;
  const [tab, setTab] = useState<TabKey>("members");
  const members = useMembers(teamID);
  const stats = useTeamSummary(teamID);
  const note = team.subscription_note;
  const memberCount = members.data?.length ?? 0;

  return (
    <Card className="card-hover">
      <CardHeader className="flex-row items-center justify-between">
        <div>
          <div className="flex items-center gap-3 flex-wrap">
            <CardTitle className="font-display tracking-tight">{team.name}</CardTitle>
            <PlanBadge plan={team.subscription_plan ?? "free"} until={team.subscription_until} />
          </div>
          <CardDescription>
            {memberCount} участник{plural(memberCount)}{note ? ` · ${note}` : ""}
          </CardDescription>
        </div>
        <TabBar className="justify-end">
          <Tab active={tab === "members"} onClick={() => setTab("members")} icon={<Users className="h-3.5 w-3.5" />}>
            Участники
          </Tab>
          <Tab active={tab === "projects"} onClick={() => setTab("projects")} icon={<FolderGit2 className="h-3.5 w-3.5" />}>
            Проекты
          </Tab>
          <Tab active={tab === "commits"} onClick={() => setTab("commits")} icon={<GitCommit className="h-3.5 w-3.5" />}>
            Коммиты
          </Tab>
          {role === "owner" && (
            <Tab active={tab === "settings"} onClick={() => setTab("settings")} icon={<Settings className="h-3.5 w-3.5" />}>
              Настройки
            </Tab>
          )}
        </TabBar>
      </CardHeader>
      <CardContent className="space-y-4">
        {tab === "members" && (
          <MembersTab teamID={teamID} role={role} members={members.data ?? []} stats={stats.data ?? []} />
        )}
        {tab === "projects" && <ProjectsTab teamID={teamID} role={role} tz={tz} />}
        {tab === "commits" && <CommitsTab teamID={teamID} tz={tz} />}
        {tab === "settings" && role === "owner" && <SettingsTab team={team} />}
      </CardContent>
    </Card>
  );
}
