import axios, { type AxiosError } from "axios";
import { clearSession, getToken } from "../lib/session-storage";

const BASE = import.meta.env.VITE_BACKEND_URL || "https://eop-api.rysdavletov.org";

export const AUTH_FAILED_EVENT = "eop:auth-failed";

function emitAuthFailed() {
  clearSession();
  window.dispatchEvent(new CustomEvent(AUTH_FAILED_EVENT));
}

export const http = axios.create({
  baseURL: BASE,
  headers: { "Content-Type": "application/json" },
});

http.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: string; code?: string }>) => {
    if (error.response?.status === 401) {
      emitAuthFailed();
    }
    const serverMsg = error.response?.data?.error;
    if (serverMsg) {
      const e = new Error(serverMsg) as Error & { code?: string; status?: number };
      e.code = error.response?.data?.code;
      e.status = error.response?.status;
      return Promise.reject(e);
    }
    return Promise.reject(error);
  },
);
