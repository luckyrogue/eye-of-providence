import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { CONTENT_SLUGS, type ContentSlug } from "../../../shared/content";
export function SlugTree({
  selected,
  onSelect,
}: {
  selected: ContentSlug | null;
  onSelect: (slug: ContentSlug) => void;
}) {
  const { t } = useTranslation("app");
  const groups = useMemo(() => {
    const map = new Map<string, ContentSlug[]>();
    for (const s of CONTENT_SLUGS) {
      const head = s.split(".")[0] ?? "other";
      const arr = map.get(head) ?? [];
      arr.push(s);
      map.set(head, arr);
    }
    return Array.from(map.entries());
  }, []);
  return (
    <div
      className="rounded-md border overflow-hidden text-sm"
      style={{ borderColor: "hsl(var(--border))" }}
    >
      <div
        className="px-3 py-2 text-[11px] uppercase tracking-wide font-mono text-muted-foreground bg-muted/30 border-b"
        style={{ borderColor: "hsl(var(--border))" }}
      >
        {t("admin.content.slug_picker_title", { defaultValue: "Content slugs" })}
      </div>
      {groups.map(([group, slugs]) => (
        <div
          key={group}
          className="border-b last:border-b-0"
          style={{ borderColor: "hsl(var(--border))" }}
        >
          <div className="px-3 py-1.5 text-[10px] uppercase tracking-wide font-mono text-muted-foreground bg-muted/15">
            {group}
          </div>
          <ul>
            {slugs.map((slug) => {
              const active = slug === selected;
              return (
                <li key={slug}>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => onSelect(slug)}
                    className={`h-auto w-full justify-start rounded-none px-3 py-2 font-mono text-[12px] ${active ? "bg-foreground/10" : ""}`}
                  >
                    {slug}
                  </Button>
                </li>
              );
            })}
          </ul>
        </div>
      ))}
    </div>
  );
}
