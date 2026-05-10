import { getTz } from "../../shared/lib/tz";
import { Dashboard } from "./dashboard";

export function DashboardRoute() {
  const tz = getTz();
  return <Dashboard tz={tz} />;
}
