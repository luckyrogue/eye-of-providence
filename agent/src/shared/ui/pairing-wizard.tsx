import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { openUrl, pairBegin, pairPoll, type PairBeginResponse } from "../api/tauri";

type Phase =
  | { kind: "idle" }
  | { kind: "starting" }
  | { kind: "waiting"; pair: PairBeginResponse }
  | { kind: "claimed" }
  | { kind: "expired" }
  | { kind: "error"; message: string };

const POLL_INTERVAL_MS = 2_500;

function dashboardHost(backend: string): string {
  try {
    return new URL(backend).host;
  } catch {
    return backend;
  }
}

export function PairingWizard({ backend, onClaimed }: { backend: string; onClaimed?: () => void }) {
  const { t } = useTranslation("agent");
  const [phase, setPhase] = useState<Phase>({ kind: "idle" });
  const pollTimer = useRef<number | null>(null);

  const stopPoll = useCallback(() => {
    if (pollTimer.current !== null) {
      window.clearInterval(pollTimer.current);
      pollTimer.current = null;
    }
  }, []);

  useEffect(() => () => stopPoll(), [stopPoll]);

  const start = useCallback(async () => {
    setPhase({ kind: "starting" });
    try {
      const pair = await pairBegin();
      setPhase({ kind: "waiting", pair });
    } catch (e) {
      console.warn("[eop] pair start failed", e);
      setPhase({ kind: "error", message: t("settings_pair_error") });
    }
  }, [t]);

  useEffect(() => {
    if (phase.kind !== "waiting") return;
    const { pair_id, secret } = phase.pair;
    let cancelled = false;
    const tick = async () => {
      try {
        const r = await pairPoll(pair_id, secret);
        if (cancelled) return;
        if (r.status === "expired") {
          stopPoll();
          setPhase({ kind: "expired" });
          return;
        }
        if (r.status === "claimed") {
          stopPoll();
          setPhase({ kind: "claimed" });
          onClaimed?.();
        }
      } catch (e) {
        console.debug("[eop] poll failed", e);
      }
    };
    pollTimer.current = window.setInterval(() => void tick(), POLL_INTERVAL_MS);
    void tick();
    return () => {
      cancelled = true;
      stopPoll();
    };
  }, [phase, stopPoll, onClaimed]);

  if (phase.kind === "idle") {
    return (
      <div className="space-y-3">
        <Button size="sm" onClick={() => void start()}>
          {t("settings_pair_start")}
        </Button>
      </div>
    );
  }
  if (phase.kind === "starting") {
    return <div className="text-sm text-muted-foreground">{t("settings_pair_busy")}</div>;
  }
  if (phase.kind === "claimed") {
    // Транзитное состояние: poll получил claimed, токен в keyring,
    // ждём пока settings.tsx refresh пересоберёт UI. Без явного state'а
    // юзер видел бы «Waiting for confirmation» ещё 3 секунды (poll
    // interval) — выглядело как баг.
    return (
      <div className="text-sm text-success" role="status">
        {t("settings_pair_claimed", { defaultValue: "✓ Привязано, синхронизация…" })}
      </div>
    );
  }
  if (phase.kind === "expired") {
    return (
      <div className="space-y-2">
        <p className="text-sm text-destructive">{t("settings_pair_expired")}</p>
        <Button size="sm" onClick={() => void start()}>
          {t("settings_pair_restart")}
        </Button>
      </div>
    );
  }
  if (phase.kind === "error") {
    return (
      <div className="space-y-2">
        <p className="text-sm text-destructive">{phase.message}</p>
        <Button size="sm" onClick={() => void start()}>
          {t("settings_pair_restart")}
        </Button>
      </div>
    );
  }

  const dashboardURL = `https://${dashboardHost(backend)}/integrations`;
  return (
    <div className="space-y-3">
      <div className="space-y-1">
        <div className="text-xs text-muted-foreground">{t("settings_pair_code")}</div>
        <div className="font-mono text-3xl tracking-[0.4em] text-center py-3 bg-muted rounded select-all">
          {phase.pair.code}
        </div>
        <div className="text-xs text-muted-foreground">{t("settings_pair_code_hint")}</div>
      </div>
      <div className="flex gap-2">
        <Button
          size="sm"
          variant="outline"
          className="flex-1"
          onClick={() => void openUrl(dashboardURL)}
        >
          {t("settings_pair_open_dashboard")}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          onClick={() => {
            stopPoll();
            setPhase({ kind: "idle" });
          }}
        >
          {t("settings_pair_cancel")}
        </Button>
      </div>
      <div className="text-xs text-muted-foreground text-center animate-pulse">
        {t("settings_pair_waiting")}
      </div>
    </div>
  );
}
