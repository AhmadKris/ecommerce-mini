import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { Customer } from "../types/customer";
import type { Role } from "../types/role";

/**
 * Fetches all available roles (admin, needs user:manage). Used by the role
 * management UI to populate the checkbox list.
 */
export function useRoles() {
  return useQuery({
    queryKey: ["admin", "roles"],
    queryFn: () => apiClient.get<Role[]>("/admin/roles").then((r) => r.data),
  });
}

/**
 * Replaces the full role set for a customer (admin, needs user:manage).
 * On success, customer queries are invalidated so the UI reflects the new roles.
 */
export function useUpdateUserRoles(customerId?: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (args: string[] | { userId: number; roleNames: string[] }) => {
      const targetId = customerId ?? (args as { userId: number; roleNames: string[] }).userId;
      const roleNames = Array.isArray(args) ? args : args.roleNames;
      return apiClient
        .put<Customer>(`/admin/customers/${targetId}/roles`, { role_names: roleNames })
        .then((r) => r.data);
    },
    onSuccess: (_, variables) => {
      const targetId = customerId ?? (variables as { userId: number; roleNames: string[] }).userId;
      void queryClient.invalidateQueries({ queryKey: ["admin", "customers", targetId] });
      void queryClient.invalidateQueries({ queryKey: ["admin", "customers"] });
    },
  });
}
