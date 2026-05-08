import React from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { Toaster } from "sonner";
import { ConfirmProvider } from "@eop/ui";
import "@eop/ui/styles.css";
import { ErrorBoundary } from "./ErrorBoundary";
import { queryClient } from "./queryClient";
import { router } from "./routes/router";

ReactDOM.createRoot(document.getElementById("root")!).render(
    <QueryClientProvider client={queryClient}>
      <Toaster position="bottom-right" toastOptions={{ className: "font-sans" }} />
      <ConfirmProvider>
        <ErrorBoundary>
          <RouterProvider router={router} />
        </ErrorBoundary>
      </ConfirmProvider>
      {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
    </QueryClientProvider>
);
