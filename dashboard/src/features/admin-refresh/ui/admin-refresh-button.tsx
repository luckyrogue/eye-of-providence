import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Button, toast } from "@eop/ui";
import { Loader2, RefreshCw } from "lucide-react";
import { adminKeys } from "@/entities/admin";

export function AdminRefreshButton({ isFetching }: { isFetching: boolean }) {
  const { t } = useTranslation("app");
  const qc = useQueryClient();

  const refresh = async () => {
    try {
      await qc.invalidateQueries({ queryKey: adminKeys.all });
      toast.success(t("admin.refreshed") || "Refreshed");
    } catch {
      toast.error(t("admin.refresh_failed") || "Refresh failed");
    }
  };

  return (
    <Button
      type="button"
      variant="outline"
      size="icon"
      onClick={() => void refresh()}
      disabled={isFetching}
      aria-label={t("admin.refresh") || "Refresh"}
      className="h-10 w-10"
    >
      {isFetching ? (
        <Loader2 className="h-4 w-4 animate-spin" />
      ) : (
        <RefreshCw className="h-4 w-4" />
      )}
    </Button>
  );
}
