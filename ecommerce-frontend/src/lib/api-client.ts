import axios, { type AxiosRequestConfig, type AxiosError } from "axios";

import { config } from "./config";
import { useAuthStore } from "../store/auth-store";
import { ApiError, type ApiErrorEnvelope } from "../types/api";
import type { AuthTokens } from "../types/auth";

/**
 * The single Axios instance for the app. Request interceptor attaches the
 * bearer token; response interceptor unwraps the backend's
 * {success, data, message, meta} envelope so callers just read
 * `response.data` (already the inner data), and normalizes
 * {success: false, error} into a thrown ApiError.
 */
export const apiClient = axios.create({
  baseURL: config.apiBaseUrl,
});

apiClient.interceptors.request.use((requestConfig) => {
  const { accessToken } = useAuthStore.getState();
  if (accessToken) {
    requestConfig.headers.set("Authorization", `Bearer ${accessToken}`);
  }
  return requestConfig;
});

interface RetryableRequestConfig extends AxiosRequestConfig {
  _retriedAfterRefresh?: boolean;
}

// A 401 from these public auth endpoints means "wrong credentials" or "bad
// token", never "session expired" — attempting a refresh here would be a
// no-op at best and, if a stale session happens to still be in the store
// (e.g. testing a wrong password on /login while already logged in from
// another tab), would silently rotate that unrelated session's refresh
// token as a side effect. Found via a real browser check, not just unit
// tests — see Known Issues in .claude/CLAUDE.md.
const PUBLIC_AUTH_PATHS = ["/auth/login", "/auth/register", "/auth/refresh"];

apiClient.interceptors.response.use(
  (response) => {
    response.data = response.data?.data ?? response.data;
    return response;
  },
  async (error: AxiosError<ApiErrorEnvelope>) => {
    const originalRequest = error.config as RetryableRequestConfig | undefined;
    const isPublicAuthCall = PUBLIC_AUTH_PATHS.some((path) => originalRequest?.url === path);

    if (
      error.response?.status === 401 &&
      originalRequest &&
      !originalRequest._retriedAfterRefresh &&
      !isPublicAuthCall
    ) {
      originalRequest._retriedAfterRefresh = true;
      const { refreshToken, setSession, clearSession } = useAuthStore.getState();

      if (refreshToken) {
        try {
          const refreshResponse = await axios.post<{ data: AuthTokens }>(
            `${config.apiBaseUrl}/auth/refresh`,
            { refresh_token: refreshToken },
          );
          setSession(refreshResponse.data.data);
          return apiClient(originalRequest);
        } catch {
          clearSession();
        }
      } else {
        clearSession();
      }
    }

    return Promise.reject(toApiError(error));
  },
);

function toApiError(error: AxiosError<ApiErrorEnvelope>): ApiError {
  const envelope = error.response?.data;
  if (envelope && !envelope.success) {
    return new ApiError(
      envelope.error.message,
      envelope.error.code,
      envelope.error.details ?? [],
      error.response?.status,
    );
  }
  return new ApiError(
    "Tidak dapat terhubung ke server",
    "NETWORK_ERROR",
    [],
    error.response?.status,
  );
}
