import { z } from "zod";

// Mirrors CreateCategoryRequest/UpdateCategoryRequest validation on the
// backend (min=2,max=255).
export const categorySchema = z.object({
  name: z.string().min(2, "Nama kategori minimal 2 karakter").max(255, "Nama kategori maksimal 255 karakter"),
});
export type CategoryFormValues = z.infer<typeof categorySchema>;
