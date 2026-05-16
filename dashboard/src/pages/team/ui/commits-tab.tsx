import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { DataTable, DataTableColumnHeader, EmptyState, type DataTableColumn } from "@eop/ui";
import { useTeamCommits, type Commit } from "../../../entities/team";
import { dtLabels } from "../../../shared/lib/data-table-labels";
import { formatDate } from "../../../shared/lib/tz";
import { renderMessageWithLinks } from "../../../shared/lib/task-refs";

export function CommitsTab({ teamID, tz }: { teamID: string; tz: string }) {
  const { t } = useTranslation(["app", "common"]);
  const commits = useTeamCommits(teamID);
  const list: Commit[] = commits.data ?? [];

  const columns = useMemo<DataTableColumn<Commit>[]>(
    () => [
      {
        accessorKey: "authored_at",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:team_detail.commits_table_time")} />
        ),
        cell: ({ row }) => (
          <span className="font-mono text-xs whitespace-nowrap">
            {formatDate(row.original.authored_at, tz)}
          </span>
        ),
        sortingFn: "datetime",
      },
      {
        accessorKey: "author",
        header: t("app:team_detail.commits_table_author"),
        cell: ({ row }) => row.original.author,
        filterFn: (row, _id, value) =>
          row.original.author.toLowerCase().includes(String(value).toLowerCase()),
      },
      {
        accessorKey: "sha",
        header: t("app:team_detail.commits_table_sha"),
        enableSorting: false,
        cell: ({ row }) => (
          <span className="font-mono text-xs">{row.original.sha.slice(0, 7)}</span>
        ),
      },
      {
        accessorKey: "message",
        header: t("app:team_detail.commits_table_message"),
        enableSorting: false,
        cell: ({ row }) => (
          <div className="max-w-md truncate">
            <CommitMessage message={row.original.message} />
          </div>
        ),
      },
      {
        id: "diff",
        accessorFn: (row) => row.lines_added + row.lines_removed,
        header: ({ column }) => (
          <DataTableColumnHeader
            column={column}
            title={t("app:team_detail.commits_table_diff")}
            align="right"
          />
        ),
        cell: ({ row }) => (
          <div className="text-right tabular-nums text-xs">
            <span className="text-success">+{row.original.lines_added}</span>{" "}
            <span className="text-destructive">-{row.original.lines_removed}</span>
          </div>
        ),
      },
      {
        accessorKey: "ai_lines_pct",
        header: ({ column }) => (
          <DataTableColumnHeader
            column={column}
            title={t("app:team_detail.commits_table_ai")}
            align="right"
          />
        ),
        cell: ({ row }) => (
          <div className="text-right tabular-nums">
            {row.original.ai_lines_pct !== null ? `${row.original.ai_lines_pct}%` : "—"}
          </div>
        ),
      },
    ],
    [t, tz],
  );

  if (list.length === 0 && !commits.isPending) {
    return (
      <EmptyState
        eyebrow={t("app:team_detail.commits_empty_eyebrow")}
        title={t("app:team_detail.commits_empty_title")}
        description={t("app:team_detail.commits_empty_lead")}
      />
    );
  }

  return (
    <DataTable
      columns={columns}
      data={list}
      isLoading={commits.isPending}
      filterableColumn={{ id: "author", placeholder: t("app:team_detail.commits_filter_author") }}
      enableColumnVisibility
      enablePagination
      pageSize={25}
      labels={dtLabels(t)}
    />
  );
}

function CommitMessage({ message }: { message: string }) {
  const segments = renderMessageWithLinks(message);
  return (
    <>
      {segments.map((s, i) =>
        s.kind === "ref" && s.ref ? (
          <a
            key={i}
            href={s.ref.url}
            target="_blank"
            rel="noreferrer noopener"
            className="text-primary hover:underline font-mono"
          >
            {s.value}
          </a>
        ) : (
          <span key={i}>{s.value}</span>
        ),
      )}
    </>
  );
}
