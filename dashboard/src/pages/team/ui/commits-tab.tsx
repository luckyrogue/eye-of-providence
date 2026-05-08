import { EmptyState } from "@eop/ui";
import { useTeamCommits, type Commit } from "../../../entities/team";
import { formatDate } from "../../../shared/lib/tz";

export function CommitsTab({ teamID, tz }: { teamID: string; tz: string }) {
  const commits = useTeamCommits(teamID);
  const list: Commit[] = commits.data ?? [];

  if (list.length === 0) {
    return (
      <EmptyState
        eyebrow="No commits yet"
        title="Установи git post-commit hook"
        description="После установки коммиты команды будут показаны здесь с разбивкой AI vs manual."
      />
    );
  }
  return (
    <div className="overflow-x-auto rounded-md border">
      <table className="w-full text-sm">
        <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
          <tr>
            <th className="py-2.5 px-3 text-left">Время</th>
            <th className="py-2.5 px-3 text-left">Автор</th>
            <th className="py-2.5 px-3 text-left">SHA</th>
            <th className="py-2.5 px-3 text-left">Сообщение</th>
            <th className="py-2.5 px-3 text-right">+/-</th>
            <th className="py-2.5 px-3 text-right">AI %</th>
          </tr>
        </thead>
        <tbody>
          {list.map((c) => (
            <tr key={c.id} className="border-t hover:bg-muted/30">
              <td className="py-2 px-3 font-mono text-xs whitespace-nowrap">{formatDate(c.authored_at, tz)}</td>
              <td className="py-2 px-3">{c.author}</td>
              <td className="py-2 px-3 font-mono text-xs">{c.sha.slice(0, 7)}</td>
              <td className="py-2 px-3 max-w-md truncate">{c.message}</td>
              <td className="py-2 px-3 text-right tabular-nums text-xs">
                <span className="text-green-600">+{c.lines_added}</span>{" "}
                <span className="text-red-600">-{c.lines_removed}</span>
              </td>
              <td className="py-2 px-3 text-right tabular-nums">
                {c.ai_lines_pct !== null ? `${c.ai_lines_pct}%` : "—"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
