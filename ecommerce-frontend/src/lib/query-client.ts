import { QueryClient } from "@tanstack/react-query";

import { ApiError } from "../types/api";

/**
 * A 4xx (not found, validation, forbidden, ...) means the request is wrong
 * in a way retrying won't fix — only retry once for everything else
 * (network errors, 5xx).
 */
export function shouldRetryQuery(failureCount: number, error: unknown): boolean {
  if (error instanceof ApiError && error.status !== undefined && error.status < 500) {
    return false;
  }
  return failureCount < 1;
}

/**
 * The single React Query client for the app. Server data always lives here
 * (see State Management in .claude/CLAUDE.md) — never copied into Zustand
 * or component state.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: shouldRetryQuery,
    },
  },
});
