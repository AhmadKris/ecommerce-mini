import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { useLogin } from "../../hooks/useAuth";
import { loginSchema, type LoginFormValues } from "../../schemas/auth";
import { ApiError } from "../../types/api";

export function Login() {
  const navigate = useNavigate();
  const login = useLogin();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({ resolver: zodResolver(loginSchema) });

  function onSubmit(values: LoginFormValues) {
    login.mutate(values, { onSuccess: () => navigate("/") });
  }

  return (
    <main className="mx-auto max-w-sm px-6 py-16">
      <h1 className="text-heading-xl text-(--ink-primary)">Masuk</h1>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 flex flex-col gap-4" noValidate>
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
          autoComplete="current-password"
          error={errors.password?.message}
          {...register("password")}
        />

        {login.isError && (
          <p role="alert" className="text-body-sm text-error-500">
            {login.error instanceof ApiError
              ? login.error.message
              : "Terjadi kesalahan, coba lagi."}
          </p>
        )}

        <Button type="submit" disabled={login.isPending}>
          {login.isPending ? "Memproses..." : "Masuk"}
        </Button>
      </form>

      <p className="text-body-sm text-(--ink-secondary) mt-4">
        Belum punya akun?{" "}
        <Link to="/register" className="text-(--ink-link)">
          Daftar
        </Link>
      </p>
    </main>
  );
}
