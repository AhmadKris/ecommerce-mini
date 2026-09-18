import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useLocation, useNavigate, useParams } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useCategories } from "../../hooks/useCategories";
import { useCreateProduct, useUpdateProduct } from "../../hooks/useAdminProducts";
import { productSchema, type ProductFormInput, type ProductFormValues } from "../../schemas/product";
import { ApiError } from "../../types/api";
import type { Product } from "../../types/product";

/**
 * Handles both create (no :id in the route) and edit. Edit reads the
 * product from router state (passed by AdminProducts's "Ubah" link) rather
 * than a GET-by-id call — the backend has no such endpoint yet (only
 * GET /products/:slug for the public detail page, see Known Issues), and
 * the admin list already has the full product in memory.
 */
export function ProductForm() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const editingProduct = (location.state as { product?: Product } | null)?.product;
  const isEditMode = id !== undefined;

  const { data: categories } = useCategories();
  const createProduct = useCreateProduct();
  const updateProduct = useUpdateProduct(Number(id));
  const mutation = isEditMode ? updateProduct : createProduct;

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ProductFormInput, unknown, ProductFormValues>({
    resolver: zodResolver(productSchema),
    defaultValues: editingProduct
      ? {
          name: editingProduct.name,
          description: editingProduct.description,
          price: editingProduct.price,
          stock: editingProduct.stock,
          categoryId: editingProduct.category_id,
          imageUrl: editingProduct.image_url,
        }
      : undefined,
  });

  if (isEditMode && !editingProduct) {
    return (
      <main className="mx-auto max-w-xl px-6 py-16">
        <p className="text-body-md text-(--ink-secondary)">
          Data produk tidak tersedia — buka form ini lewat tombol "Ubah" di daftar produk, bukan
          langsung lewat URL.
        </p>
      </main>
    );
  }

  function onSubmit(values: ProductFormValues) {
    mutation.mutate(values, { onSuccess: () => navigate("/admin/products") });
  }

  return (
    <main className="mx-auto max-w-xl px-6 py-16">
      <h1 className="text-heading-xl text-(--ink-primary) mb-6">
        {isEditMode ? "Ubah Produk" : "Tambah Produk"}
      </h1>

      <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4" noValidate>
        <Input label="Nama produk" error={errors.name?.message} {...register("name")} />

        <div className="flex flex-col gap-1">
          <label htmlFor="description" className="text-label-md text-(--ink-primary)">
            Deskripsi
          </label>
          <textarea
            id="description"
            rows={4}
            className="text-body-md rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
            {...register("description")}
          />
          {errors.description && (
            <p className="text-body-sm text-error-500">{errors.description.message}</p>
          )}
        </div>

        <Input label="Harga (Rp)" type="number" error={errors.price?.message} {...register("price")} />
        <Input label="Stok" type="number" error={errors.stock?.message} {...register("stock")} />

        <div className="flex flex-col gap-1">
          <label htmlFor="categoryId" className="text-label-md text-(--ink-primary)">
            Kategori
          </label>
          <select
            id="categoryId"
            className="text-body-md rounded-md border border-(--border-default) px-3 py-2.5 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
            {...register("categoryId")}
          >
            <option value="">Pilih kategori</option>
            {categories?.map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </select>
          {errors.categoryId && <p className="text-body-sm text-error-500">{errors.categoryId.message}</p>}
        </div>

        <Input label="URL gambar (opsional)" error={errors.imageUrl?.message} {...register("imageUrl")} />

        {mutation.isError && (
          <p role="alert" className="text-body-sm text-error-500">
            {mutation.error instanceof ApiError ? mutation.error.message : "Gagal menyimpan produk."}
          </p>
        )}

        <Button type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? "Menyimpan..." : "Simpan"}
        </Button>
      </form>
    </main>
  );
}
