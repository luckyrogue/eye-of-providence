import { useTranslation } from "react-i18next";
import { Activity, Settings as SettingsIcon, Shield, Users } from "lucide-react";
import { BottomNavItem } from "./bottom-nav-item";
export function BottomNav({ isSuperAdmin }: { isSuperAdmin: boolean }) {
  const { t } = useTranslation("common");
  return (
    <nav
      className="md:hidden fixed bottom-0 inset-x-0 border-t bg-background/95 backdrop-blur-md z-20"
      style={{ paddingBottom: "env(safe-area-inset-bottom)" }}
      aria-label={t("nav.bottom")}
    >
      <div className="flex items-stretch">
        <BottomNavItem
          to="/dashboard"
          icon={<Activity className="h-5 w-5" />}
          label={t("nav.dashboard")}
        />
        <BottomNavItem to="/team" icon={<Users className="h-5 w-5" />} label={t("nav.team")} />
        <BottomNavItem
          to="/settings"
          icon={<SettingsIcon className="h-5 w-5" />}
          label={t("nav.settings")}
        />
        {isSuperAdmin && (
          <BottomNavItem
            to="/admin"
            icon={<Shield className="h-5 w-5" />}
            label={t("nav.admin")}
            accent
          />
        )}
      </div>
    </nav>
  );
}
