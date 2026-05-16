import { useId, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Input, SimpleSelect, type SimpleSelectOption } from "@eop/ui";
import { useAdminAddMember, type AdminUser } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

type Role = "member" | "admin" | "owner";

// Inline-форма «add team member». Без table-разметки — годится и для DataTable
// (через renderSubComponent), и для любых других обёрток.
export function AddMemberForm({
  teamID,
  users,
  onCancel,
  onAdded,
}: {
  teamID: string;
  users: AdminUser[];
  onCancel: () => void;
  onAdded: () => void;
}) {
  const { t } = useTranslation(["app", "common"]);
  const addMember = useAdminAddMember();
  const runToast = useMutationToast();
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<Role>("member");
  const datalistId = useId();

  const roleOptions = useMemo<SimpleSelectOption[]>(
    () => [
      { value: "member", label: t("app:admin.add_member_role_member") },
      { value: "admin", label: t("app:admin.add_member_role_admin") },
      { value: "owner", label: t("app:admin.add_member_role_owner") },
    ],
    [t],
  );

  async function add() {
    if (!email.trim()) return;
    const ok = await runToast(addMember.mutateAsync({ teamID, email: email.trim(), role }), {
      success: t("app:admin.add_member_added"),
      error: t("app:admin.add_member_failed"),
    });
    if (ok !== null) onAdded();
  }

  return (
    <div className="bg-muted/20 px-3 py-3">
      <div className="flex flex-wrap items-center gap-2">
        <Input
          list={datalistId}
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder={t("app:admin.add_member_email_placeholder")}
          className="min-w-48 flex-1"
        />
        <datalist id={datalistId}>
          {users.map((u) => (
            <option key={u.id} value={u.email} />
          ))}
        </datalist>
        <SimpleSelect
          value={role}
          onValueChange={(v) => setRole(v as Role)}
          options={roleOptions}
          triggerClassName="w-32 font-mono"
        />
        <Button
          size="sm"
          onClick={() => void add()}
          disabled={addMember.isPending || !email.trim()}
        >
          {t("app:admin.add_member_submit")}
        </Button>
        <Button size="sm" variant="ghost" onClick={onCancel}>
          {t("common:actions.cancel")}
        </Button>
      </div>
    </div>
  );
}
