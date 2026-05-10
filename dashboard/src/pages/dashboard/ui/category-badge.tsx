import { useTranslation } from "react-i18next";
import { Badge } from "@eop/ui";
import { CATEGORY_VARIANTS } from "../model/categories";

export function CategoryBadge({ cat }: { cat: string }) {
  const { t } = useTranslation("app");
  return (
    <Badge variant={CATEGORY_VARIANTS[cat] ?? "default"}>
      {t(`dashboard.category.${cat}` as const, { defaultValue: cat })}
    </Badge>
  );
}
