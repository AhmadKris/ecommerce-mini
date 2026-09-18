import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { PaginatedResponse } from "../types/api";
import type { Order } from "../types/order";

/** Lists the logged-in user's own order history, newest first. */
export function useOrders(page = 1) {
  return useQuery({
    queryKey: ["orders", { page }],
    queryFn: () =>
      apiClient.get<PaginatedResponse<Order>>("/orders", { params: { page } }).then((response) => response.data),
  });
}

interface CheckoutInput {
  shippingAddress: string;
  /**
   * Sent as Idempotency-Key so retrying the same checkout attempt (e.g. the
   * user clicking "Buat Pesanan" again after a dropped connection) replays
   * the first order instead of creating a duplicate — see backend's
   * middleware.Idempotency. Callers should reuse the same key across
   * retries of one attempt and generate a new one for a genuinely new
   * checkout.
   */
  idempotencyKey: string;
}

/** Checks out the current cart into an order. */
export function useCheckout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ shippingAddress, idempotencyKey }: CheckoutInput) =>
      apiClient
        .post<Order>(
          "/orders",
          { shipping_address: shippingAddress },
          { headers: { "Idempotency-Key": idempotencyKey } },
        )
        .then((response) => response.data),
    onSuccess: () => {
      // Checkout clears the cart and decrements product stock server-side —
      // invalidate all three rather than hand-patching each cache entry.
      void queryClient.invalidateQueries({ queryKey: ["cart"] });
      void queryClient.invalidateQueries({ queryKey: ["products"] });
      void queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
  });
}
