import { QueryCache, QueryClient } from "@tanstack/react-query";
import { toast } from "@eop/ui";

// httpErrorStatus — частая форма HTTP-ошибки (axios / fetch wrapper):
// у объекта может быть .status / .response.status / .code.
export function httpErrorStatus(err: unknown): number | undefined {
  if (typeof err !== "object" || err === null) return undefined;
  const e = err as { status?: number; response?: { status?: number } };
  return e.status ?? e.response?.status;
}

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
  // Global onError: всплываем тостом только для ошибок, не относящихся к
  // auth (401/403 редиректит middleware) и не относящихся к network/offline
  // (отдельный offline-banner это покрывает). Background refetch не должен
  // спамить тостами.
  queryCache: new QueryCache({
    onError: (error, query) => {
      const status = httpErrorStatus(error);
      if (status === 401 || status === 403) return;
      // 5xx на /me → сессия сбрасывается в useAuthRedirect; тост не нужен.
      if (status != null && status >= 500 && query.queryKey[0] === "me") return;
      if (!navigator.onLine) return;
      // Игнорируем background refetch — initial load уже показал error UI.
      if (query.state.data !== undefined) return;
      const msg = error instanceof Error ? error.message : String(error);
      toast.error(msg);
    },
  }),
});
