import { SUPPORTED_LOCALES, type Locale } from "@eop/i18n";

export type { Locale };

/** Try current UI locale, then English, then Russian. */
export const MARKDOWN_LOCALE_FALLBACK: readonly Locale[] = ["en", "ru"];

export function resolveContentLocale(raw: string | undefined): Locale {
  const base = raw?.split("-")[0]?.toLowerCase() ?? "ru";
  return (SUPPORTED_LOCALES as readonly string[]).includes(base) ? (base as Locale) : "ru";
}

export function localizedMarkdownPaths(
  base: "legal" | "docs",
  file: string,
  locale: Locale,
): string[] {
  const order: Locale[] = [locale, ...MARKDOWN_LOCALE_FALLBACK.filter((l) => l !== locale)];
  const seen = new Set<Locale>();
  return order
    .filter((l) => {
      if (seen.has(l)) return false;
      seen.add(l);
      return true;
    })
    .map((l) => `/${base}/${l}/${file}`);
}

export async function fetchLocalizedMarkdown(
  base: "legal" | "docs",
  file: string,
  locale: Locale,
): Promise<{ text: string; path: string; localeUsed: Locale }> {
  for (const path of localizedMarkdownPaths(base, file, locale)) {
    const localeUsed = path.split("/")[2] as Locale;
    const r = await fetch(path);
    if (r.ok) return { text: await r.text(), path, localeUsed };
  }
  throw new Error(`missing markdown: ${base}/${file}`);
}
