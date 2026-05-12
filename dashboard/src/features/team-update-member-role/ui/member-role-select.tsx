import { useTranslation } from "react-i18next";
import { SimpleSelect } from "@eop/ui";
import { useUpdateMemberRole } from "../../../entities/team";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

export function MemberRoleSelect({
  teamID,
  userID,
  value,
  disabled,
}: {
  teamID: string;
  userID: string;
  value: string;
  disabled?: boolean;
}) {
  const { t } = useTranslation("app");
  const updateRole = useUpdateMemberRole(teamID);
  const runToast = useMutationToast();

  async function changeRole(role: string) {
    if (role === value) return;
    await runToast(updateRole.mutateAsync({ userID, role }), {
      success: t("team_detail.member_role_updated"),
      error: t("team_detail.member_role_update_failed"),
    });
  }

  return (
    <SimpleSelect
      value={value}
      onValueChange={changeRole}
      disabled={disabled || updateRole.isPending}
      triggerClassName="h-10 w-32 font-mono text-xs"
      options={[
        { value: "owner", label: t("team_detail.role.owner") },
        { value: "admin", label: t("team_detail.role.admin") },
        { value: "member", label: t("team_detail.role.member") },
      ]}
    />
  );
}
