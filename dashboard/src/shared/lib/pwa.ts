// PWA bootstrap helpers.
//
// registerSW — регистрирует service-worker (только в production: dev-build
// постоянно меняется, sw caching мешает HMR).
//
// useInstallPrompt — capture'ит beforeinstallprompt event (Android Chrome /
// Edge / Samsung Internet) для UI-кнопки "Установить app". iOS не emits
// этот event — там install через Share → Add to home screen, мы только
// показываем hint один раз.

import { useEffect, useState } from "react";

const INSTALL_HINT_DISMISSED_KEY = "eop_install_hint_dismissed";

// BeforeInstallPromptEvent — chromium-only experimental API, не в TS-libs.
type BeforeInstallPromptEvent = Event & {
  readonly platforms: string[];
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
};

export function registerSW() {
  if (!("serviceWorker" in navigator) || import.meta.env.DEV) return;
  // Wait for window load чтобы не конкурировать с initial render.
  window.addEventListener("load", () => {
    navigator.serviceWorker.register("/sw.js").catch((err) => {
      console.warn("[eop] sw register failed", err);
    });
  });
}

// isStandalone — пользователь уже запустил из home-screen (iOS / Android PWA).
export function isStandalone(): boolean {
  if (typeof window === "undefined") return false;
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    // iOS-specific
    ("standalone" in window.navigator &&
      (window.navigator as { standalone?: boolean }).standalone === true)
  );
}

// isIOS — нам нужен иной copy ("Share → Add to home screen" вместо install
// button), потому что Safari не реализует beforeinstallprompt.
export function isIOS(): boolean {
  if (typeof navigator === "undefined") return false;
  return /iPad|iPhone|iPod/.test(navigator.userAgent) && !("MSStream" in window);
}

// useInstallPrompt — returns:
//   - canInstall: можно показать "Install" button (chromium на Android/Desktop)
//   - showIOSHint: iOS-specific instruction (Share → Add to home screen)
//   - install: вызвать prompt (только если canInstall)
//   - dismiss: спрятать hint навсегда (localStorage flag)
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
