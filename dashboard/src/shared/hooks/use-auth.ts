// Тонкая обёртка над localStorage для auth state. Один источник правды для:
// - "залогинен ли пользователь" (есть ли user_id)
// - logout (чистит ВСЕ eop_* keys)
// - сохранение auth response после успешного login/register

import { useCallback, useSyncExternalStore } from "react";
import type { AuthResponse } from "../../entities/user";

const KEY = "eop_user_id";

function subscribe(cb: () => void) {
  window.addEventListener("storage", cb);
  return () => window.removeEventListener("storage", cb);
}

function useUserId(): string | null {
  // useSyncExternalStore — реакция на cross-tab logout (`storage` event).
  return useSyncExternalStore(
    subscribe,
    () => localStorage.getItem(KEY),
    () => null,
  );
}

export function useAuth() {
  const userId = useUserId();

  const setAuth = useCallback((r: AuthResponse) => {
    localStorage.setItem(KEY, r.user_id);
    if (r.display_name) localStorage.setItem("eop_display_name", r.display_name);
    // useSyncExternalStore слушает `storage` event только cross-tab.
    // Same-tab — диспатчим вручную (generic Event достаточно — listener
    // не читает поля из event-объекта).
    window.dispatchEvent(new Event("storage"));
  }, []);

  const logout = useCallback(() => {
    for (let i = localStorage.length - 1; i >= 0; i--) {
      const k = localStorage.key(i);
      if (k && k.startsWith("eop_")) localStorage.removeItem(k);
    }
    window.dispatchEvent(new Event("storage"));
  }, []);

  return { userId, isAuthed: !!userId, setAuth, logout };
}
