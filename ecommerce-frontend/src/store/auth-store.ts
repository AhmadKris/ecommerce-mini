import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

import { decodeAccessToken } from "../lib/jwt";
import type { AuthTokens } from "../types/auth";

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  userId: number | null;
  roles: string[];
  permissions: string[];
  setSession: (tokens: AuthTokens) => void;
  clearSession: () => void;
}

/**
 * Session state (tokens, decoded identity/roles/permissions). Persisted to
 * sessionStorage rather than kept in-memory-only or in localStorage — see
 * "Known Issues" in .claude/CLAUDE.md for why: the backend returns
 * refresh_token in the JSON body (not an httpOnly cookie), so it's
 * necessarily JS-readable either way; sessionStorage at least clears on tab
 * close instead of persisting indefinitely like localStorage would.
 */
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      userId: null,
      roles: [],
      permissions: [],
      setSession: (tokens) => {
        const claims = decodeAccessToken(tokens.access_token);
        set({
          accessToken: tokens.access_token,
          refreshToken: tokens.refresh_token,
          userId: claims.user_id,
          roles: claims.roles,
          permissions: claims.permissions,
        });
      },
      clearSession: () =>
        set({
          accessToken: null,
          refreshToken: null,
          userId: null,
          roles: [],
          permissions: [],
        }),
    }),
    {
      name: "auth-storage",
      storage: createJSONStorage(() => sessionStorage),
    },
  ),
);
