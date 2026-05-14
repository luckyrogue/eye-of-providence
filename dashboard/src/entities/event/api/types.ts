type EventRow = {
  ts: string;
  user_id: string;
  app_bundle: string;
  category: string;
  source: string;
  ai_provider?: string;
  ai_channel?: string;
  duration_ms: number;
  chars_in: number;
};
type LangCell = {
  lang: string;
  category: string;
  chars: number;
  ms: number;
};
type HeatmapCell = {
  dow: number;
  hour: number;
  category: string;
  ms: number;
};
type TrendPoint = {
  date: string;
  category: string;
  chars: number;
  ms: number;
};
export type { EventRow, LangCell, HeatmapCell, TrendPoint };
