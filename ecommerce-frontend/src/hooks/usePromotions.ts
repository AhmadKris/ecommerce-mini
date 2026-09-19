import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { Promotion } from "../types/promotion";

/** Lists every promotion, unpaginated (admin, needs promotion:manage). */
export function usePromotions() {
  return useQuery({
    queryKey: ["promotions"],
    queryFn: () => apiClient.get<{ items: Promotion[] }>("/admin/promotions").then((response) => response.data.items),
  });
}

export interface PromotionInput {
  code?: string;
  type: Promotion["type"];
  value: number;
  minimum_purchase: number;
  usage_limit: number;
  starts_at: string;
  ends_at: string;
  status?: Promotion["status"];
}

/** Creates a promotion code (admin, needs promotion:manage). */
export function useCreatePromotion() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PromotionInput) => apiClient.post<Promotion>("/admin/promotions", input).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["promotions"] }),
  });
}

/** Updates a promotion's terms — code is not editable, see backend rationale. */
export function useUpdatePromotion() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...input }: PromotionInput & { id: number }) =>
      apiClient.put<Promotion>(`/admin/promotions/${id}`, input).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["promotions"] }),
  });
}

/** Deletes a promotion (admin, needs promotion:manage). */
export function useDeletePromotion() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`/admin/promotions/${id}`),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["promotions"] }),
  });
}
