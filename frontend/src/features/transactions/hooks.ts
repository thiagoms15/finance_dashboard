import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "../../lib/api/client";
import { useSessionStore } from "../auth/store";

function useCurrency() {
  return useSessionStore((state) => state.preferredCurrency);
}

export function useTransactions() {
  return useQuery({
    queryKey: ["transactions"],
    queryFn: api.transactions
  });
}

export function useTransactionMutations() {
  const queryClient = useQueryClient();
  const currency = useCurrency();

  const invalidate = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["transactions"] }),
      queryClient.invalidateQueries({ queryKey: ["portfolio", currency] }),
      queryClient.invalidateQueries({ queryKey: ["portfolio-summary", currency] }),
      queryClient.invalidateQueries({ queryKey: ["portfolio-performance", currency] })
    ]);
  };

  return {
    create: useMutation({
      mutationFn: api.createTransaction,
      onSuccess: invalidate
    }),
    update: useMutation({
      mutationFn: ({ id, body }: { id: string; body: Record<string, string> }) =>
        api.updateTransaction(id, body),
      onSuccess: invalidate
    }),
    remove: useMutation({
      mutationFn: api.deleteTransaction,
      onSuccess: invalidate
    })
  };
}
