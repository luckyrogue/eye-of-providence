import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, PromptDialog } from "@eop/ui";
import { Plus, Users } from "lucide-react";
import { useTeams, useBetaInfo, useCreateTeam } from "../../entities/team";
import { useMe } from "../../entities/user";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";
import { BetaBanner } from "../../widgets/beta-banner";
import { TeamDetail } from "./ui/team-detail";

export function Teams({ tz }: { tz: string }) {
  const { t } = useTranslation(["app", "errors", "common"]);
  const teams = useTeams();
  const beta = useBetaInfo();
  const me = useMe();
  const createTeam = useCreateTeam();
  const runToast = useMutationToast();
  const [activeTeam, setActiveTeam] = useState<string | null>(localStorage.getItem("eop_team"));
  const [newTeamOpen, setNewTeamOpen] = useState(false);

  const teamsList = teams.data ?? [];
  // 1 owner = 1 company invariant (бэкенд тоже ловит, но UI не должен звать
  // запрос, обречённый на 403). Super_admin исключён.
  const isSuperAdmin = me.data?.global_role === "super_admin";
  const alreadyOwner = teamsList.some((t) => t.role === "owner");
  const ownerBlocked = alreadyOwner && !isSuperAdmin;

  // Sync activeTeam с реальным списком: если ничего не выбрано, или выбранная
  // команда удалена — переключиться на первую доступную. Делаем в effect, не
  // в render, чтобы не нарушать React contract.
  useEffect(() => {
    if (teamsList.length === 0) return;
    const exists = activeTeam && teamsList.some((t) => t.id === activeTeam);
    if (!exists) {
      const first = teamsList[0].id;
      setActiveTeam(first);
      localStorage.setItem("eop_team", first);
    }
  }, [teamsList, activeTeam]);

  function switchTeam(id: string) {
    setActiveTeam(id);
    localStorage.setItem("eop_team", id);
  }

  async function onCreateTeam(name: string) {
    try {
      const r = await createTeam.mutateAsync(name);
      runToast(Promise.resolve(r), { success: t("app:teams.created_toast") });
      switchTeam(r.id);
      setNewTeamOpen(false);
    } catch (e) {
      const code = (e as { code?: string }).code;
      const errorMsg = code ? t(`errors:${code}`, { defaultValue: t("errors:generic") }) : t("errors:generic");
      runToast(Promise.reject(new Error(errorMsg)), { error: t("errors:team_create_failed") });
    }
  }

  const slotsLeft = beta.data?.slots_remaining ?? -1;
  const betaFull = !!(beta.data?.limit && beta.data.limit > 0 && slotsLeft === 0);
  const activeT = teamsList.find((t) => t.id === activeTeam);

  return (
    <div className="space-y-4">
      {beta.data && beta.data.limit > 0 && <BetaBanner beta={beta.data} />}

      <Card className="card-hover">
        <CardHeader className="flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div className="min-w-0 flex-1">
            <CardTitle className="flex items-center gap-2 font-display tracking-tight">
              <Users className="h-4 w-4 shrink-0" /> {t("app:teams.title")}
            </CardTitle>
            <CardDescription>{t("app:teams.lead")}</CardDescription>
          </div>
          <Button
            size="sm"
            onClick={() => setNewTeamOpen(true)}
            disabled={createTeam.isPending || betaFull || ownerBlocked}
            className="w-full sm:w-auto shrink-0"
            title={
              ownerBlocked
                ? t("app:teams.blocked_owner")
                : betaFull
                  ? t("app:teams.blocked_beta_full")
                  : undefined
            }
          >
            <Plus className="h-3.5 w-3.5 mr-1" /> {t("app:teams.new_team")}
          </Button>
        </CardHeader>
        <CardContent>
          {teamsList.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("app:teams.empty")}</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {teamsList.map((team) => (
                <button
                  key={team.id}
                  type="button"
                  onClick={() => switchTeam(team.id)}
                  className={`rounded-md border px-3 py-1.5 text-sm transition-colors ${
                    activeTeam === team.id
                      ? "bg-primary text-primary-foreground border-primary"
                      : "bg-card hover:bg-secondary"
                  }`}
                >
                  {team.name}
                  <span className="ml-2 font-mono text-[10px] uppercase tracking-widest2 opacity-70">
                    {t(`app:team_detail.role.${team.role}` as const, { defaultValue: team.role })}
                  </span>
                </button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {activeTeam && activeT && (
        <TeamDetail key={activeTeam} teamID={activeTeam} team={activeT} tz={tz} />
      )}

      <PromptDialog
        open={newTeamOpen}
        title={t("app:teams.create_dialog_title")}
        description={t("app:teams.create_dialog_lead")}
        label={t("app:teams.create_dialog_label")}
        placeholder={t("app:teams.create_dialog_placeholder")}
        confirmText={t("common:actions.create")}
        busy={createTeam.isPending}
        onClose={() => setNewTeamOpen(false)}
        onConfirm={onCreateTeam}
      />
    </div>
  );
}
