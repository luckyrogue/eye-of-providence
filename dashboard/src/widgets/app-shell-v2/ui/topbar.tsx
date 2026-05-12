// Topbar — фокус на работающих контролах. Убрано:
//   - Search ⌘K          (command palette не реализован)
//   - Date picker         (range не подсоединён к widgets — каждый виджет
//                          использует свой fixed days; вернём когда будет
//                          shared date-range context)
//   - Bell notifications  (notification system не реализован)
//   - Pause/Tracking      (паузить tracking — это agent-side фича в Tauri,
//                          в web dashboard паузить нечего)
//
// Остаются: mobile menu toggle, breadcrumb (display-only), LangSwitch, Logout.

import { useTranslation } from "react-i18next";
import { Menu } from "lucide-react";
import { LangSwitch } from "../../../shared/ui/lang-switch";

export function Topbar({
  crumb,
  onMenuClick,
  onLogout,
}: {
  crumb: { section: string; now: string };
  onMenuClick?: () => void;
  onLogout: () => void;
}) {
  const { t } = useTranslation("common");
  return (
    <header className="eop-topbar">
      {/* eslint-disable-next-line no-restricted-syntax -- icon-only mobile menu toggle */}
      <button
        type="button"
        onClick={onMenuClick}
        className="md:hidden grid place-items-center w-9 h-9 rounded-lg border"
        style={{ borderColor: "hsl(var(--border))" }}
        aria-label={t("topbar.menu_toggle")}
      >
        <Menu className="h-5 w-5" />
      </button>

      <div className="crumb hidden sm:flex">
        <span>{crumb.section}</span>
        {crumb.now && (
          <>
            <span className="sep">/</span>
            <span className="now">{crumb.now}</span>
          </>
        )}
      </div>

      <div className="ml-auto flex items-center gap-2.5">
        <LangSwitch />

        {/* eslint-disable-next-line no-restricted-syntax -- text-only link-style logout */}
        <button
          type="button"
          onClick={onLogout}
          className="text-sm text-muted-foreground hover:text-foreground transition-colors px-2"
        >
          {t("actions.logout")}
        </button>
      </div>
    </header>
  );
}
