import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, cn } from "@eop/ui";
import {
  connectionStatus,
  isPaused,
  pendingCount,
  setPaused,
  type ConnectionStatus,
} from "../shared/api/tauri";

const POLL_MS = 5000;

function StatusRow({ label, value, online }: { label: string; value: string; online: boolean }) {
  const dot = online ? "bg-success" : "bg-destructive";
  return (
    <li className="flex items-center gap-3 rounded-md border p-3">
      <span className={`h-2 w-2 shrink-0 rounded-full ${dot}`} />
      <div className="flex flex-1 items-baseline justify-between gap-3">
        <span className="text-sm font-medium">{label}</span>
        <span className="text-sm text-muted-foreground">{value}</span>
      </div>
    </li>
  );
}

function PlayIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden>
      <path d="M8 5v14l11-7z" />
    </svg>
  );
}

function PauseIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden>
      <path d="M6 5h4v14H6V5zm8 0h4v14h-4V5z" />
    </svg>
  );
}

export function StatusPage() {
  const { t } = useTranslation("agent");
  const [pending, setPending] = useState<number | null>(null);
  const [conn, setConn] = useState<ConnectionStatus | null>(null);
  const [paused, setPausedState] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const [p, c, pause] = await Promise.all([pendingCount(), connectionStatus(), isPaused()]);
      setPending(p);
      setConn(c);
      setPausedState(pause);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  }, []);

  useEffect(() => {
    void refresh();
    const id = setInterval(() => void refresh(), POLL_MS);
    return () => clearInterval(id);
  }, [refresh]);

  const onToggleTracking = useCallback(async () => {
    try {
      const next = !paused;
      await setPaused(next);
      setPausedState(next);
    } catch (e) {
      console.warn("[eop] set_paused failed", e);
    }
  }, [paused]);

  const trackingDesc = error
    ? t("status_error")
    : paused
      ? t("status_paused")
      : pending === null
        ? t("status_loading")
        : pending === 0
          ? t("status_all_sent")
          : t("status_buffer", { count: pending });

  const backendOnline = conn?.backend === "online";
  const localOnline = conn?.local_api === "online";
  const trackingActive = !paused;

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t("status_title")}</CardTitle>
          <CardDescription>{trackingDesc}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <button
              type="button"
              onClick={() => void onToggleTracking()}
              title={paused ? t("tracking_play") : t("tracking_pause")}
              aria-label={paused ? t("tracking_play") : t("tracking_pause")}
              className={cn(
                "flex h-14 w-14 shrink-0 items-center justify-center rounded-full transition-colors",
                trackingActive
                  ? "bg-primary text-primary-foreground shadow-sm hover:bg-primary/90"
                  : "border border-border bg-muted text-muted-foreground hover:bg-secondary",
              )}
            >
              {paused ? <PlayIcon className="ml-0.5 h-6 w-6" /> : <PauseIcon className="h-6 w-6" />}
            </button>
            <div className="min-w-0 flex-1 space-y-1">
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    "h-2 w-2 rounded-full",
                    trackingActive ? "bg-success" : "bg-muted-foreground",
                  )}
                />
                <span className="text-sm font-medium">
                  {trackingActive ? t("status_active") : t("status_paused_label")}
                </span>
              </div>
              <p className="text-xs text-muted-foreground">{t("status_hint")}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("conn_title")}</CardTitle>
          <CardDescription>{t("conn_hint")}</CardDescription>
        </CardHeader>
        <CardContent>
          <ul className="space-y-2">
            <StatusRow
              label={t("conn_backend")}
              value={
                conn === null
                  ? t("conn_loading")
                  : backendOnline
                    ? t("conn_online")
                    : t("conn_offline")
              }
              online={backendOnline}
            />
            <StatusRow
              label={t("conn_local_api")}
              value={
                conn === null
                  ? t("conn_loading")
                  : localOnline
                    ? t("conn_local_online", { port: conn.local_api_port })
                    : t("conn_local_offline", { port: conn?.local_api_port ?? 7373 })
              }
              online={localOnline}
            />
            <StatusRow
              label={t("conn_account")}
              value={
                conn === null
                  ? t("conn_loading")
                  : conn.paired
                    ? t("conn_paired")
                    : t("conn_not_paired")
              }
              online={conn?.paired ?? false}
            />
          </ul>
        </CardContent>
      </Card>

      {error && (
        <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-xs text-destructive">
          {error}
        </div>
      )}
    </div>
  );
}
