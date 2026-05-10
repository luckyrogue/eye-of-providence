import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, PlanBadge, Tab, TabBar } from "@eop/ui";
import { FolderGit2, GitCommit, Settings, Users } from "lucide-react";
import { useMembers, useTeamSummary, type Team } from "../../../entities/team";
import { MembersTab } from "./members-tab";
import { ProjectsTab } from "./projects-tab";
import { CommitsTab } from "./commits-tab";
import { SettingsTab } from "./settings-tab";

type TabKey = "members" | "projects" | "commits" | "settings";

export function TeamDetail({ teamID, team, tz }: { teamID: string; team: Team; tz: string }) {
  const { t } = useTranslation(["app", "common"]);
  const role = team.role;
  const [tab, setTab] = useState<TabKey>("members");
  const members = useMembers(teamID);
  const stats = useTeamSummary(teamID);
  const note = team.subscription_note;
  const memberCount = members.data?.length ?? 0;

  // i18next plural — _one/_few/_many подбираются автоматически по locale
  // (русский имеет all 3, English только one/many).
  const memberCountLabel = t("team_detail.members_count", { count: memberCount });

  return (
    <Card className="card-hover">
      <CardHeader className="flex-col lg:flex-row lg:items-center lg:justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-3 flex-wrap">
            <CardTitle className="font-display tracking-tight">{team.name}</CardTitle>
            <PlanBadge
              plan={team.subscription_plan ?? "free"}
              until={team.subscription_until}
              untilLabel={t("common:plan_badge.until", { defaultValue: "until" })}
              expiredLabel={t("common:plan_badge.expired", { defaultValue: "expired" })}
            />
          </div>
          <CardDescription>
            {memberCountLabel}
            {note ? ` · ${note}` : ""}
          </CardDescription>
        </div>
        <TabBar className="justify-start lg:justify-end overflow-x-auto -mx-1 px-1">
          <Tab active={tab === "members"} onClick={() => setTab("members")} icon={<Users className="h-3.5 w-3.5" />}>
            {t("team_detail.tabs.members")}
          </Tab>
          <Tab active={tab === "projects"} onClick={() => setTab("projects")} icon={<FolderGit2 className="h-3.5 w-3.5" />}>
            {t("team_detail.tabs.projects")}
          </Tab>
          <Tab active={tab === "commits"} onClick={() => setTab("commits")} icon={<GitCommit className="h-3.5 w-3.5" />}>
            {t("team_detail.tabs.commits")}
          </Tab>
          {role === "owner" && (
            <Tab active={tab === "settings"} onClick={() => setTab("settings")} icon={<Settings className="h-3.5 w-3.5" />}>
              {t("team_detail.tabs.settings")}
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
