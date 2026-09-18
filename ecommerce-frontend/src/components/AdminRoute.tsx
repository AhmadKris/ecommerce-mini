import { Navigate, Outlet } from "react-router-dom";

import { useAuthStore } from "../store/auth-store";

/**
 * Gates admin routes on the "admin" role from the decoded JWT. UI-level
 * only, same caveat as usePermission — the backend's RequirePermission
 * middleware is the real enforcement point, this just avoids showing a
 * customer an admin screen full of 403s.
 */
export function AdminRoute() {
  const isAuthenticated = useAuthStore((state) => state.accessToken !== null);
  const isAdmin = useAuthStore((state) => state.roles.includes("admin"));

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  if (!isAdmin) {
    return <Navigate to="/" replace />;
  }
  return <Outlet />;
}
