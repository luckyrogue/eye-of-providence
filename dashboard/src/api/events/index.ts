import { http } from "../../lib/http";
import type { IngestReq } from "./req";
import type {
  HeatmapRes,
  IngestRes,
  LanguagesRes,
  RecentRes,
  SummaryRes,
  TrendRes,
} from "./res";

export type * from "./types";
export type * from "./req";
export type * from "./res";

export const fetchRecent = (limit = 50) =>
  http.get<RecentRes>(`/v1/events/recent`, { params: { limit } }).then((r) => r.data.events ?? []);

export const fetchSummary = (days = 7) =>
  http
    .get<SummaryRes>(`/v1/summary/categories`, { params: { days } })
    .then((r) => r.data.categories ?? {});

export const fetchLanguages = (days = 30) =>
  http.get<LanguagesRes>(`/v1/summary/languages`, { params: { days } }).then((r) => r.data.cells ?? []);

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
