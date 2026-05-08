import { http } from "../../lib/http";
import type { GenerateReportRes, ListReportsRes } from "./res";
import type { ReportPeriod } from "./types";

export type * from "./types";
export type * from "./res";

export const generateReport = (period: ReportPeriod = "weekly") =>
  http.post<GenerateReportRes>(`/v1/reports/generate`, undefined, { params: { period } }).then((r) => r.data);

export const listReports = () =>
  http.get<ListReportsRes>("/v1/reports/").then((r) => r.data.reports ?? []);
