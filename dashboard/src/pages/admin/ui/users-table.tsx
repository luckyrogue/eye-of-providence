import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, EmptyState } from "@eop/ui";
import type { AdminUser } from "../../../entities/admin";
import { DeleteUserButton } from "../../../features/admin-delete-user";
import { UserRoleSelect } from "../../../features/admin-update-user-role";
import { formatDate } from "../../../shared/lib/tz";

export function UsersTable({ users, tz }: { users: AdminUser[]; tz: string }) {
  const { t } = useTranslation("app");

  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">
          {t("admin.all_users_title", { defaultValue: "All users" })}
        </CardTitle>
        <CardDescription>
          {t("admin.all_users_lead", { count: users.length, defaultValue: `${users.length} accounts` })}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {users.length === 0 ? (
          <EmptyState
            eyebrow={t("admin.all_users_empty_eyebrow", { defaultValue: "No users" })}
            title={t("admin.all_users_empty_title", { defaultValue: "Nobody yet" })}
          />
        ) : (
          <div className="overflow-x-auto rounded-md border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="py-2.5 px-3 text-left">{t("admin.users_email")}</th>
                  <th className="py-2.5 px-3 text-left">{t("admin.users_name")}</th>
                  <th className="py-2.5 px-3 text-left">{t("admin.users_global_role")}</th>
                  <th className="py-2.5 px-3 text-right">{t("admin.users_team")}</th>
                  <th className="py-2.5 px-3 text-left">{t("admin.table_created", { defaultValue: "Created" })}</th>
                  <th className="py-2.5 px-3 text-right" />
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id} className="border-t hover:bg-muted/30">
                    <td className="py-2 px-3 font-mono text-xs">{u.email}</td>
                    <td className="py-2 px-3">{u.display_name}</td>
                    <td className="py-2 px-3">
                      <UserRoleSelect userID={u.id} value={u.global_role} />
                    </td>
                    <td className="py-2 px-3 text-right tabular-nums">{u.teams_count ?? 0}</td>
                    <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(u.created_at, tz)}</td>
                    <td className="py-2 px-3 text-right">
                      <DeleteUserButton userID={u.id} email={u.email} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
