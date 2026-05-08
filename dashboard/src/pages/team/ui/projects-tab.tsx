import { useState } from "react";
import { Button, EmptyState, PromptDialog } from "@eop/ui";
import { Plus } from "lucide-react";
import { useProjects, useCreateProject } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { formatDate } from "../../../shared/lib/tz";

type Stage = "closed" | "name" | "repo";

export function ProjectsTab({ teamID, role, tz }: { teamID: string; role: string; tz: string }) {
  const projects = useProjects(teamID);
  const create = useCreateProject(teamID);
  const runToast = useMutationToast();
  const [stage, setStage] = useState<Stage>("closed");
  const [name, setName] = useState("");

  function close() {
    setStage("closed");
    setName("");
  }

  async function commit(repoURL: string) {
    const r = await runToast(create.mutateAsync({ name, repoURL }), {
      success: "Проект создан",
      error: "Не удалось создать проект",
    });
    if (r) close();
  }

  const list = projects.data ?? [];

  return (
    <div className="space-y-3">
      {(role === "owner" || role === "admin") && (
        <Button size="sm" onClick={() => setStage("name")} disabled={create.isPending}>
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

      <PromptDialog
        open={stage === "name"}
        title="Новый проект"
        label="Название"
        placeholder="frontend, api, infra…"
        confirmText="Дальше"
        onClose={close}
        onConfirm={(v) => {
          setName(v);
          setStage("repo");
        }}
      />
      <PromptDialog
        open={stage === "repo"}
        title="Repo URL"
        description="Опционально — можно оставить пустым и заполнить позже."
        label="URL"
        placeholder="https://github.com/acme/frontend"
        confirmText="Создать"
        busy={create.isPending}
        onClose={close}
        onConfirm={commit}
      />
    </div>
  );
}
