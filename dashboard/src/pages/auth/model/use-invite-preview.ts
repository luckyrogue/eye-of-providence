import { useEffect, useState } from "react";
import { previewInvite, type InvitePreview } from "../../../entities/team";
export function useInvitePreview() {
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [invitePreview, setInvitePreview] = useState<InvitePreview | null>(null);
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get("invite");
    if (!code) return;
    setInviteCode(code);
    previewInvite(code)
      .then(setInvitePreview)
      .catch(() => {});
  }, []);
  return { inviteCode, invitePreview };
}
