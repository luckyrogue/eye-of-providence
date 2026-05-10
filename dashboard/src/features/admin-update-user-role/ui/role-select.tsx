import { useTranslation } from "react-i18next";
import { SimpleSelect } from "@eop/ui";
import { useAdminUpdateUser } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

export function UserRoleSelect({ userID, value }: { userID: string; value: string }) {
  const { t } = useTranslation("app");
  const update = useAdminUpdateUser();
  const runToast = useMutationToast();

  async function setRole(role: string) {
    if (role === value) return;
    await runToast(update.mutateAsync({ userID, payload: { global_role: role } }), {
      success: t("admin.users_role_changed"),
      error: t("admin.users_role_change_failed"),
    });
  }

  return (
    <SimpleSelect
      value={value}
      onValueChange={setRole}
      disabled={update.isPending}
      triggerClassName="h-8 w-36 font-mono text-xs"
      options={[
        { value: "user", label: t("admin.users_role_user") },
        { value: "super_admin", label: t("admin.users_role_super_admin") },
      ]}
    />
  );
}
