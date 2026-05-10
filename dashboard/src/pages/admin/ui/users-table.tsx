import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, EmptyState, SimpleSelect, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useAdminDeleteUser, useAdminUpdateUser, type AdminUser } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { formatDate } from "../../../shared/lib/tz";
import { IconButton } from "./icon-button";

export function UsersTable({ users, tz }: { users: AdminUser[]; tz: string }) {
  const { t } = useTranslation("app");
  const update = useAdminUpdateUser();
  const del = useAdminDeleteUser();
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function setRole(uid: string, role: string) {
    await runToast(update.mutateAsync({ userID: uid, payload: { global_role: role } }), {
      success: t("admin.users_role_changed"),
      error: t("admin.users_role_change_failed"),
    });
  }

  async function destroy(uid: string, email: string) {
    const ok = await confirm({
      title: t("admin.user_delete_confirm_title", { email, defaultValue: `Delete ${email}?` }),
      description: t("admin.user_delete_confirm_lead", {
        defaultValue: "All events and reports will be erased. Irreversible.",
      }),
      destructive: true,
      confirmText: t("admin.users_delete"),
    });
    if (!ok) return;
    await runToast(del.mutateAsync(uid), {
      success: t("admin.users_deleted"),
      error: t("admin.users_delete_failed"),
    });
  }

  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">
          {t("admin.all_users_title", { defaultValue: "All users" })}
        </CardTitle>
        <CardDescription>{t("admin.all_users_lead", { count: users.length, defaultValue: `${users.length} accounts` })}</CardDescription>
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
                      <SimpleSelect
                        value={u.global_role}
                        onValueChange={(v) => setRole(u.id, v)}
                        disabled={update.isPending}
                        triggerClassName="h-8 w-36 font-mono text-xs"
                        options={[
                          { value: "user", label: t("admin.users_role_user") },
                          { value: "super_admin", label: t("admin.users_role_super_admin") },
                        ]}
                      />
                    </td>
                    <td className="py-2 px-3 text-right tabular-nums">{u.teams_count ?? 0}</td>
                    <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(u.created_at, tz)}</td>
                    <td className="py-2 px-3 text-right">
                      <IconButton
                        danger
                        title={t("admin.user_delete_btn", { defaultValue: "Delete user" })}
                        onClick={() => destroy(u.id, u.email)}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </IconButton>
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
