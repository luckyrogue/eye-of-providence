// Маппинг категории события на цветовой тон UI-компонента <Badge>.
// Backend категории: manual / ai / refactor / idle / other / reading.
// Категории, не описанные явно — дефолт "neutral".

export const CATEGORY_TONES: Record<string, "blue" | "purple" | "amber" | "neutral"> = {
  manual: "blue",
  ai: "purple",
  refactor: "amber",
  idle: "neutral",
  other: "neutral",
  reading: "neutral",
};
