import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { PaginatedResponse } from "../types/api";
import type { InventoryMovement, InventoryMovementType } from "../types/inventory";

export interface AdjustInventoryInput {
  product_id: number;
  type: InventoryMovementType;
  quantity: number;
  reason: string;
  reference?: string;
}

/** Applies a stock adjustment (admin, needs inventory:adjust). "in"/"out"
 * are deltas; "correction" sets stock to the given absolute value — see
 * backend's AdjustInventoryRequest for the same distinction. */
export function useAdjustInventory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: AdjustInventoryInput) =>
      apiClient.post<InventoryMovement>("/admin/inventory/adjustments", input).then((r) => r.data),
    onSuccess: (movement) => {
      void queryClient.invalidateQueries({ queryKey: ["inventory", "movements", movement.product_id] });
      // Adjusting stock changes the product's own stock field too.
      void queryClient.invalidateQueries({ queryKey: ["products"] });
    },
  });
}

/** Lists a product's stock movement history, newest first (admin, needs
 * inventory:read). */
export function useInventoryMovements(productId: number, page = 1) {
  return useQuery({
    queryKey: ["inventory", "movements", productId, { page }],
    queryFn: () =>
      apiClient
        .get<PaginatedResponse<InventoryMovement>>(`/admin/inventory/${productId}/movements`, { params: { page } })
        .then((response) => response.data),
    enabled: Number.isFinite(productId) && productId > 0,
  });
}
