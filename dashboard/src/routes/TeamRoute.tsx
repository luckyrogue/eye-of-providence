import { Teams } from "../Teams";
import { getTz } from "../tz";

export function TeamRoute() {
  return <Teams tz={getTz()} />;
}
