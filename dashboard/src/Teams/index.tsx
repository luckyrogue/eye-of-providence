import { useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Plus, Users } from "lucide-react";
import { useTeams, useBetaInfo, useCreateTeam } from "../api/teams";
import { useMutationToast } from "../hooks/useMutationToast";
import { BetaBanner } from "./BetaBanner";
import { TeamDetail } from "./TeamDetail";
import { translateRole } from "./utils";

export function Teams({ tz }: { tz: string }) {
  const teams = useTeams();
  const beta = useBetaInfo();
  const createTeam = useCreateTeam();
  const runToast = useMutationToast();
  const [activeTeam, setActiveTeam] = useState<string | null>(localStorage.getItem("eop_team"));

  const teamsList = teams.data ?? [];
  // Если activeTeam ещё не выбран и список загрузился — выбрать первый.
  if (teamsList.length > 0 && !activeTeam) {
    const first = teamsList[0].id;
    setActiveTeam(first);
    localStorage.setItem("eop_team", first);
  }
  // Если activeTeam удалён из списка (deleted) — сбросить.
  if (activeTeam && teamsList.length > 0 && !teamsList.find((t) => t.id === activeTeam)) {
    localStorage.removeItem("eop_team");
    setActiveTeam(teamsList[0]?.id ?? null);
  }

  function switchTeam(id: string) {
    setActiveTeam(id);
    localStorage.setItem("eop_team", id);
  }

  async function onNewTeam() {
    const name = prompt("Название команды");
    if (!name?.trim()) return;
    const r = await runToast(createTeam.mutateAsync(name.trim()), {
      success: "Команда создана",
      error: "Не удалось создать команду",
    });
    if (r) switchTeam(r.id);
  }

  const slotsLeft = beta.data?.slots_remaining ?? -1;
  const betaFull = !!(beta.data?.limit && beta.data.limit > 0 && slotsLeft === 0);
  const activeT = teamsList.find((t) => t.id === activeTeam);

  return (
    <div className="space-y-4">
      {beta.data && beta.data.limit > 0 && <BetaBanner beta={beta.data} />}

      <Card className="card-hover">
        <CardHeader className="flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2 font-display tracking-tight">
              <Users className="h-4 w-4" /> Мои команды
            </CardTitle>
            <CardDescription>Можешь состоять в нескольких — переключайся внизу.</CardDescription>
          </div>
          <Button size="sm" onClick={onNewTeam} disabled={createTeam.isPending || betaFull}>
            <Plus className="h-3.5 w-3.5 mr-1" /> Новая команда
          </Button>
        </CardHeader>
        <CardContent>
          {teamsList.length === 0 ? (
            <p className="text-sm text-muted-foreground">У тебя пока нет команд. Создай новую или попроси приглашение.</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {teamsList.map((t) => (
                <button
                  key={t.id}
                  onClick={() => switchTeam(t.id)}
                  className={`rounded-md border px-3 py-1.5 text-sm transition-colors ${
                    activeTeam === t.id
                      ? "bg-primary text-primary-foreground border-primary"
                      : "bg-card hover:bg-secondary"
                  }`}
                >
                  {t.name}
                  <span className="ml-2 font-mono text-[10px] uppercase tracking-widest2 opacity-70">
                    {translateRole(t.role)}
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
    </div>
  );
}
