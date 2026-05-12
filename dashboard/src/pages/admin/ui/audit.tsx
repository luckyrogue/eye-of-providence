import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Input, SimpleSelect } from "@eop/ui";
import { useAdminAudit, type AuditEntry } from "../../../entities/admin";
import { formatDate } from "../../../shared/lib/tz";

const ACTION_OPTIONS = [
  "",
  "team.deleted",
  "team.created",
  "team.member_added",
  "user.deleted",
  "user.role_changed",
  "subscription.set",
  "sso.saved",
  "sso.deleted",
  "device.claimed",
  "device.revoked",
];

const TARGET_OPTIONS = ["", "team", "user", "sso", "device", "payment"];

export function AuditLog({ tz }: { tz: string }) {
  const { t } = useTranslation("app");
  const [action, setAction] = useState("");
  const [targetType, setTargetType] = useState("");
  const [targetID, setTargetID] = useState("");
  const filter = {
    // SimpleSelect не разрешает empty value, поэтому используем sentinel
    // "__any_*__" и маппим обратно в undefined для backend.
    action: action && action !== "__any_action__" ? action : undefined,
    target_type: targetType && targetType !== "__any_target__" ? targetType : undefined,
    target_id: targetID || undefined,
    limit: 200,
  };
  const { data, isPending, isError, error } = useAdminAudit(filter);

  return (
    <div className="eop-card">
      <div className="card-head">
        <div>
          <div className="card-title">{t("admin.audit_title")}</div>
          <div className="card-sub">
            {data ? t("admin.audit_sub", { n: data.length }) : t("admin.loading")}
          </div>
        </div>
      </div>

      <div className="flex flex-wrap gap-2 mb-4">
        <SimpleSelect
          value={action}
          onValueChange={setAction}
          options={ACTION_OPTIONS.map((a) => ({
            value: a || "__any_action__",
            label: a || t("admin.audit_filter_any_action"),
          }))}
          placeholder={t("admin.audit_filter_any_action")}
          triggerClassName="min-w-[200px]"
        />
        <SimpleSelect
          value={targetType}
          onValueChange={setTargetType}
          options={TARGET_OPTIONS.map((tt) => ({
            value: tt || "__any_target__",
            label: tt || t("admin.audit_filter_any_target"),
          }))}
          placeholder={t("admin.audit_filter_any_target")}
          triggerClassName="min-w-[160px]"
        />
        <Input
          type="text"
          value={targetID}
          onChange={(e) => setTargetID(e.target.value)}
          placeholder={t("admin.audit_filter_target_id_ph")}
          className="font-mono flex-1 min-w-[200px]"
        />
      </div>

      {isError && (
        <div className="text-[13px] mb-4" style={{ color: "hsl(var(--destructive))" }}>
          {error?.message ?? t("admin.error_lead")}
        </div>
      )}

      {isPending ? (
        <div className="text-[13px] text-muted-foreground py-3">{t("admin.loading")}</div>
      ) : !data || data.length === 0 ? (
        <div className="text-[13px] text-muted-foreground py-3">{t("admin.audit_empty")}</div>
      ) : (
        <div className="log-table">
          <div className="log-row head">
            <span>{t("admin.audit_col_time") || "TIME"}</span>
            <span>{t("admin.audit_col_actor") || "ACTOR"}</span>
            <span>{t("admin.audit_col_action_target") || "ACTION · TARGET"}</span>
            <span>{t("admin.audit_col_ip") || "IP"}</span>
            <span />
          </div>
          {data.map((e) => (
            <AuditRow key={e.id} entry={e} tz={tz} />
          ))}
        </div>
      )}
    </div>
  );
}

function AuditRow({ entry, tz }: { entry: AuditEntry; tz: string }) {
  return (
    <div className="log-row">
      <span className="log-time">{formatDate(entry.created_at, tz)}</span>
      <span className="log-file truncate">{entry.actor_email ?? entry.actor_id ?? "—"}</span>
      <span className="log-file truncate">
        <span className={`tag ${tagForAction(entry.action)}`}>{entry.action}</span>
        {entry.target_type && (
          <span className="path ml-2">
            · {entry.target_type}/{shortID(entry.target_id)}
          </span>
        )}
        {entry.metadata && Object.keys(entry.metadata).length > 0 && (
          <span className="path ml-2">· {renderMeta(entry.metadata)}</span>
        )}
      </span>
      <span className="log-lines">{entry.ip ?? "—"}</span>
      <span />
    </div>
  );
}

function tagForAction(action: string): string {
  if (action.startsWith("team.deleted") || action.startsWith("user.deleted")) return "refactor";
  if (action.startsWith("sso")) return "ai-agent";
  if (action.startsWith("subscription")) return "paste-ai";
  if (action.startsWith("user.role")) return "ai-inline";
  if (action.startsWith("device")) return "typed";
  return "ai-inline";
}

function shortID(id?: string): string {
  if (!id) return "—";
  if (id.length > 8) return id.slice(0, 8);
  return id;
}

function renderMeta(meta: Record<string, unknown>): string {
  const pairs = Object.entries(meta)
    .filter(([k]) => k !== "ip" && k !== "user_agent")
    .slice(0, 3)
    .map(([k, v]) => `${k}=${typeof v === "object" ? JSON.stringify(v) : String(v)}`);
  return pairs.join(" ");
}
