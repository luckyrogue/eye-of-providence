import ReactDOM from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { ConfirmProvider, Toaster } from "@eop/ui";
import "@eop/ui/styles.css";
import "../shared/i18n";
import { ErrorBoundary } from "./error-boundary";
import { queryClient } from "../shared/config/query-client";
import { router } from "./router";
import { registerSW } from "../shared/lib/pwa";
registerSW();
ReactDOM.createRoot(document.getElementById("root")!).render(
  <QueryClientProvider client={queryClient}>
    <Toaster />
    <ConfirmProvider>
      <ErrorBoundary>
        <RouterProvider router={router} />
      </ErrorBoundary>
    </ConfirmProvider>
    {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
  </QueryClientProvider>,
);
