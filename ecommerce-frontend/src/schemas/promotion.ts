import { z } from "zod";

// Mirrors CreatePromotionRequest/UpdatePromotionRequest validation on the
// backend. value/minimum_purchase use z.number() (not .min(1)) since 0 is a
// legitimate minimum_purchase, per the "explicit zero" convention.
export const promotionSchema = z.object({
  code: z.string().min(3, "Kode promo minimal 3 karakter"),
  type: z.enum(["percentage", "fixed"]),
  value: z.coerce.number().positive("Nilai diskon harus lebih dari 0"),
  minimum_purchase: z.coerce.number().min(0, "Minimum pembelian tidak boleh negatif"),
  usage_limit: z.coerce.number().int().min(1, "Batas pemakaian minimal 1"),
  starts_at: z.string().min(1, "Tanggal mulai wajib diisi"),
  ends_at: z.string().min(1, "Tanggal berakhir wajib diisi"),
});
// Input: raw string values straight out of HTML number inputs, before
// z.coerce runs. Output: parsed numbers, what onSubmit actually receives —
// same two-type split as ProductFormInput/ProductFormValues, needed for the
// same reason (z.coerce.number() fields).
export type PromotionFormInput = z.input<typeof promotionSchema>;
export type PromotionFormValues = z.output<typeof promotionSchema>;

export const promotionUpdateSchema = promotionSchema.omit({ code: true }).extend({
  status: z.enum(["active", "inactive"]),
});
export type PromotionUpdateFormInput = z.input<typeof promotionUpdateSchema>;
export type PromotionUpdateFormValues = z.output<typeof promotionUpdateSchema>;
