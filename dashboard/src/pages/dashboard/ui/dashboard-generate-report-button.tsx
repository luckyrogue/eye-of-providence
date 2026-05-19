import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { Loader2, Sparkles } from "lucide-react";
import { useGenerateReport } from "../api";
import { useMutationToast } from "@/shared/hooks";

export function DashboardGenerateReportButton() {
  const { t } = useTranslation("app");
  const generate = useGenerateReport();
  const runToast = useMutationToast();

  async function handleGenerate() {
    await runToast(generate.mutateAsync("weekly"), {
      success: t("dashboard.report_generated_toast"),
      error: t("dashboard.report_failed_toast"),
    });
  }

  return (
    <Button
      type="button"
      onClick={handleGenerate}
      disabled={generate.isPending}
      className="btn-eop-primary inline-flex h-10 items-center gap-2 rounded-lg px-3 text-[13px] font-medium"
    >
      {generate.isPending ? (
        <Loader2 className="h-3.5 w-3.5 animate-spin" />
      ) : (
        <Sparkles className="h-3.5 w-3.5" />
      )}
      {t("dashboard.page_head_action_report")}
    </Button>
  );
}
