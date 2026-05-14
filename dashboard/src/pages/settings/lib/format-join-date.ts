const ISO_DATE: Intl.DateTimeFormatOptions = { year: "numeric", month: "long", day: "numeric" };
export function formatJoinDate(iso: string, locale: string): string {
  return new Date(iso).toLocaleDateString(locale === "kk" ? "kk-KZ" : locale, ISO_DATE);
}
