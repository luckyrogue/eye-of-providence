import { useCallback } from "react";
import { toast } from "@eop/ui";
export function useMutationToast() {
  return useCallback(
    async <T>(
      promise: Promise<T>,
      messages: {
        success?: string;
        error?: string;
      } = {},
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
