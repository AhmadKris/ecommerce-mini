import { useQuery } from "@tanstack/react-query";

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
