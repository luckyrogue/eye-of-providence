import { StatTile } from "@eop/ui";
import { Building2, Crown, Users } from "lucide-react";
import type { AdminStats } from "../api/admin";

export function Overview({ stats }: { stats: AdminStats }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile label="Пользователей" value={stats.users_total} icon={<Users className="h-4 w-4" />} />
      <StatTile
        label="Компаний"
        value={stats.teams_total}
        hint={stats.beta_limit > 0 ? `beta limit · ${stats.beta_limit}` : "лимит снят"}
        icon={<Building2 className="h-4 w-4" />}
      />
      <StatTile label="Membership-связей" value={stats.members_total} icon={<Crown className="h-4 w-4" />} />
    </div>
  );
}
