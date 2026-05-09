import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { useAdminAddMember, type AdminUser } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

export function AddMemberRow({
  teamID,
  users,
  email,
  setEmail,
  role,
  setRole,
  onCancel,
  onAdded,
}: {
  teamID: string;
  users: AdminUser[];
  email: string;
  setEmail: (v: string) => void;
  role: "member" | "admin" | "owner";
  setRole: (v: "member" | "admin" | "owner") => void;
  onCancel: () => void;
  onAdded: () => void;
}) {
  const { t } = useTranslation(["app", "common"]);
  const addMember = useAdminAddMember();
  const runToast = useMutationToast();

  async function add() {
    if (!email.trim()) return;
    const ok = await runToast(addMember.mutateAsync({ teamID, email: email.trim(), role }), {
      success: t("app:admin.add_member_added"),
      error: t("app:admin.add_member_failed"),
    });
    if (ok !== null) onAdded();
  }

  return (
    <tr className="bg-muted/20">
      <td colSpan={6} className="px-3 py-3">
        <div className="flex flex-wrap gap-2 items-center">
          <input
            list="users-emails"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="email@example.com"
            className="flex-1 min-w-48 rounded-md border bg-background px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          />
          <datalist id="users-emails">
            {users.map((u) => <option key={u.id} value={u.email} />)}
          </datalist>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as typeof role)}
            className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono"
          >
            <option value="member">member</option>
            <option value="admin">admin</option>
            <option value="owner">owner</option>
          </select>
          <Button size="sm" onClick={add} disabled={addMember.isPending || !email.trim()}>
            {t("app:admin.add_member_submit")}
          </Button>
          <Button size="sm" variant="ghost" onClick={onCancel}>
            {t("common:actions.cancel")}
          </Button>
        </div>
      </td>
    </tr>
  );
}
