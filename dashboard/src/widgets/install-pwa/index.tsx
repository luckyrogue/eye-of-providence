// Mobile install prompt — два режима:
//   1. Chromium (Android / Desktop): beforeinstallprompt → button "Install app"
//   2. iOS Safari: hint card "Share → Add to Home Screen"
//
// Скрывается после dismiss (localStorage flag) или после фактической установки
// (display-mode: standalone).

import { useTranslation } from "react-i18next";
import { Button, Card, CardContent } from "@eop/ui";
import { Download, X } from "lucide-react";
import { useInstallPrompt } from "../../shared/lib/pwa";

export function InstallPWA() {
  const { t } = useTranslation("pwa");
  const { canInstall, showIOSHint, install, dismiss } = useInstallPrompt();

  if (!canInstall && !showIOSHint) return null;

  return (
    <Card className="border-primary/20 bg-primary/5">
      <CardContent className="pt-6 flex items-start gap-3">
        <Download className="h-5 w-5 text-primary mt-0.5 shrink-0" />
        <div className="min-w-0 flex-1">
          <h4 className="font-medium text-sm">{t("title")}</h4>
          <p className="text-xs text-muted-foreground mt-1">
            {showIOSHint ? t("ios_hint") : t("hint")}
          </p>
          {canInstall && (
            <Button size="sm" className="mt-3" onClick={() => void install()}>
              <Download className="h-3.5 w-3.5 mr-1" />
              {t("install")}
            </Button>
          )}
        </div>
        <button
          type="button"
          onClick={dismiss}
          aria-label={t("dismiss")}
          className="text-muted-foreground hover:text-foreground transition-colors p-1"
        >
          <X className="h-4 w-4" />
        </button>
      </CardContent>
    </Card>
  );
}
