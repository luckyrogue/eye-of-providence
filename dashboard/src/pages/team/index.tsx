import { Teams } from "./ui/teams";
import { getTz } from "@/shared/lib/tz";

export function TeamRoute() {
  return <Teams tz={getTz()} />;
}
