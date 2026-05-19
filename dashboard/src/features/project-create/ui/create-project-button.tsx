import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, PromptDialog } from "@eop/ui";
import { Plus } from "lucide-react";
import { useCreateProject } from "@/entities/team";
import { useMutationToast } from "@/shared/hooks/use-mutation-toast";

type Stage = "closed" | "name" | "repo";

export function CreateProjectButton({ teamID }: { teamID: string }) {
  const { t } = useTranslation(["app", "common"]);
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

  return (
    <>
      <Button size="sm" onClick={() => setStage("name")} disabled={create.isPending}>
        <Plus className="h-3.5 w-3.5 mr-1" /> {t("app:team_detail.projects_new")}
      </Button>

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
    </>
  );
}
