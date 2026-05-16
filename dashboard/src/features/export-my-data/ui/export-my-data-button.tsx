import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { Download } from "lucide-react";
import { useExportMyData } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

// ExportMyDataButton — GDPR Article 20 "Right to data portability".
// Single-click → backend стримит JSON со всеми данными пользователя
// (profile + devices + projects + consent + reports + api_tokens (без
// hashed_token) + events history), браузер сохраняет файл локально.
//
// Это HTTP GET (idempotent) — без confirmation modal. Если юзер передумал,
// он просто не кликнет «save» в браузере или удалит файл.
export function ExportMyDataButton() {
  const { t } = useTranslation("common");
  const exportData = useExportMyData();
  const runToast = useMutationToast();

  async function run() {
    await runToast(exportData.mutateAsync(), {
      success: t("settings.export_success", "Data exported"),
      error: t("settings.export_failed", "Export failed"),
    });
  }

  return (
    <Button variant="outline" size="sm" onClick={run} disabled={exportData.isPending}>
      <Download className="h-4 w-4 mr-2" />
      {t("settings.export_btn", "Export my data")}
    </Button>
  );
}
