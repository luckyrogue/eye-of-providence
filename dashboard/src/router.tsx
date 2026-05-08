// Конфигурация роутера через createBrowserRouter (data router API).
// Преимущества:
// - роуты заданы декларативно как объекты — легче статически анализировать
// - готовы к loader/action API (потенциально заменим useEffect-based redirects)
// - lazy() для code-splitting на route уровне (включим когда bundle станет больше)

import { createBrowserRouter, Navigate } from "react-router-dom";
import { Landing } from "./Landing";
import { AppLayout } from "./routes/AppLayout";
import { AuthRoute } from "./routes/AuthRoute";
import { OnboardingRoute } from "./routes/OnboardingRoute";
import { DashboardRoute } from "./routes/Dashboard";
import { TeamRoute } from "./routes/TeamRoute";
import { SettingsRoute } from "./routes/SettingsRoute";
import { AdminRoute } from "./routes/AdminRoute";

export const router = createBrowserRouter([
  // Public
  { path: "/", element: <Landing /> },
  { path: "/landing", element: <Navigate to="/" replace /> },
  { path: "/login", element: <AuthRoute mode="login" /> },
  { path: "/signup", element: <AuthRoute mode="register" /> },
  { path: "/onboarding", element: <OnboardingRoute /> },

  // App layout — sticky header + nav, защищённые routes внутри.
  {
    element: <AppLayout />,
    children: [
      { path: "/dashboard", element: <DashboardRoute /> },
      { path: "/team", element: <TeamRoute /> },
      { path: "/settings", element: <SettingsRoute /> },
      { path: "/admin", element: <AdminRoute /> },
    ],
  },

  // 404 → главная.
  { path: "*", element: <Navigate to="/" replace /> },
]);
