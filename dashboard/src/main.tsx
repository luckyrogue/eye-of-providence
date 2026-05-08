import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import "@eop/ui/styles.css";
import { ErrorBoundary } from "./ErrorBoundary";
import { Landing } from "./Landing";
import { queryClient } from "./queryClient";
import { AppLayout } from "./routes/AppLayout";
import { AuthRoute } from "./routes/AuthRoute";
import { OnboardingRoute } from "./routes/OnboardingRoute";
import { DashboardRoute } from "./routes/Dashboard";
import { TeamRoute } from "./routes/TeamRoute";
import { SettingsRoute } from "./routes/SettingsRoute";
import { AdminRoute } from "./routes/AdminRoute";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <Toaster position="bottom-right" toastOptions={{ className: "font-sans" }} />
      <BrowserRouter>
        <ErrorBoundary>
          <Routes>
            {/* Public */}
            <Route path="/" element={<Landing />} />
            <Route path="/landing" element={<Navigate to="/" replace />} />
            <Route path="/login" element={<AuthRoute mode="login" />} />
            <Route path="/signup" element={<AuthRoute mode="register" />} />
            <Route path="/onboarding" element={<OnboardingRoute />} />

            {/* App layout — sticky header + nav */}
            <Route element={<AppLayout />}>
              <Route path="/dashboard" element={<DashboardRoute />} />
              <Route path="/team" element={<TeamRoute />} />
              <Route path="/settings" element={<SettingsRoute />} />
              <Route path="/admin" element={<AdminRoute />} />
            </Route>

            {/* 404 → главная */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </ErrorBoundary>
      </BrowserRouter>
    </QueryClientProvider>
  </React.StrictMode>,
);
