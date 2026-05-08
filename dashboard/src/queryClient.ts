import { QueryClient } from "@tanstack/react-query";

// Дефолты под наш use-case:
// - 30s staleTime: дашборд данные не нужно дёргать каждый рендер,
//   но и хочется свежесть после CRUD операций.
// - retry один раз — больше смысла мало (если 401 — перенаправим, если 500 —
//   тут уж не помочь).
// - refetchOnWindowFocus: false — для SaaS'а лишний loader при tab-switch
//   раздражает; явные mutate'ы инвалидируют queries.
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
});
