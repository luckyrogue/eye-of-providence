import type { ChangelogEntry } from "../model/types";
import { DateHeading } from "./date-heading";
import { Row } from "./row";
function groupByDate(entries: ChangelogEntry[]): Map<string, ChangelogEntry[]> {
  const groups = new Map<string, ChangelogEntry[]>();
  for (const e of entries) {
    const arr = groups.get(e.date) ?? [];
    arr.push(e);
    groups.set(e.date, arr);
  }
  return groups;
}
export function Timeline({ entries }: { entries: ChangelogEntry[] }) {
  const groups = groupByDate(entries);
  const dates = Array.from(groups.keys());
  return (
    <ol className="space-y-8">
      {dates.map((date) => (
        <li key={date}>
          <DateHeading date={date} />
          <ul className="space-y-2 mt-3">
            {groups.get(date)!.map((e) => (
              <Row key={e.hash} entry={e} />
            ))}
          </ul>
        </li>
      ))}
    </ol>
  );
}
