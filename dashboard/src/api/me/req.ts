// Запросы /v1/me — fetcher функции + TanStack hooks колокированы с request слоем.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../lib/http";
import type { MeRes, ProfileRes } from "./res";

// --- Query keys ---

export const meKeys = {
  me: ["me"] as const,
  profile: ["me.profile"] as const,
};

// --- Fetchers ---

export const fetchMe = () => http.get<MeRes>("/v1/me").then((r) => r.data);

export const fetchProfile = () => http.get<ProfileRes>("/v1/me/").then((r) => r.data);

export async function deleteMyData(): Promise<void> {
  await http.delete("/v1/me/data");
  localStorage.removeItem("eop_token");
  localStorage.removeItem("eop_user_id");
}

// --- Hooks ---

export const useMe = () => useQuery({ queryKey: meKeys.me, queryFn: fetchMe });
export const useProfile = () => useQuery({ queryKey: meKeys.profile, queryFn: fetchProfile });

export function useDeleteMyData() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteMyData,
    onSuccess: () => qc.clear(),
  });
}
