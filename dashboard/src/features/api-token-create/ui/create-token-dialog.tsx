import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  SecretField,
  SimpleSelect,
} from "@eop/ui";
import { useCreateToken, type APIToken } from "../../../entities/user";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";

// CreateTokenDialog — модалка с двумя стадиями: форма → отображение plaintext
// secret'а ровно один раз. Plaintext доступен ТОЛЬКО на момент создания
// (backend хранит только хеш).
export function CreateTokenDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useTranslation("developer");
  const create = useCreateToken();
  const runToast = useMutationToast();

  const [name, setName] = useState("");
  const [scope, setScope] = useState<APIToken["scope"]>("read");
  const [ttl, setTtl] = useState(0);
  const [secret, setSecret] = useState<string | null>(null);

  function reset() {
    setName("");
    setScope("read");
    setTtl(0);
    setSecret(null);
  }

  function close() {
    reset();
    onClose();
  }

  async function submit() {
    if (!name.trim()) return;
    const r = await runToast(create.mutateAsync({ name: name.trim(), scope, ttlDays: ttl }), {});
    if (r) setSecret(r.token);
  }

  return (
    <>
      <Dialog open={open && !secret} onOpenChange={(o) => !o && close()}>
        <DialogContent className="max-w-md p-6">
          <DialogHeader>
            <DialogTitle>{t("tokens_create")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <label className="text-xs text-muted-foreground">{t("tokens_name")}</label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t("tokens_name_placeholder")}
                maxLength={64}
              />
            </div>
            <div>
              <label className="text-xs text-muted-foreground">{t("tokens_scope")}</label>
              <SimpleSelect
                value={scope}
                onValueChange={(v) => setScope(v as APIToken["scope"])}
                triggerClassName="w-full"
                options={[
                  { value: "read", label: t("tokens_scope_read") },
                  { value: "write:ingest", label: t("tokens_scope_write_ingest") },
                  { value: "admin", label: t("tokens_scope_admin") },
                ]}
              />
            </div>
            <div>
              <label className="text-xs text-muted-foreground">{t("tokens_ttl")}</label>
              <Input
                type="number"
                min={0}
                max={365}
                value={ttl}
                onChange={(e) => setTtl(Math.max(0, Math.min(365, Number(e.target.value) || 0)))}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" size="sm" onClick={close}>
              {t("tokens_close")}
            </Button>
            <Button size="sm" onClick={submit} disabled={!name.trim() || create.isPending}>
              {t("tokens_create")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!secret} onOpenChange={(o) => !o && close()}>
        <DialogContent className="max-w-lg p-6">
          <DialogHeader>
            <DialogTitle>{t("tokens_show_warning")}</DialogTitle>
          </DialogHeader>
          {secret && (
            <SecretField
              value={secret}
              copyLabel={t("tokens_copy")}
              copiedLabel={t("tokens_copied")}
            />
          )}
          <DialogFooter>
            <Button size="sm" onClick={close}>
              {t("tokens_close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
