import { useTranslation } from "react-i18next";
import { DangerZone } from "@eop/ui";
import type { Team } from "../../../entities/team";
import { DeleteTeamButton } from "../../../features/team-delete";
import { RenameTeamForm } from "../../../features/team-rename";
export function SettingsTab({ team }: { team: Team }) {
  const { t } = useTranslation("app");
  return (
    <div className="space-y-6">
      <RenameTeamForm team={team} />

      <DangerZone
        title={t("team_detail.settings_danger_title")}
        description={t("team_detail.settings_danger_lead")}
        action={<DeleteTeamButton team={team} />}
      />
    </div>
  );
}
