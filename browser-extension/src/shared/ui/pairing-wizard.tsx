import { useCallback, useEffect, useRef, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import {
  DEFAULT_BACKEND,
  backendDisplayHost,
  dashboardUrlFor,
  pairBegin,
  pairPoll,
  setConfig,
  type PairBeginResponse,
} from "../api/backend";

type Phase =
  | { kind: "idle" }
  | { kind: "starting" }
  | { kind: "waiting"; pair: PairBeginResponse }
  | { kind: "expired" }
  | { kind: "error"; message: string };

const POLL_INTERVAL_MS = 2_500;

// Единый flow для popup и options. Polling до claim или expire (POLL_INTERVAL_MS).
// После успешного claim — onClaimed; в нём caller сохраняет token и
// переключает UI в connected-state.
export function PairingWizard({
  backend = DEFAULT_BACKEND,
  onClaimed,
}: {
  backend?: string;
  onClaimed?: () => void;
}) {
  const { t } = useTranslation("popup");
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
      const pair = await pairBegin(backend);
      setPhase({ kind: "waiting", pair });
    } catch (e) {
      console.warn("[eop] pair start failed", e);
      setPhase({ kind: "error", message: t("pair_error") });
    }
  }, [backend, t]);

  // Авто-poll после получения code.
  useEffect(() => {
    if (phase.kind !== "waiting") return;
    const { pair_id, secret } = phase.pair;
    let cancelled = false;
    const tick = async () => {
      try {
        const r = await pairPoll(pair_id, secret, backend);
        if (cancelled) return;
        if (r.status === "expired") {
          stopPoll();
          setPhase({ kind: "expired" });
          return;
        }
        if (r.status === "claimed" && r.token) {
          stopPoll();
          await setConfig(r.token, backend);
          onClaimed?.();
        }
      } catch (e) {
        // network-flake — просто следующий тик попробует снова
        console.debug("[eop] poll failed", e);
      }
    };
    pollTimer.current = window.setInterval(() => void tick(), POLL_INTERVAL_MS);
    void tick();
    return () => {
      cancelled = true;
      stopPoll();
    };
  }, [phase, backend, stopPoll, onClaimed]);

  if (phase.kind === "idle") {
    return (
      <div className="space-y-3">
        <p className="text-xs text-muted-foreground">
          <Trans
            ns="popup"
            i18nKey="pair_lead"
            values={{ host: backendDisplayHost(backend) }}
            components={{ code: <code className="rounded bg-muted px-1" /> }}
          />
        </p>
        <Button size="sm" className="w-full" onClick={() => void start()}>
          {t("pair_start")}
        </Button>
      </div>
    );
  }
  if (phase.kind === "starting") {
    return <div className="text-xs text-muted-foreground">{t("pair_busy")}</div>;
  }
  if (phase.kind === "expired") {
    return (
      <div className="space-y-2">
        <p className="text-xs text-destructive">{t("pair_expired")}</p>
        <Button size="sm" className="w-full" onClick={() => void start()}>
          {t("pair_restart")}
        </Button>
      </div>
    );
  }
  if (phase.kind === "error") {
    return (
      <div className="space-y-2">
        <p className="text-xs text-destructive">{phase.message}</p>
        <Button size="sm" className="w-full" onClick={() => void start()}>
          {t("pair_restart")}
        </Button>
      </div>
    );
  }
  // waiting
  const dashboardURL = dashboardUrlFor(backend) + "/settings";
  return (
    <div className="space-y-3">
      <div className="space-y-1">
        <div className="text-xs text-muted-foreground">{t("pair_code_title")}</div>
        <div className="font-mono text-3xl tracking-[0.4em] text-center py-3 bg-muted rounded select-all">
          {phase.pair.code}
        </div>
        <div className="text-[11px] text-muted-foreground">{t("pair_code_hint")}</div>
      </div>
      <div className="flex gap-2">
        <Button
          size="sm"
          variant="outline"
          className="flex-1"
          onClick={() => void chrome.tabs.create({ url: dashboardURL })}
        >
          {t("pair_open_dashboard")}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          onClick={() => {
            stopPoll();
            setPhase({ kind: "idle" });
          }}
        >
          {t("pair_cancel")}
        </Button>
      </div>
      <div className="text-[11px] text-muted-foreground text-center animate-pulse">
        {t("pair_waiting")}
      </div>
    </div>
  );
}
