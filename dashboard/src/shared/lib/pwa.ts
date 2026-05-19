import { useEffect, useState } from "react";

const INSTALL_HINT_DISMISSED_KEY = "eop_install_hint_dismissed";

type BeforeInstallPromptEvent = Event & {
  readonly platforms: string[];
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
};

export function registerSW() {
  if (!("serviceWorker" in navigator) || import.meta.env.DEV) return;
  window.addEventListener("load", () => {
    navigator.serviceWorker
      .register("/sw.js")
      .then((reg) => {
        reg.addEventListener("updatefound", () => {
          const newWorker = reg.installing;
          if (!newWorker) return;
          newWorker.addEventListener("statechange", () => {
            if (newWorker.state === "installed" && navigator.serviceWorker.controller) {
              window.dispatchEvent(new CustomEvent("eop:sw-update-available"));
            }
          });
        });
      })
      .catch((err) => {
        console.warn("[eop] sw register failed", err);
      });
  });
}

export function useUpdatePrompt() {
  const [available, setAvailable] = useState(false);
  useEffect(() => {
    const handler = () => setAvailable(true);
    window.addEventListener("eop:sw-update-available", handler);
    return () => window.removeEventListener("eop:sw-update-available", handler);
  }, []);
  return {
    available,
    reload: () => {
      window.location.reload();
    },
  };
}

export function useOnlineStatus() {
  const [online, setOnline] = useState(typeof navigator !== "undefined" ? navigator.onLine : true);
  useEffect(() => {
    const on = () => setOnline(true);
    const off = () => setOnline(false);
    window.addEventListener("online", on);
    window.addEventListener("offline", off);
    return () => {
      window.removeEventListener("online", on);
      window.removeEventListener("offline", off);
    };
  }, []);
  return online;
}

export function isStandalone(): boolean {
  if (typeof window === "undefined") return false;
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    // iOS-specific
    ("standalone" in window.navigator &&
      (window.navigator as { standalone?: boolean }).standalone === true)
  );
}

export function isIOS(): boolean {
  if (typeof navigator === "undefined") return false;
  return /iPad|iPhone|iPod/.test(navigator.userAgent) && !("MSStream" in window);
}

export function useInstallPrompt() {
  const [event, setEvent] = useState<BeforeInstallPromptEvent | null>(null);
  const [dismissed, setDismissed] = useState(
    () =>
      typeof localStorage !== "undefined" &&
      localStorage.getItem(INSTALL_HINT_DISMISSED_KEY) === "1",
  );
  const [installed, setInstalled] = useState(() => isStandalone());

  useEffect(() => {
    function onBeforeInstall(e: Event) {
      e.preventDefault();
      setEvent(e as BeforeInstallPromptEvent);
    }
    function onInstalled() {
      setInstalled(true);
      setEvent(null);
    }
    window.addEventListener("beforeinstallprompt", onBeforeInstall);
    window.addEventListener("appinstalled", onInstalled);
    return () => {
      window.removeEventListener("beforeinstallprompt", onBeforeInstall);
      window.removeEventListener("appinstalled", onInstalled);
    };
  }, []);

  return {
    canInstall: !!event && !installed && !dismissed,
    showIOSHint: !installed && !dismissed && isIOS(),
    installed,
    install: async () => {
      if (!event) return;
      await event.prompt();
      const { outcome } = await event.userChoice;
      if (outcome === "accepted") setEvent(null);
    },
    dismiss: () => {
      localStorage.setItem(INSTALL_HINT_DISMISSED_KEY, "1");
      setDismissed(true);
    },
  };
}
