import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useLogin, useRegister } from "../../hooks/useAuth";
import { registerSchema, type RegisterFormValues } from "../../schemas/auth";
import { ApiError } from "../../types/api";

export function Register() {
  const navigate = useNavigate();
  const registerAccount = useRegister();
  const login = useLogin();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormValues>({ resolver: zodResolver(registerSchema) });

  function onSubmit(values: RegisterFormValues) {
    registerAccount.mutate(values, {
      // Auto-login right after registration — the backend doesn't return
      // tokens from /register, only the created user, so a second call is
      // unavoidable. Saves the user from re-typing credentials on /login.
      onSuccess: () => {
        login.mutate(
          { email: values.email, password: values.password },
          { onSuccess: () => navigate("/") },
        );
      },
    });
  }

  const pending = registerAccount.isPending || login.isPending;
  const failure = registerAccount.error ?? login.error;

  return (
    <main className="mx-auto max-w-sm px-6 py-16">
      <h1 className="text-heading-xl text-(--ink-primary)">Daftar</h1>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 flex flex-col gap-4" noValidate>
        <Input
          label="Nama"
          autoComplete="name"
          error={errors.name?.message}
          {...register("name")}
        />
        <Input
          label="Email"
          type="email"
          autoComplete="email"
          error={errors.email?.message}
          {...register("email")}
        />
        <Input
          label="Password"
          type="password"
          autoComplete="new-password"
          error={errors.password?.message}
          {...register("password")}
        />

        {failure && (
          <p role="alert" className="text-body-sm text-error-500">
            {failure instanceof ApiError ? failure.message : "Terjadi kesalahan, coba lagi."}
          </p>
        )}

        <Button type="submit" disabled={pending}>
          {pending ? "Memproses..." : "Daftar"}
        </Button>
      </form>

      <p className="text-body-sm text-(--ink-secondary) mt-4">
        Sudah punya akun?{" "}
        <Link to="/login" className="text-(--ink-link)">
          Masuk
        </Link>
      </p>
    </main>
  );
}
