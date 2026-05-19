import { useTranslation } from "react-i18next";
import { EmptyState } from "@eop/ui";
import { useProjects } from "@/entities/team";
import { CreateProjectButton } from "@/features/project-create";
import { isReadOnlyRole } from "../lib/roles";
import { formatDate } from "@/shared/lib/tz";

export function ProjectsTab({ teamID, role, tz }: { teamID: string; role: string; tz: string }) {
  const { t } = useTranslation("app");
  const projects = useProjects(teamID);
  const list = projects.data ?? [];
  const canCreate = !isReadOnlyRole(role) && (role === "owner" || role === "admin");

  return (
    <div className="space-y-3">
      {canCreate && (
        <div className="flex justify-end">
          <CreateProjectButton teamID={teamID} />
        </div>
      )}
      {list.length === 0 ? (
        <EmptyState
          eyebrow={t("team_detail.projects_empty_eyebrow")}
          title={t("team_detail.projects_empty_title")}
          description={t("team_detail.projects_empty_lead")}
        />
      ) : (
        <ul className="space-y-2">
          {list.map((p) => (
            <li key={p.id} className="rounded-md border p-3">
              <div className="font-medium text-sm">{p.name}</div>
              {p.repo_url && (
                <div className="text-xs text-muted-foreground font-mono">{p.repo_url}</div>
              )}
              <div className="text-xs text-muted-foreground mt-1">
                {t("team_detail.projects_created", { date: formatDate(p.created_at, tz) })}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
