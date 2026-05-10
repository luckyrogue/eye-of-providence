import { FolderGit2, GitCommit, Settings, Users } from "lucide-react";
import type { LucideIcon } from "lucide-react";

// TEAM_DETAIL_TABS — список табов на детали команды. owner-only фильтруется
// в page-shell. id используется как `TabKey`, i18nKey/icon — для рендера.
export const TEAM_DETAIL_TABS = [
  { id: "members", i18nKey: "team_detail.tabs.members", icon: Users },
  { id: "projects", i18nKey: "team_detail.tabs.projects", icon: FolderGit2 },
  { id: "commits", i18nKey: "team_detail.tabs.commits", icon: GitCommit },
  { id: "settings", i18nKey: "team_detail.tabs.settings", icon: Settings, ownerOnly: true },
] as const satisfies ReadonlyArray<{
  id: string;
  i18nKey: string;
  icon: LucideIcon;
  ownerOnly?: boolean;
}>;

export type TeamDetailTabKey = (typeof TEAM_DETAIL_TABS)[number]["id"];
