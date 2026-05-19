import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Tabs,
  TabsList,
  TabsTrigger,
} from "@eop/ui";
import { Users } from "lucide-react";
import { useBetaInfo, useTeams } from "@/entities/team";
import { useMe } from "@/entities/user";
import { CreateTeamButton } from "@/features/team-create";
import { SESSION_KEYS } from "@/shared/lib/session-storage";
import { BetaBanner } from "@/widgets/beta-banner";
import { TeamDetail } from "@/widgets/team-detail";

export function Teams({ tz }: { tz: string }) {
  const { t } = useTranslation("app");
  const teams = useTeams();
  const beta = useBetaInfo();
  const me = useMe();
  const [activeTeam, setActiveTeam] = useState<string | null>(
    localStorage.getItem(SESSION_KEYS.team),
  );

  const teamsList = useMemo(() => teams.data ?? [], [teams.data]);
  const isSuperAdmin = me.data?.global_role === "super_admin";
  const alreadyOwner = teamsList.some((tt) => tt.role === "owner");
  const ownerBlocked = alreadyOwner && !isSuperAdmin;

  useEffect(() => {
    if (teamsList.length === 0) return;
    const exists = activeTeam && teamsList.some((tt) => tt.id === activeTeam);
    if (!exists) {
      const first = teamsList[0].id;
      setActiveTeam(first);
      localStorage.setItem(SESSION_KEYS.team, first);
    }
  }, [teamsList, activeTeam]);

  function switchTeam(id: string) {
    setActiveTeam(id);
    localStorage.setItem(SESSION_KEYS.team, id);
  }

  const slotsLeft = beta.data?.slots_remaining ?? -1;
  const betaFull = !!(beta.data?.limit && beta.data.limit > 0 && slotsLeft === 0);
  const activeT = teamsList.find((tt) => tt.id === activeTeam);

  const blockedReason = ownerBlocked
    ? t("teams.blocked_owner")
    : betaFull
      ? t("teams.blocked_beta_full")
      : undefined;

  return (
    <div className="space-y-4">
      <div className="page-head">
        <div>
          <h1>{t("nav.team", { ns: "common" })}</h1>
          <div className="sub font-mono">{t("teams.lead")}</div>
        </div>
      </div>

      {beta.data && beta.data.limit > 0 && <BetaBanner beta={beta.data} />}

      <Card className="card-hover">
        <CardHeader className="flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div className="min-w-0 flex-1">
            <CardTitle className="flex items-center gap-2 font-display tracking-tight">
              <Users className="h-4 w-4 shrink-0" /> {t("teams.title")}
            </CardTitle>
            <CardDescription>{t("teams.lead")}</CardDescription>
          </div>
          <CreateTeamButton
            disabled={betaFull || ownerBlocked}
            blockedReason={blockedReason}
            onCreated={switchTeam}
            className="w-full sm:w-auto shrink-0"
          />
        </CardHeader>
        <CardContent>
          {teamsList.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("teams.empty")}</p>
          ) : (
            <Tabs value={activeTeam ?? ""} onValueChange={switchTeam}>
              <TabsList className="flex-wrap gap-2">
                {teamsList.map((team) => (
                  <TabsTrigger
                    key={team.id}
                    value={team.id}
                    className="border data-[state=active]:border-primary inline-flex max-w-full min-w-0 flex-nowrap items-end gap-2 whitespace-nowrap"
                  >
                    <span className="min-w-0 truncate">{team.name}</span>
                    <span className="shrink-0 font-mono text-[10px] uppercase tracking-widest2 opacity-70">
                      {t(`team_detail.role.${team.role}` as const)}
                    </span>
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
          )}
        </CardContent>
      </Card>

      {activeTeam && activeT && (
        <TeamDetail key={activeTeam} teamID={activeTeam} team={activeT} tz={tz} />
      )}
    </div>
  );
}
