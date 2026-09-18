import { Navigate, Outlet } from "react-router-dom";

import { useAuthStore } from "../store/auth-store";

/** Redirects to /login when there's no session. Wraps route(s) via <Outlet />. */
export function ProtectedRoute() {
  const isAuthenticated = useAuthStore((state) => state.accessToken !== null);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}
