import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { PaginatedResponse } from "../types/api";
import type { Order } from "../types/order";

/** Lists every order across all users (admin, needs order:read_all) —
 * unlike useOrders, which is scoped to the logged-in user's own history. */
export function useAdminOrders(page = 1) {
  return useQuery({
    queryKey: ["admin", "orders", { page }],
    queryFn: () =>
      apiClient
        .get<PaginatedResponse<Order>>("/admin/orders", { params: { page } })
        .then((response) => response.data),
  });
}

/** Fetches one order with its items, for the admin order detail page. */
export function useAdminOrder(id: number) {
  return useQuery({
    queryKey: ["admin", "orders", id],
    queryFn: () => apiClient.get<Order>(`/admin/orders/${id}`).then((response) => response.data),
    enabled: Number.isFinite(id),
  });
}

/** Transitions an order to a new status (admin, needs order:update_status).
 * Backend rejects invalid transitions with 409 — this hook doesn't validate
 * that itself, it just surfaces whatever the backend decides. */
export function useUpdateOrderStatus(id: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (status: string) =>
      apiClient.patch<Order>(`/admin/orders/${id}/status`, { status }).then((r) => r.data),
    // Prefix match invalidates both the detail query (["admin","orders",id])
    // and the list query (["admin","orders",{page}]) in one call.
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin", "orders"] }),
  });
}
