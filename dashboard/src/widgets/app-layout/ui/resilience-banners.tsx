import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { useOnlineStatus, useUpdatePrompt } from "../../../shared/lib/pwa";
export function ResilienceBanners() {
  const { t } = useTranslation("common");
  const online = useOnlineStatus();
  const update = useUpdatePrompt();
  if (online && !update.available) return null;
  return (
    <div className="sticky top-0 z-40 space-y-1 px-4 md:px-6 py-2">
      {!online && (
        <div className="rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-1.5 text-xs text-amber-900 dark:text-amber-100">
          {t("offline.banner")}
        </div>
      )}
      {update.available && (
        <div className="flex items-center justify-between gap-2 rounded-md border border-primary/40 bg-primary/10 px-3 py-1.5 text-xs">
          <span>{t("update.available")}</span>
          <Button size="sm" variant="default" onClick={update.reload}>
            {t("update.reload")}
          </Button>
        </div>
      )}
    </div>
  );
}
