import { QueryCache, QueryClient } from "@tanstack/react-query";
import { toast } from "@eop/ui";
export function httpErrorStatus(err: unknown): number | undefined {
  if (typeof err !== "object" || err === null) return undefined;
  const e = err as {
    status?: number;
    response?: {
      status?: number;
    };
  };
  return e.status ?? e.response?.status;
}
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
  queryCache: new QueryCache({
    onError: (error, query) => {
      const status = httpErrorStatus(error);
      if (status === 401 || status === 403) return;
      if (status != null && status >= 500 && query.queryKey[0] === "me") return;
      if (!navigator.onLine) return;
      if (query.state.data !== undefined) return;
      const msg = error instanceof Error ? error.message : String(error);
      toast.error(msg);
    },
  }),
});
