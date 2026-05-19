import { useTranslation } from "react-i18next";
import { Button, toast } from "@eop/ui";
import { Upload } from "lucide-react";

export function DashboardExportEventsButton() {
  const { t } = useTranslation("app");

  return (
    <Button
      type="button"
      variant="outline"
      onClick={() => toast.info(t("dashboard.export_not_available"))}
      className="inline-flex h-10 items-center gap-2 rounded-lg border px-3 text-[13px]"
      style={{ borderColor: "hsl(var(--eop-line-strong))" }}
    >
      <Upload className="h-3.5 w-3.5" />
      {t("dashboard.page_head_action_export")}
    </Button>
  );
}
