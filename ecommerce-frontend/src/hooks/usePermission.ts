import { hasPermission } from "../lib/rbac";
import { useAuthStore } from "../store/auth-store";

/**
 * Returns whether the current session's access token carries `code`.
 * UI-level only — hide/show accordingly, but the backend remains the real
 * enforcement point (see Auth & RBAC di Frontend in .claude/CLAUDE.md).
 */
export function usePermission(code: string): boolean {
  const permissions = useAuthStore((state) => state.permissions);
  return hasPermission(permissions, code);
}
