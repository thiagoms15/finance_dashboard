import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api } from "../../lib/api/client";
import { useSessionStore } from "../auth/store";

function useCurrency() {
  return useSessionStore((state) => state.preferredCurrency);
}

export function useDividends() {
  return useQuery({
    queryKey: ["dividends"],
    queryFn: api.dividends
  });
}

export function useDividendMutations() {
  const queryClient = useQueryClient();
  const currency = useCurrency();

  const invalidate = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["dividends"] }),
      queryClient.invalidateQueries({ queryKey: ["portfolio", currency] }),
      queryClient.invalidateQueries({ queryKey: ["portfolio-summary", currency] })
    ]);
  };

  return {
    create: useMutation({
      mutationFn: api.createDividend,
      onSuccess: invalidate
    }),
    update: useMutation({
      mutationFn: ({ id, body }: { id: string; body: Record<string, string> }) =>
        api.updateDividend(id, body),
      onSuccess: invalidate
    }),
    remove: useMutation({
      mutationFn: api.deleteDividend,
      onSuccess: invalidate
    })
  };
}
