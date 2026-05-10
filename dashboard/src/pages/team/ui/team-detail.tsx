import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, PlanBadge, Tab, TabBar } from "@eop/ui";
import { useMembers, useTeamSummary, type Team } from "../../../entities/team";
import { TEAM_DETAIL_TABS, type TeamDetailTabKey } from "../model/team-detail-tabs";
import { CommitsTab } from "./commits-tab";
import { MembersTab } from "./members-tab";
import { ProjectsTab } from "./projects-tab";
import { SettingsTab } from "./settings-tab";

export function TeamDetail({ teamID, team, tz }: { teamID: string; team: Team; tz: string }) {
  const { t } = useTranslation(["app", "common"]);
  const role = team.role;
  const [tab, setTab] = useState<TeamDetailTabKey>("members");
  const members = useMembers(teamID);
  const stats = useTeamSummary(teamID);
  const note = team.subscription_note;
  const memberCount = members.data?.length ?? 0;

  // i18next plural — _one/_few/_many подбираются автоматически по locale
  // (русский имеет all 3, English только one/many).
  const memberCountLabel = t("team_detail.members_count", { count: memberCount });

  const visibleTabs = TEAM_DETAIL_TABS.filter((it) => !("ownerOnly" in it) || role === "owner");

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
          {visibleTabs.map((it) => {
            const Icon = it.icon;
            return (
              <Tab key={it.id} active={tab === it.id} onClick={() => setTab(it.id)} icon={<Icon className="h-3.5 w-3.5" />}>
                {t(it.i18nKey)}
              </Tab>
            );
          })}
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
