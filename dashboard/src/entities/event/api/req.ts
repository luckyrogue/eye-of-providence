import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "@/shared/api";
import type { EventRow } from "./types";
import type { HeatmapRes, IngestRes, LanguagesRes, RecentRes, SummaryRes, TrendRes } from "./res";

export type IngestReq = { events: Partial<EventRow & { file_lang: string }>[] };

// Все query keys строятся от общего корня `all`, чтобы инвалидация
// одного ключа уровня сущности (`events`) очищала все её вариации.
export const eventsKeys = {
  all: ["events"] as const,
  recent: (limit: number) => [...eventsKeys.all, "recent", limit] as const,
  summary: (days: number) => [...eventsKeys.all, "summary", days] as const,
  languages: (days: number) => [...eventsKeys.all, "languages", days] as const,
  heatmap: (days: number, tz: string) => [...eventsKeys.all, "heatmap", days, tz] as const,
  trend: (days: number, tz: string) => [...eventsKeys.all, "trend", days, tz] as const,
  cost: () => [...eventsKeys.all, "cost"] as const,
};

export const fetchRecent = (limit = 50) =>
  http.get<RecentRes>(`/v1/events/recent`, { params: { limit } }).then((r) => r.data.events ?? []);

export const fetchSummary = (days = 7) =>
  http
    .get<SummaryRes>(`/v1/summary/categories`, { params: { days } })
    .then((r) => r.data.categories ?? {});

export const fetchLanguages = (days = 30) =>
  http
    .get<LanguagesRes>(`/v1/summary/languages`, { params: { days } })
    .then((r) => r.data.cells ?? []);

export const fetchHeatmap = (days = 30, tz?: string) =>
  http.get<HeatmapRes>(`/v1/heatmap`, { params: { days, tz } }).then((r) => r.data.cells ?? []);

export const fetchTrend = (days = 30, tz?: string) =>
  http.get<TrendRes>(`/v1/trend`, { params: { days, tz } }).then((r) => r.data.points ?? []);

export const fetchCost = () =>
  http.get<Record<string, number>>("/v1/admin/cost").then((r) => r.data);

export const ingestDemoEvent = () => {
  const body: IngestReq = {
    events: [
      {
        app_bundle: "com.microsoft.VSCode",
        category: "manual",
        source: "os",
        duration_ms: 30000,
        file_lang: "ts",
        chars_in: 120,
      },
      {
        app_bundle: "chatgpt.com",
        category: "ai",
        source: "browser",
        ai_provider: "openai",
        ai_channel: "chat",
        duration_ms: 60000,
      },
    ],
  };
  return http.post<IngestRes>("/v1/ingest", body).then((r) => r.data);
};

export const useRecent = (limit = 20) =>
  useQuery({ queryKey: eventsKeys.recent(limit), queryFn: () => fetchRecent(limit) });

export const useSummary = (days = 7) =>
  useQuery({ queryKey: eventsKeys.summary(days), queryFn: () => fetchSummary(days) });

export const useLanguages = (days = 30) =>
  useQuery({ queryKey: eventsKeys.languages(days), queryFn: () => fetchLanguages(days) });

export const useHeatmap = (days: number, tz: string) =>
  useQuery({ queryKey: eventsKeys.heatmap(days, tz), queryFn: () => fetchHeatmap(days, tz) });

export const useTrend = (days: number, tz: string) =>
  useQuery({ queryKey: eventsKeys.trend(days, tz), queryFn: () => fetchTrend(days, tz) });

export const useCost = () => useQuery({ queryKey: eventsKeys.cost(), queryFn: fetchCost });

export function useIngestDemo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ingestDemoEvent,
    // Один root-prefix invalidate чистит recent / summary / languages / heatmap / trend / cost.
    onSuccess: () => qc.invalidateQueries({ queryKey: eventsKeys.all }),
  });
}
