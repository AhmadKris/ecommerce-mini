import { useMutation } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { LoginFormValues, RegisterFormValues } from "../schemas/auth";
import { useAuthStore } from "../store/auth-store";
import type { AuthTokens } from "../types/auth";

interface RegisteredUser {
  id: number;
  name: string;
  email: string;
}

/** Registers a new customer account. Does not log the user in by itself. */
export function useRegister() {
  return useMutation({
    mutationFn: (payload: RegisterFormValues) =>
      apiClient.post<RegisteredUser>("/auth/register", payload).then((response) => response.data),
  });
}

/** Logs in and stores the resulting session in the auth store. */
export function useLogin() {
  const setSession = useAuthStore((state) => state.setSession);

  return useMutation({
    mutationFn: (payload: LoginFormValues) =>
      apiClient.post<AuthTokens>("/auth/login", payload).then((response) => response.data),
    onSuccess: (tokens) => setSession(tokens),
  });
}
