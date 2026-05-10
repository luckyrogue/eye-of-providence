import { lazy, Suspense, type ReactNode } from "react";
import { createBrowserRouter, Navigate } from "react-router-dom";
import { Skeleton } from "@eop/ui";
import { AppLayout } from "../widgets/app-layout";

const Landing = lazy(() => import("../pages/landing").then((m) => ({ default: m.Landing })));
const AuthRoute = lazy(() => import("../pages/auth").then((m) => ({ default: m.AuthRoute })));
const ForgotPasswordRoute = lazy(() =>
  import("../pages/password-reset").then((m) => ({ default: m.ForgotPasswordRoute })),
);
const ResetPasswordRoute = lazy(() =>
  import("../pages/password-reset").then((m) => ({ default: m.ResetPasswordRoute })),
);
const OnboardingRoute = lazy(() =>
  import("../pages/onboarding").then((m) => ({ default: m.OnboardingRoute })),
);
const ChangelogRoute = lazy(() =>
  import("../pages/changelog").then((m) => ({ default: m.ChangelogRoute })),
);
const DashboardRoute = lazy(() =>
  import("../pages/dashboard").then((m) => ({ default: m.DashboardRoute })),
);
const TeamRoute = lazy(() => import("../pages/team").then((m) => ({ default: m.TeamRoute })));
const SettingsRoute = lazy(() =>
  import("../pages/settings").then((m) => ({ default: m.SettingsRoute })),
);
const AdminRoute = lazy(() => import("../pages/admin").then((m) => ({ default: m.AdminRoute })));

function RouteFallback() {
  return (
    <div className="mx-auto max-w-6xl px-6 py-12 space-y-4">
      <Skeleton className="h-10 w-48" />
      <Skeleton className="h-32 w-full" />
      <Skeleton className="h-32 w-full" />
    </div>
  );
}

const wrap = (node: ReactNode) => <Suspense fallback={<RouteFallback />}>{node}</Suspense>;

export const router = createBrowserRouter([
  { path: "/", element: wrap(<Landing />) },
  { path: "/landing", element: <Navigate to="/" replace /> },
  { path: "/login", element: wrap(<AuthRoute mode="login" />) },
  { path: "/signup", element: wrap(<AuthRoute mode="register" />) },
  { path: "/forgot-password", element: wrap(<ForgotPasswordRoute />) },
  { path: "/reset-password", element: wrap(<ResetPasswordRoute />) },
  { path: "/onboarding", element: wrap(<OnboardingRoute />) },
  { path: "/changelog", element: wrap(<ChangelogRoute />) },
  {
    element: <AppLayout />,
    children: [
      { path: "/dashboard", element: wrap(<DashboardRoute />) },
      { path: "/team", element: wrap(<TeamRoute />) },
      { path: "/settings", element: wrap(<SettingsRoute />) },
      { path: "/admin", element: wrap(<AdminRoute />) },
    ],
  },
  { path: "*", element: <Navigate to="/" replace /> },
]);
