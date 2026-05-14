import { useTranslation } from "react-i18next";
import { Menu } from "lucide-react";
import { Button } from "@eop/ui";
import { LangSwitch } from "../../../shared/ui/lang-switch";
export function Topbar({
  crumb,
  onMenuClick,
  onLogout,
}: {
  crumb: {
    section: string;
    now: string;
  };
  onMenuClick?: () => void;
  onLogout: () => void;
}) {
  const { t } = useTranslation("common");
  return (
    <header className="eop-topbar">
      <Button
        type="button"
        variant="outline"
        size="icon"
        onClick={onMenuClick}
        className="md:hidden"
        style={{ borderColor: "hsl(var(--border))" }}
        aria-label={t("topbar.menu_toggle")}
      >
        <Menu className="h-5 w-5" />
      </Button>

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

        <Button
          type="button"
          variant="link"
          onClick={onLogout}
          className="h-auto px-2 py-0 text-sm text-muted-foreground hover:text-foreground"
        >
          {t("actions.logout")}
        </Button>
      </div>
    </header>
  );
}
