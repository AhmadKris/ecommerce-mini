import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import { useAuthStore } from "../store/auth-store";
import type { Cart } from "../types/cart";

const CART_QUERY_KEY = ["cart"];

/** Fetches the logged-in user's cart. Disabled entirely when logged out. */
export function useCart() {
  const isAuthenticated = useAuthStore((state) => state.accessToken !== null);

  return useQuery({
    queryKey: CART_QUERY_KEY,
    queryFn: () => apiClient.get<Cart>("/cart").then((response) => response.data),
    enabled: isAuthenticated,
  });
}

/**
 * Every cart mutation endpoint returns the full updated cart (see
 * CartResponse on the backend) — writing that response straight into the
 * query cache is strictly fresher than invalidating and refetching, and
 * costs one request instead of two.
 */
function useSyncCartFromMutation() {
  const queryClient = useQueryClient();
  return (cart: Cart) => queryClient.setQueryData(CART_QUERY_KEY, cart);
}

export function useAddCartItem() {
  const syncCart = useSyncCartFromMutation();
  return useMutation({
    mutationFn: (payload: { productId: number; quantity: number }) =>
      apiClient
        .post<Cart>("/cart/items", { product_id: payload.productId, quantity: payload.quantity })
        .then((response) => response.data),
    onSuccess: syncCart,
  });
}

export function useUpdateCartItem() {
  const syncCart = useSyncCartFromMutation();
  return useMutation({
    mutationFn: (payload: { itemId: number; quantity: number }) =>
      apiClient
        .put<Cart>(`/cart/items/${payload.itemId}`, { quantity: payload.quantity })
        .then((response) => response.data),
    onSuccess: syncCart,
  });
}

export function useRemoveCartItem() {
  const syncCart = useSyncCartFromMutation();
  return useMutation({
    mutationFn: (itemId: number) =>
      apiClient.delete<Cart>(`/cart/items/${itemId}`).then((response) => response.data),
    onSuccess: syncCart,
  });
}
