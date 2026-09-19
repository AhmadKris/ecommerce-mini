import { useMutation } from "@tanstack/react-query";

import { apiClient } from "../lib/api-client";
import type { LoginFormValues, RegisterFormValues } from "../schemas/auth";
import { useAuthStore } from "../store/auth-store";
import type { AuthTokens } from "../types/auth";

interface MessageResponse {
  message: string;
}

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

/** Requests a password reset. The backend always replies with the same
 * generic message whether or not the email is registered — see
 * AuthService.ForgotPassword on the backend for why. */
export function useForgotPassword() {
  return useMutation({
    mutationFn: (email: string) =>
      apiClient.post<MessageResponse>("/auth/forgot-password", { email }).then((response) => response.data),
  });
}

/** Redeems a reset token (from the forgot-password email/log) for a new
 * password. The token is single-use — a second call with the same token
 * fails even if the first succeeded. */
export function useResetPassword() {
  return useMutation({
    mutationFn: ({ token, newPassword }: { token: string; newPassword: string }) =>
      apiClient
        .post<MessageResponse>("/auth/reset-password", { token, new_password: newPassword })
        .then((response) => response.data),
  });
}
