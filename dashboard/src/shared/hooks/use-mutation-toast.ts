import { useCallback } from "react";
import { toast } from "@eop/ui";

// Хелпер чтобы не дублировать try/catch в каждом mutation вызове.
//
// run(promise, { success, error }) — выполняет промис, тостит результат,
// возвращает true если ОК, false если упало (без re-throw).
export function useMutationToast() {
  return useCallback(
    async <T>(
      promise: Promise<T>,
      messages: { success?: string; error?: string } = {},
    ): Promise<T | null> => {
      try {
        const r = await promise;
        if (messages.success) toast.success(messages.success);
        return r;
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        toast.error(messages.error ? `${messages.error}: ${msg}` : msg);
        return null;
      }
    },
    [],
  );
}
