import { useMutation } from "@tanstack/react-query";

import { api } from "../../lib/api/client";
import { useSessionStore } from "./store";

export function useLogin() {
  const setAuth = useSessionStore((state) => state.setAuth);

  return useMutation({
    mutationFn: api.login,
    onSuccess: (data) => setAuth(data.accessToken, data.user, data.expiresIn)
  });
}

export function useRegister() {
  const setAuth = useSessionStore((state) => state.setAuth);

  return useMutation({
    mutationFn: api.register,
    onSuccess: (data) => setAuth(data.accessToken, data.user, data.expiresIn)
  });
}

export function useRequestPasswordReset() {
  return useMutation({
    mutationFn: api.requestPasswordReset
  });
}

export function useConfirmPasswordReset() {
  return useMutation({
    mutationFn: api.confirmPasswordReset
  });
}
