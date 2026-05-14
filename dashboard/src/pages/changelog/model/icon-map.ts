import { Sparkles, Wrench, Zap, BookOpen, Hammer } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { ChangelogType } from "./types";
export const TYPE_ICON: Record<
  ChangelogType,
  {
    icon: LucideIcon;
    color: string;
  }
> = {
  feat: { icon: Sparkles, color: "text-purple-500" },
  fix: { icon: Wrench, color: "text-amber-500" },
  perf: { icon: Zap, color: "text-emerald-500" },
  refactor: { icon: Hammer, color: "text-blue-500" },
  docs: { icon: BookOpen, color: "text-muted-foreground" },
};
