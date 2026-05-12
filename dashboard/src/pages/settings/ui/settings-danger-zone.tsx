import { useTranslation } from "react-i18next";
import { DangerZone } from "@eop/ui";
import { DeleteMyDataButton } from "../../../features/delete-my-data";

export function SettingsDangerZone({ onWiped }: { onWiped: () => void }) {
  const { t } = useTranslation("common");
  return (
    <DangerZone
      title={t("settings.danger_title")}
      description={t("settings.danger_lead")}
      action={<DeleteMyDataButton onWiped={onWiped} />}
    />
  );
}
