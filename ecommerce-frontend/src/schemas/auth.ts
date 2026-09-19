import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().email("Email tidak valid"),
  password: z.string().min(1, "Password wajib diisi"),
});
export type LoginFormValues = z.infer<typeof loginSchema>;

/**
 * Mirrors the backend's RegisterRequest validation (see model.RegisterRequest
 * and validatePasswordStrength in the Go service) so a weak password is
 * caught client-side instead of round-tripping to the server first.
 */
export const registerSchema = z.object({
  name: z.string().min(2, "Nama minimal 2 karakter").max(255, "Nama maksimal 255 karakter"),
  email: z.string().email("Email tidak valid").max(255),
  password: z
    .string()
    .min(8, "Password minimal 8 karakter")
    .max(72, "Password maksimal 72 karakter")
    .refine((value) => /[A-Z]/.test(value), "Password harus mengandung huruf besar")
    .refine((value) => /[a-z]/.test(value), "Password harus mengandung huruf kecil")
    .refine((value) => /[0-9]/.test(value), "Password harus mengandung angka"),
});
export type RegisterFormValues = z.infer<typeof registerSchema>;

export const forgotPasswordSchema = z.object({
  email: z.string().email("Email tidak valid"),
});
export type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

/** Mirrors registerSchema's password strength rules — the backend enforces
 * the same rule (validatePasswordStrength) for a reset password too. */
export const resetPasswordSchema = z.object({
  newPassword: z
    .string()
    .min(8, "Password minimal 8 karakter")
    .max(72, "Password maksimal 72 karakter")
    .refine((value) => /[A-Z]/.test(value), "Password harus mengandung huruf besar")
    .refine((value) => /[a-z]/.test(value), "Password harus mengandung huruf kecil")
    .refine((value) => /[0-9]/.test(value), "Password harus mengandung angka"),
});
export type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>;
