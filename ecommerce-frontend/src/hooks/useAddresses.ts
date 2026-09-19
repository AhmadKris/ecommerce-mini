import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { Address } from "../types/address";

/** Lists the logged-in user's saved addresses. */
export function useAddresses() {
  return useQuery({
    queryKey: ["addresses"],
    queryFn: () => apiClient.get<{ items: Address[] }>("/profile/addresses").then((response) => response.data.items),
  });
}

export interface AddressInput {
  recipient_name: string;
  phone: string;
  address_line: string;
  city: string;
  province: string;
  postal_code: string;
  is_default: boolean;
}

/** Saves a new address. */
export function useCreateAddress() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: AddressInput) => apiClient.post<Address>("/profile/addresses", input).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["addresses"] }),
  });
}

/** Replaces an existing address (full replace, no partial update). */
export function useUpdateAddress() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...input }: AddressInput & { id: number }) =>
      apiClient.put<Address>(`/profile/addresses/${id}`, input).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["addresses"] }),
  });
}

/** Deletes an address. */
export function useDeleteAddress() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`/profile/addresses/${id}`),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["addresses"] }),
  });
}
