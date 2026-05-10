// useAuth — единственный publicly-exposed entry для auth state.
// Реакция на cross-tab logout через useSyncExternalStore + `storage` event.

import { useCallback, useSyncExternalStore } from "react";
import {
  clearSession,
  getUserId,
  notifySessionChange,
  setSession,
} from "../../../shared/lib/session-storage";

export type AuthPayload = {
  user_id: string;
  token: string;
  display_name?: string;
};

function subscribe(cb: () => void) {
  window.addEventListener("storage", cb);
  return () => window.removeEventListener("storage", cb);
}

function useUserId(): string | null {
  return useSyncExternalStore(subscribe, getUserId, () => null);
}

export function useAuth() {
  const userId = useUserId();

  const setAuth = useCallback((r: AuthPayload) => {
    setSession({ user_id: r.user_id, token: r.token, display_name: r.display_name });
    notifySessionChange();
  }, []);

  const logout = useCallback(() => {
    clearSession();
    notifySessionChange();
  }, []);

  return { userId, isAuthed: !!userId, setAuth, logout };
}
