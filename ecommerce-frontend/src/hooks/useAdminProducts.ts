import { useMutation, useQueryClient } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { Product } from "../types/product";
import type { ProductFormValues } from "../schemas/product";

function toRequestBody(values: ProductFormValues) {
  return {
    name: values.name,
    sku: values.sku,
    description: values.description,
    price: values.price,
    stock: values.stock,
    category_id: values.categoryId,
    image_url: values.imageUrl,
  };
}

/** Creates a product (admin, needs product:create). */
export function useCreateProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (values: ProductFormValues) =>
      apiClient.post<Product>("/admin/products", toRequestBody(values)).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["products"] }),
  });
}

/** Updates a product (admin, needs product:update). */
export function useUpdateProduct(id: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (values: ProductFormValues) =>
      apiClient.put<Product>(`/admin/products/${id}`, toRequestBody(values)).then((r) => r.data),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["products"] }),
  });
}
