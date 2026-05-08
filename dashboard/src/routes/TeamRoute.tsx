import { Teams } from "../Teams";
import { getTz } from "../utils/tz";

export function TeamRoute() {
  return <Teams tz={getTz()} />;
}
