import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Link, useSearchParams } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useResetPassword } from "../../hooks/useAuth";
import { resetPasswordSchema, type ResetPasswordFormValues } from "../../schemas/auth";
import { ApiError } from "../../types/api";

/**
 * Landing page for the link a forgot-password request produces
 * (`/reset-password?token=...`). There's no email provider wired into the
 * backend yet (see .claude/CLAUDE.md backend Known Issues) — the token is
 * currently only visible in the server log, so this page is reachable
 * today by pasting that token into the URL by hand. The UI itself is the
 * real, permanent piece; only the delivery mechanism is a placeholder.
 */
export function ResetPassword() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const resetPassword = useResetPassword();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetPasswordFormValues>({ resolver: zodResolver(resetPasswordSchema) });

  if (!token) {
    return (
      <main className="mx-auto max-w-sm px-6 py-16">
        <p role="alert" className="text-body-md text-error-500">
          Tautan reset password tidak valid — token tidak ditemukan.
        </p>
        <Link to="/forgot-password" className="text-body-sm text-(--ink-link) mt-2 inline-block">
          Minta tautan baru
        </Link>
      </main>
    );
  }

  if (resetPassword.isSuccess) {
    return (
      <main className="mx-auto max-w-sm px-6 py-16">
        <h1 className="text-heading-xl text-(--ink-primary)">Password Berhasil Direset</h1>
        <Link to="/login" className="text-body-sm text-(--ink-link) mt-4 inline-block">
          Masuk dengan password baru
        </Link>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-sm px-6 py-16">
      <h1 className="text-heading-xl text-(--ink-primary)">Buat Password Baru</h1>

      <form
        onSubmit={handleSubmit((values) => resetPassword.mutate({ token, newPassword: values.newPassword }))}
        className="mt-6 flex flex-col gap-4"
        noValidate
      >
        <Input
          label="Password Baru"
          type="password"
          autoComplete="new-password"
          error={errors.newPassword?.message}
          {...register("newPassword")}
        />

        {resetPassword.isError && (
          <p role="alert" className="text-body-sm text-error-500">
            {resetPassword.error instanceof ApiError ? resetPassword.error.message : "Terjadi kesalahan, coba lagi."}
          </p>
        )}

        <Button type="submit" disabled={resetPassword.isPending}>
          {resetPassword.isPending ? "Memproses..." : "Reset Password"}
        </Button>
      </form>
    </main>
  );
}
