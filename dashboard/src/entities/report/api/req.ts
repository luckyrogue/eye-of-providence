import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";
import type { GenerateReportRes, ListReportsRes } from "./res";
import type { ReportPeriod } from "./types";

export const reportsKeys = {
  list: ["reports"] as const,
};

export const generateReport = (period: ReportPeriod = "weekly") =>
  http
    .post<GenerateReportRes>(`/v1/reports/generate`, undefined, { params: { period } })
    .then((r) => r.data);

export const listReports = () =>
  http.get<ListReportsRes>("/v1/reports/").then((r) => r.data.reports ?? []);

export const useReports = () => useQuery({ queryKey: reportsKeys.list, queryFn: listReports });

export function useGenerateReport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (period: ReportPeriod) => generateReport(period),
    onSuccess: () => qc.invalidateQueries({ queryKey: reportsKeys.list }),
  });
}
