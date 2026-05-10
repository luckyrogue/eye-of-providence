import { useTranslation } from "react-i18next";
import { Badge } from "@eop/ui";
import { CATEGORY_TONES } from "../model/categories";

export function CategoryBadge({ cat }: { cat: string }) {
  const { t } = useTranslation("app");
  return (
    <Badge tone={CATEGORY_TONES[cat] ?? "neutral"}>
      {t(`dashboard.category.${cat}` as const, { defaultValue: cat })}
    </Badge>
  );
}
