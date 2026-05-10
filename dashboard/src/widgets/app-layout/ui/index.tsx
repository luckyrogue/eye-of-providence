// AppLayout — главный chrome для авторизованной части приложения.
// Bottom-nav на mobile (PWA-style) + top-nav на desktop. Содержимое
// заворачиваем в pb-16 чтобы bottom-nav не накрывал последний row.

import { Outlet } from "react-router-dom";
import { useMe } from "../../../entities/user";
import { useAuthRedirect } from "../lib/use-auth-redirect";
import { BottomNav } from "./bottom-nav";
import { HeaderNav } from "./header-nav";

export function AppLayout() {
  const { doLogout } = useAuthRedirect();
  const me = useMe();
  const isSuperAdmin = me.data?.global_role === "super_admin";

  return (
    <main
      className="min-h-screen bg-gradient-to-br from-background via-background to-primary/5"
      style={{
        paddingBottom: "calc(4rem + env(safe-area-inset-bottom))",
      }}
    >
      <HeaderNav isSuperAdmin={isSuperAdmin} onLogout={doLogout} />

      <div className="mx-auto max-w-6xl px-4 md:px-6 py-4 md:py-8 space-y-4 md:space-y-6">
        <Outlet />
      </div>

      <BottomNav isSuperAdmin={isSuperAdmin} />
    </main>
  );
}
