import { Button, EmptyState } from "@eop/ui";
import { Plus } from "lucide-react";
import { useProjects, useCreateProject } from "../api/teams";
import { useMutationToast } from "../hooks/useMutationToast";
import { formatDate } from "../utils/tz";

export function ProjectsTab({ teamID, role, tz }: { teamID: string; role: string; tz: string }) {
  const projects = useProjects(teamID);
  const create = useCreateProject(teamID);
  const runToast = useMutationToast();

  async function add() {
    const name = prompt("Название проекта");
    if (!name?.trim()) return;
    const repoURL = prompt("Repo URL (опционально)") || "";
    await runToast(create.mutateAsync({ name: name.trim(), repoURL }), {
      success: "Проект создан",
      error: "Не удалось создать проект",
    });
  }

  const list = projects.data ?? [];

  return (
    <div className="space-y-3">
      {(role === "owner" || role === "admin") && (
        <Button size="sm" onClick={add} disabled={create.isPending}>
          <Plus className="h-3.5 w-3.5 mr-1" /> Новый проект
        </Button>
      )}
      {list.length === 0 ? (
        <EmptyState
          eyebrow="No projects"
          title="Создай первый проект"
          description="Проекты привязывают коммиты и метрики к репозиториям. После создания установи git post-commit hook."
        />
      ) : (
        <ul className="space-y-2">
          {list.map((p) => (
            <li key={p.id} className="rounded-md border p-3">
              <div className="font-medium text-sm">{p.name}</div>
              {p.repo_url && <div className="text-xs text-muted-foreground font-mono">{p.repo_url}</div>}
              <div className="text-xs text-muted-foreground mt-1">создан {formatDate(p.created_at, tz)}</div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
