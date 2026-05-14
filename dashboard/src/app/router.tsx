import { lazy, Suspense, type ReactNode } from "react";
import { createBrowserRouter, Navigate } from "react-router-dom";
import { AppShellV2 } from "../widgets/app-shell-v2";
import { RouteFallback } from "./route-fallback";
import { NotFound } from "../pages/not-found";
import { RouteError } from "../pages/route-error";
const Landing = lazy(() => import("../pages/landing").then((m) => ({ default: m.Landing })));
const AuthRoute = lazy(() => import("../pages/auth").then((m) => ({ default: m.AuthRoute })));
const ForgotPasswordRoute = lazy(() =>
  import("../pages/password-reset").then((m) => ({ default: m.ForgotPasswordRoute })),
);
const ResetPasswordRoute = lazy(() =>
  import("../pages/password-reset").then((m) => ({ default: m.ResetPasswordRoute })),
);
const AuthCompleteRoute = lazy(() =>
  import("../pages/auth-complete").then((m) => ({ default: m.AuthCompleteRoute })),
);
const OnboardingRoute = lazy(() =>
  import("../pages/onboarding").then((m) => ({ default: m.OnboardingRoute })),
);
const ChangelogRoute = lazy(() =>
  import("../pages/changelog").then((m) => ({ default: m.ChangelogRoute })),
);
const PricingRoute = lazy(() =>
  import("../pages/pricing").then((m) => ({ default: m.PricingRoute })),
);
const DashboardRoute = lazy(() =>
  import("../pages/dashboard").then((m) => ({ default: m.DashboardRoute })),
);
const TeamRoute = lazy(() => import("../pages/team").then((m) => ({ default: m.TeamRoute })));
const SettingsRoute = lazy(() =>
  import("../pages/settings").then((m) => ({ default: m.SettingsRoute })),
);
const IntegrationsRoute = lazy(() =>
  import("../pages/integrations").then((m) => ({ default: m.IntegrationsRoute })),
);
const AdminRoute = lazy(() => import("../pages/admin").then((m) => ({ default: m.AdminRoute })));
const wrap = (node: ReactNode) => <Suspense fallback={<RouteFallback />}>{node}</Suspense>;
const eb = { errorElement: <RouteError /> };
export const router = createBrowserRouter([
  { path: "/", element: wrap(<Landing />), ...eb },
  { path: "/landing", element: <Navigate to="/" replace />, ...eb },
  { path: "/login", element: wrap(<AuthRoute mode="login" />), ...eb },
  { path: "/signup", element: wrap(<AuthRoute mode="register" />), ...eb },
  { path: "/forgot-password", element: wrap(<ForgotPasswordRoute />), ...eb },
  { path: "/reset-password", element: wrap(<ResetPasswordRoute />), ...eb },
  { path: "/auth/complete", element: wrap(<AuthCompleteRoute />), ...eb },
  { path: "/onboarding", element: wrap(<OnboardingRoute />), ...eb },
  { path: "/changelog", element: wrap(<ChangelogRoute />), ...eb },
  { path: "/pricing", element: wrap(<PricingRoute />), ...eb },
  {
    element: <AppShellV2 />,
    ...eb,
    children: [
      { path: "/dashboard", element: wrap(<DashboardRoute />), ...eb },
      { path: "/team", element: wrap(<TeamRoute />), ...eb },
      { path: "/integrations", element: wrap(<IntegrationsRoute />), ...eb },
      { path: "/settings", element: wrap(<SettingsRoute />), ...eb },
      { path: "/admin", element: wrap(<AdminRoute />), ...eb },
    ],
  },
  { path: "*", element: <NotFound /> },
]);
