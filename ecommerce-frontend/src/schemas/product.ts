import { z } from "zod";

// Mirrors CreateProductRequest/UpdateProductRequest validation on the
// backend. Price/stock skip a "required" style check the same way the
// backend DTO does — 0 is a legitimate value, not "missing" (see
// CreateProductRequest's comment on backend).
export const productSchema = z.object({
  name: z.string().min(2, "Nama produk minimal 2 karakter").max(255, "Nama produk maksimal 255 karakter"),
  sku: z.string().min(1, "SKU wajib diisi").max(64, "SKU maksimal 64 karakter"),
  description: z.string().max(5000, "Deskripsi maksimal 5000 karakter").optional().default(""),
  price: z.coerce.number().gte(0, "Harga tidak boleh negatif"),
  stock: z.coerce.number().int("Stok harus bilangan bulat").gte(0, "Stok tidak boleh negatif"),
  categoryId: z.coerce.number().int().positive("Pilih kategori"),
  imageUrl: z
    .string()
    .url("URL gambar tidak valid")
    .max(500)
    .optional()
    .or(z.literal(""))
    .default(""),
});
// Input: raw string values straight out of HTML number inputs, before
// z.coerce runs. Output: parsed numbers, what onSubmit actually receives.
// Needed as two distinct types because of the coerce+default combination —
// see ProductForm's useForm call.
export type ProductFormInput = z.input<typeof productSchema>;
export type ProductFormValues = z.output<typeof productSchema>;
