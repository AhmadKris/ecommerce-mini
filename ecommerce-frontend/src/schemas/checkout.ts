import { z } from "zod";

// Mirrors CheckoutRequest.ShippingAddress validation on the backend
// (min=10,max=500).
export const checkoutSchema = z.object({
  shippingAddress: z
    .string()
    .min(10, "Alamat pengiriman minimal 10 karakter")
    .max(500, "Alamat pengiriman maksimal 500 karakter"),
});
export type CheckoutFormValues = z.infer<typeof checkoutSchema>;
