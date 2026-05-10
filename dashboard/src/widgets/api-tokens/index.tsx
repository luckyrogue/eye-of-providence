// API tokens management — Settings page section.
//
// 3 состояния:
//   - List: existing tokens с last_used_at + revoke button
//   - Create: form (name, scope, ttl_days)
//   - Show-secret: после создания показываем plaintext один раз
//
// Plaintext доступен ровно при создании; UI настойчиво напоминает скопировать.

import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Button, Card, CardContent, CardDescription, CardHeader, CardTitle,
  Input, Modal, Select, useConfirm,
} from "@eop/ui";
import { Copy, KeyRound, Plus, Trash2 } from "lucide-react";
import {
  useTokens, useCreateToken, useRevokeToken, type APIToken,
} from "../../entities/user";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";

export function APITokensWidget() {
  const { t } = useTranslation("developer");
  const tokens = useTokens();
  const create = useCreateToken();
  const revoke = useRevokeToken();
  const confirm = useConfirm();
  const runToast = useMutationToast();

  const [showCreate, setShowCreate] = useState(false);
  const [showSecret, setShowSecret] = useState<string | null>(null);

  // Create-form fields
  const [name, setName] = useState("");
  const [scope, setScope] = useState<APIToken["scope"]>("read");
  const [ttl, setTtl] = useState(0);

  async function submitCreate() {
    if (!name.trim()) return;
    const r = await runToast(create.mutateAsync({ name: name.trim(), scope, ttlDays: ttl }), {});
    if (r) {
      setShowSecret(r.token);
      setShowCreate(false);
      setName("");
      setScope("read");
      setTtl(0);
    }
  }

  async function doRevoke(tok: APIToken) {
    const ok = await confirm({
      title: t("tokens_revoke_confirm"),
      description: t("tokens_revoke_confirm_lead"),
      destructive: true,
      confirmText: t("tokens_revoke_btn"),
    });
    if (!ok) return;
    await runToast(revoke.mutateAsync(tok.id), {});
  }

  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <KeyRound className="h-4 w-4 text-muted-foreground" />
            <CardTitle>{t("tokens_title")}</CardTitle>
          </div>
          <CardDescription className="mt-1">{t("tokens_lead")}</CardDescription>
        </div>
        <Button size="sm" onClick={() => setShowCreate(true)}>
          <Plus className="h-3.5 w-3.5 mr-1" />
          {t("tokens_create")}
        </Button>
      </CardHeader>
      <CardContent>
        {tokens.isPending ? (
          <p className="text-sm text-muted-foreground">…</p>
        ) : !tokens.data || tokens.data.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("tokens_empty")}</p>
        ) : (
          <ul className="divide-y">
            {tokens.data.map((tok) => <TokenRow key={tok.id} token={tok} onRevoke={() => doRevoke(tok)} />)}
          </ul>
        )}
      </CardContent>

      {/* Create form */}
      <Modal open={showCreate} onClose={() => setShowCreate(false)} className="max-w-md">
        <div className="p-6 space-y-4">
          <h3 className="font-display text-lg">{t("tokens_create")}</h3>
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
              <Select value={scope} onChange={(e) => setScope(e.target.value as APIToken["scope"])} className="w-full px-3 py-2 text-sm">
                <option value="read">{t("tokens_scope_read")}</option>
                <option value="write:ingest">{t("tokens_scope_write_ingest")}</option>
                <option value="admin">{t("tokens_scope_admin")}</option>
              </Select>
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
          <div className="flex justify-end gap-2 pt-2">
            <Button variant="ghost" size="sm" onClick={() => setShowCreate(false)}>
              {t("tokens_close")}
            </Button>
            <Button size="sm" onClick={submitCreate} disabled={!name.trim() || create.isPending}>
              {t("tokens_create")}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Show-secret */}
      <Modal open={!!showSecret} onClose={() => setShowSecret(null)} className="max-w-lg">
        <div className="p-6 space-y-4">
          <h3 className="font-display text-lg">{t("tokens_show_warning")}</h3>
          {showSecret && <SecretField value={showSecret} />}
          <div className="flex justify-end pt-2">
            <Button size="sm" onClick={() => setShowSecret(null)}>{t("tokens_close")}</Button>
          </div>
        </div>
      </Modal>
    </Card>
  );
}

function TokenRow({ token, onRevoke }: { token: APIToken; onRevoke: () => void }) {
  const { t } = useTranslation("developer");
  const lastUsed = token.last_used_at
    ? t("tokens_last_used", { at: new Date(token.last_used_at).toLocaleString() })
    : t("tokens_never_used");
  const expires = token.expires_at
    ? t("tokens_expires", { at: new Date(token.expires_at).toLocaleDateString() })
    : t("tokens_no_expiry");

  return (
    <li className="flex items-center justify-between py-3 gap-4">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="font-medium text-sm">{token.name}</span>
          <code className="font-mono text-xs px-1.5 py-0.5 rounded bg-muted">{token.prefix}…</code>
          <span className="text-[11px] uppercase tracking-wider text-muted-foreground">{token.scope}</span>
        </div>
        <div className="text-xs text-muted-foreground mt-1">
          {lastUsed} · {expires}
        </div>
      </div>
      <Button variant="ghost" size="sm" onClick={onRevoke}>
        <Trash2 className="h-3.5 w-3.5 mr-1" />
        {t("tokens_revoke")}
      </Button>
    </li>
  );
}

function SecretField({ value }: { value: string }) {
  const { t } = useTranslation("developer");
  const [copied, setCopied] = useState(false);
  function copy() {
    void navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }
  return (
    <div className="flex items-center gap-2">
      <input
        readOnly
        value={value}
        className="flex-1 rounded-md border bg-background px-3 py-2 text-xs font-mono"
      />
      <Button size="sm" variant="outline" onClick={copy}>
        <Copy className="h-3.5 w-3.5 mr-1" />
        {copied ? t("tokens_copied") : t("tokens_copy")}
      </Button>
    </div>
  );
}
