import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Link } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useForgotPassword } from "../../hooks/useAuth";
import { forgotPasswordSchema, type ForgotPasswordFormValues } from "../../schemas/auth";
import { ApiError } from "../../types/api";

export function ForgotPassword() {
  const forgotPassword = useForgotPassword();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordFormValues>({ resolver: zodResolver(forgotPasswordSchema) });

  if (forgotPassword.isSuccess) {
    return (
      <main className="mx-auto max-w-sm px-6 py-16">
        <h1 className="text-heading-xl text-(--ink-primary)">Cek Email Kamu</h1>
        <p className="text-body-md text-(--ink-secondary) mt-2">{forgotPassword.data.message}</p>
        <Link to="/login" className="text-body-sm text-(--ink-link) mt-4 inline-block">
          Kembali ke halaman masuk
        </Link>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-sm px-6 py-16">
      <h1 className="text-heading-xl text-(--ink-primary)">Lupa Password</h1>
      <p className="text-body-sm text-(--ink-secondary) mt-2">
        Masukkan email akunmu, kami akan kirim instruksi untuk membuat password baru.
      </p>

      <form
        onSubmit={handleSubmit((values) => forgotPassword.mutate(values.email))}
        className="mt-6 flex flex-col gap-4"
        noValidate
      >
        <Input label="Email" type="email" autoComplete="email" error={errors.email?.message} {...register("email")} />

        {forgotPassword.isError && (
          <p role="alert" className="text-body-sm text-error-500">
            {forgotPassword.error instanceof ApiError
              ? forgotPassword.error.message
              : "Terjadi kesalahan, coba lagi."}
          </p>
        )}

        <Button type="submit" disabled={forgotPassword.isPending}>
          {forgotPassword.isPending ? "Mengirim..." : "Kirim Instruksi Reset"}
        </Button>
      </form>

      <p className="text-body-sm text-(--ink-secondary) mt-4">
        Ingat passwordmu?{" "}
        <Link to="/login" className="text-(--ink-link)">
          Masuk
        </Link>
      </p>
    </main>
  );
}
