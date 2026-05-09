import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, EmptyState, PromptDialog } from "@eop/ui";
import { Plus } from "lucide-react";
import { useProjects, useCreateProject } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { formatDate } from "../../../shared/lib/tz";

type Stage = "closed" | "name" | "repo";

export function ProjectsTab({ teamID, role, tz }: { teamID: string; role: string; tz: string }) {
  const { t } = useTranslation(["app", "common"]);
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
      success: t("app:team_detail.projects_create_success"),
      error: t("app:team_detail.projects_create_failed"),
    });
    if (r) close();
  }

  const list = projects.data ?? [];

  return (
    <div className="space-y-3">
      {(role === "owner" || role === "admin") && (
        <Button size="sm" onClick={() => setStage("name")} disabled={create.isPending}>
          <Plus className="h-3.5 w-3.5 mr-1" /> {t("app:team_detail.projects_new")}
        </Button>
      )}
      {list.length === 0 ? (
        <EmptyState
          eyebrow={t("app:team_detail.projects_empty_eyebrow")}
          title={t("app:team_detail.projects_empty_title")}
          description={t("app:team_detail.projects_empty_lead")}
        />
      ) : (
        <ul className="space-y-2">
          {list.map((p) => (
            <li key={p.id} className="rounded-md border p-3">
              <div className="font-medium text-sm">{p.name}</div>
              {p.repo_url && <div className="text-xs text-muted-foreground font-mono">{p.repo_url}</div>}
              <div className="text-xs text-muted-foreground mt-1">
                {t("app:team_detail.projects_created", { date: formatDate(p.created_at, tz) })}
              </div>
            </li>
          ))}
        </ul>
      )}

      <PromptDialog
        open={stage === "name"}
        title={t("app:team_detail.projects_dialog_name_title")}
        label={t("app:team_detail.projects_dialog_name_label")}
        placeholder={t("app:team_detail.projects_dialog_name_placeholder")}
        confirmText={t("common:actions.continue")}
        onClose={close}
        onConfirm={(v) => {
          setName(v);
          setStage("repo");
        }}
      />
      <PromptDialog
        open={stage === "repo"}
        title={t("app:team_detail.projects_dialog_repo_title")}
        description={t("app:team_detail.projects_dialog_repo_lead")}
        label={t("app:team_detail.projects_dialog_repo_label")}
        placeholder="https://github.com/acme/frontend"
        confirmText={t("common:actions.create")}
        busy={create.isPending}
        onClose={close}
        onConfirm={commit}
      />
    </div>
  );
}
