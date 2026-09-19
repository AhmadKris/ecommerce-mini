import { z } from "zod";

// Mirrors CreateReviewRequest validation on the backend.
export const reviewSchema = z.object({
  rating: z.coerce.number().int().min(1, "Rating minimal 1").max(5, "Rating maksimal 5"),
  title: z.string().min(3, "Judul minimal 3 karakter").max(255, "Judul maksimal 255 karakter"),
  body: z.string().min(10, "Ulasan minimal 10 karakter").max(2000, "Ulasan maksimal 2000 karakter"),
});
// Same z.input/z.output split as ProductFormInput/ProductFormValues — see
// that schema's comment (needed because of the z.coerce.number() field).
export type ReviewFormInput = z.input<typeof reviewSchema>;
export type ReviewFormValues = z.output<typeof reviewSchema>;
