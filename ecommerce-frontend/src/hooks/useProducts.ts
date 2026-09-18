import { useQuery } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { PaginatedResponse } from "../types/api";
import type { Product } from "../types/product";

export interface ProductFilters {
  category?: string;
  search?: string;
  page?: number;
  limit?: number;
}

/** Lists products for GET /products, filtered/paginated per `filters`. */
export function useProducts(filters: ProductFilters = {}) {
  return useQuery({
    queryKey: ["products", filters],
    queryFn: () =>
      apiClient
        .get<PaginatedResponse<Product>>("/products", { params: filters })
        .then((response) => response.data),
  });
}

/** Fetches a single product by slug for GET /products/:slug. */
export function useProduct(slug: string) {
  return useQuery({
    queryKey: ["products", slug],
    queryFn: () => apiClient.get<Product>(`/products/${slug}`).then((response) => response.data),
    enabled: slug.length > 0,
  });
}
