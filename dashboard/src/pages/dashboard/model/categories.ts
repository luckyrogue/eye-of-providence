import type { BadgeVariant } from "@eop/ui";

// Маппинг категории события на цветовой variant <Badge>.
// Backend категории: manual / ai / refactor / idle / other / reading.
// Категории, не описанные явно — дефолт "default".

export const CATEGORY_VARIANTS: Record<string, BadgeVariant> = {
  manual: "blue",
  ai: "purple",
  refactor: "amber",
  idle: "default",
  other: "default",
  reading: "default",
};
