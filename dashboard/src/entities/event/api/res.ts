import type { EventRow, HeatmapCell, LangCell, TrendPoint } from "./types";

type RecentRes = { events: EventRow[] };
type SummaryRes = { categories: Record<string, number> };
type ProvenanceRes = { provenance: Record<string, number> };
type LanguagesRes = { cells: LangCell[] };
type HeatmapRes = { cells: HeatmapCell[] };
type TrendRes = { points: TrendPoint[] };
type IngestRes = { accepted: number; rejected: number };

export type { RecentRes, SummaryRes, ProvenanceRes, LanguagesRes, HeatmapRes, TrendRes, IngestRes };
