import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { Category } from "../types/product";

/** Lists every category (public, unpaginated — the catalog is small). */
export function useCategories() {
  return useQuery({
    queryKey: ["categories"],
    queryFn: () => apiClient.get<{ items: Category[] }>("/categories").then((response) => response.data.items),
  });
}

/** Creates a category (admin, needs category:create). */
export function useCreateCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => apiClient.post<Category>("/admin/categories", { name }).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["categories"] }),
  });
}

/** Renames a category (admin, needs category:update). */
export function useUpdateCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, name }: { id: number; name: string }) =>
      apiClient.put<Category>(`/admin/categories/${id}`, { name }).then((r) => r.data),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["categories"] });
      // A rename changes the category name shown on every product that
      // references it.
      void queryClient.invalidateQueries({ queryKey: ["products"] });
    },
  });
}

/** Deletes a category (admin, needs category:delete). Backend rejects with
 * 409 if any product still references it. */
export function useDeleteCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`/admin/categories/${id}`),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["categories"] }),
  });
}
