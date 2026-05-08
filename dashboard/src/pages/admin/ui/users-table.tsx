import { Card, CardContent, CardDescription, CardHeader, CardTitle, EmptyState, Select, useConfirm } from "@eop/ui";
import { Trash2 } from "lucide-react";
import { useAdminDeleteUser, useAdminUpdateUser, type AdminUser } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { formatDate } from "../../../shared/lib/tz";
import { IconButton } from "./icon-button";

export function UsersTable({ users, tz }: { users: AdminUser[]; tz: string }) {
  const update = useAdminUpdateUser();
  const del = useAdminDeleteUser();
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function setRole(uid: string, role: string) {
    await runToast(update.mutateAsync({ userID: uid, payload: { global_role: role } }), {
      success: "Роль обновлена",
      error: "Не удалось изменить роль",
    });
  }

  async function destroy(uid: string, email: string) {
    const ok = await confirm({
      title: `Удалить ${email}?`,
      description: "Все события и отчёты будут стерты. Необратимо.",
      destructive: true,
      confirmText: "Удалить",
    });
    if (!ok) return;
    await runToast(del.mutateAsync(uid), { success: "Пользователь удалён", error: "Не удалось удалить" });
  }

  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">Все пользователи</CardTitle>
        <CardDescription>{users.length} учётных записей</CardDescription>
      </CardHeader>
      <CardContent>
        {users.length === 0 ? (
          <EmptyState eyebrow="No users" title="Пока никого нет" />
        ) : (
          <div className="overflow-x-auto rounded-md border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="py-2.5 px-3 text-left">Email</th>
                  <th className="py-2.5 px-3 text-left">Имя</th>
                  <th className="py-2.5 px-3 text-left">Global role</th>
                  <th className="py-2.5 px-3 text-right">Команды</th>
                  <th className="py-2.5 px-3 text-left">Создан</th>
                  <th className="py-2.5 px-3 text-right" />
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id} className="border-t hover:bg-muted/30">
                    <td className="py-2 px-3 font-mono text-xs">{u.email}</td>
                    <td className="py-2 px-3">{u.display_name}</td>
                    <td className="py-2 px-3">
                      <Select
                        mono
                        value={u.global_role}
                        onChange={(e) => setRole(u.id, e.target.value)}
                        disabled={update.isPending}
                        className="px-2 py-0.5 text-xs"
                      >
                        <option value="user">user</option>
                        <option value="super_admin">super_admin</option>
                      </Select>
                    </td>
                    <td className="py-2 px-3 text-right tabular-nums">{u.teams_count ?? 0}</td>
                    <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(u.created_at, tz)}</td>
                    <td className="py-2 px-3 text-right">
                      <IconButton danger title="Удалить пользователя" onClick={() => destroy(u.id, u.email)}>
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
