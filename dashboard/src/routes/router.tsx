import { createBrowserRouter, Navigate } from "react-router-dom";
import { Landing } from "../Landing";
import { AppLayout } from "./AppLayout";
import { AuthRoute } from "./AuthRoute";
import { OnboardingRoute } from "./OnboardingRoute";
import { DashboardRoute } from "./Dashboard";
import { TeamRoute } from "./TeamRoute";
import { SettingsRoute } from "./SettingsRoute";
import { AdminRoute } from "./AdminRoute";

export const router = createBrowserRouter([
  { path: "/", element: <Landing /> },
  { path: "/landing", element: <Navigate to="/" replace /> },
  { path: "/login", element: <AuthRoute mode="login" /> },
  { path: "/signup", element: <AuthRoute mode="register" /> },
  { path: "/onboarding", element: <OnboardingRoute /> },
  {
    element: <AppLayout />,
    children: [
      { path: "/dashboard", element: <DashboardRoute /> },
      { path: "/team", element: <TeamRoute /> },
      { path: "/settings", element: <SettingsRoute /> },
      { path: "/admin", element: <AdminRoute /> },
    ],
  },
  { path: "*", element: <Navigate to="/" replace /> },
]);
