import { useQuery } from "@tanstack/react-query";

import { api } from "../../lib/api/client";

export function usePortfolio(currency: string) {
  return useQuery({
    queryKey: ["portfolio", currency],
    queryFn: () => api.portfolio(currency)
  });
}

export function usePortfolioSummary(currency: string) {
  return useQuery({
    queryKey: ["portfolio-summary", currency],
    queryFn: () => api.portfolioSummary(currency)
  });
}

export function usePortfolioPerformance(currency: string) {
  return useQuery({
    queryKey: ["portfolio-performance", currency],
    queryFn: () => api.portfolioPerformance(currency)
  });
}
