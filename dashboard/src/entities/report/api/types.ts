type Report = {
  id: string;
  period: string;
  model: string;
  body_md: string;
  generated_at: string;
  prompt_version: string;
};
type ReportPeriod = "weekly" | "monthly" | "daily";
export type { Report, ReportPeriod };
