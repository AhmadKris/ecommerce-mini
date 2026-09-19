import { useQuery } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { PaginatedResponse } from "../types/api";
import type { Customer } from "../types/customer";

/** Lists registered accounts (admin, needs customer:read), optionally
 * filtered by a name/email search. */
export function useAdminCustomers(page = 1, search?: string) {
  return useQuery({
    queryKey: ["admin", "customers", { page, search }],
    queryFn: () =>
      apiClient
        .get<PaginatedResponse<Customer>>("/admin/customers", { params: { page, search: search || undefined } })
        .then((response) => response.data),
  });
}

/** Fetches one customer with roles and permissions, for the admin customer
 * detail page. */
export function useAdminCustomer(id: number) {
  return useQuery({
    queryKey: ["admin", "customers", id],
    queryFn: () => apiClient.get<Customer>(`/admin/customers/${id}`).then((response) => response.data),
    enabled: Number.isFinite(id),
  });
}
