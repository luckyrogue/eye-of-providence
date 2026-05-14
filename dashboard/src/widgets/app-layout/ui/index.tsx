import { Outlet } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useMe } from "../../../entities/user";
import { useAuthRedirect } from "../lib/use-auth-redirect";
import { BottomNav } from "./bottom-nav";
import { HeaderNav } from "./header-nav";
import { ResilienceBanners } from "./resilience-banners";
export function AppLayout() {
  const { doLogout } = useAuthRedirect();
  const { t } = useTranslation("common");
  const me = useMe();
  const isSuperAdmin = me.data?.global_role === "super_admin";
  return (
    <div
      className="min-h-screen bg-gradient-to-br from-background via-background to-primary/5"
      style={{
        paddingBottom: "calc(4rem + env(safe-area-inset-bottom))",
      }}
    >
      <a href="#main" className="skip-link">
        {t("a11y.skip_to_content")}
      </a>

      <HeaderNav isSuperAdmin={isSuperAdmin} onLogout={doLogout} />
      <ResilienceBanners />

      <main
        id="main"
        className="mx-auto max-w-6xl px-4 md:px-6 py-4 md:py-8 space-y-4 md:space-y-6"
      >
        <Outlet />
      </main>

      <BottomNav isSuperAdmin={isSuperAdmin} />
    </div>
  );
}
