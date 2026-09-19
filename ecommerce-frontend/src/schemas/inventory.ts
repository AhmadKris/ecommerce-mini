import { z } from "zod";

// Mirrors AdjustInventoryRequest validation on the backend. Quantity uses
// z.coerce.number() (not .min(1)) since 0 is valid for "correction" (e.g.
// reconciling a stock opname that finds zero units left).
export const inventoryAdjustmentSchema = z.object({
  productId: z.coerce.number().int().positive("Pilih produk"),
  type: z.enum(["in", "out", "correction"]),
  quantity: z.coerce.number().int().gte(0, "Kuantitas tidak boleh negatif"),
  reason: z.string().min(1, "Alasan wajib diisi").max(500),
  reference: z.string().max(255).optional(),
});
// Input/output split needed because of z.coerce.number() fields — same
// reason as ProductFormInput/ProductFormValues.
export type InventoryAdjustmentFormInput = z.input<typeof inventoryAdjustmentSchema>;
export type InventoryAdjustmentFormValues = z.output<typeof inventoryAdjustmentSchema>;
