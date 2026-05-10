// DateHeading — формат YYYY-MM-DD → "10 May 2026" в локали юзера.

export function DateHeading({ date }: { date: string }) {
  const d = new Date(date + "T00:00:00Z");
  const formatted = d.toLocaleDateString(undefined, {
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  });
  return (
    <h3 className="font-mono text-[11px] uppercase tracking-widest3 text-muted-foreground">
      {formatted}
    </h3>
  );
}
