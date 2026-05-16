import type { Report } from "./types";

type GenerateReportRes = Report;
type ListReportsRes = { reports: Report[] };

export type { GenerateReportRes, ListReportsRes };
