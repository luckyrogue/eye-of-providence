import { useTranslation } from "react-i18next";
import { DangerZone } from "@eop/ui";
import { DeleteMyDataButton } from "./delete-my-data-button";
import { ExportMyDataButton } from "./export-my-data-button";

export function SettingsDangerZone({ onWiped }: { onWiped: () => void }) {
  const { t } = useTranslation("common");
  // Export рядом с Delete — это пара GDPR-операций (Article 20 + Article 17).
  // Юзеру в момент удаления могут понадобиться его данные «на руки» — поэтому
  // оба кнопки в одной DangerZone, не разнесены по разным карточкам.
  return (
    <DangerZone
      title={t("settings.danger_title")}
      description={t("settings.danger_lead")}
      action={
        <div className="flex flex-col sm:flex-row gap-2 sm:items-center">
          <ExportMyDataButton />
          <DeleteMyDataButton onWiped={onWiped} />
        </div>
      }
    />
  );
}
