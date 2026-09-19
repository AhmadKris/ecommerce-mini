import { z } from "zod";

// Mirrors CreateAddressRequest/UpdateAddressRequest validation on the
// backend (full replace on update, no pointer-field partial updates — see
// UpdateAddressRequest's comment on backend for why).
export const addressSchema = z.object({
  recipientName: z.string().min(2, "Nama penerima minimal 2 karakter").max(255),
  phone: z.string().min(5, "Nomor telepon minimal 5 karakter").max(30),
  addressLine: z.string().min(10, "Alamat minimal 10 karakter").max(500),
  city: z.string().min(1, "Kota wajib diisi").max(255),
  province: z.string().min(1, "Provinsi wajib diisi").max(255),
  postalCode: z.string().min(1, "Kode pos wajib diisi").max(20),
  isDefault: z.boolean(),
});
export type AddressFormValues = z.infer<typeof addressSchema>;
