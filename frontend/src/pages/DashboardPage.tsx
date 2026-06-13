import { AllocationChart } from "../components/charts/AllocationChart";
import { PerformanceChart } from "../components/charts/PerformanceChart";
import { Link } from "react-router-dom";

import { Button, Card, SecondaryButton } from "../components/ui/primitives";
import { IncomeCard } from "../components/widgets/IncomeCard";
import { SummaryCards } from "../components/widgets/SummaryCards";
import { TopMovers } from "../components/widgets/TopMovers";
import { useSessionStore } from "../features/auth/store";
import { usePortfolio, usePortfolioPerformance, usePortfolioSummary } from "../features/portfolio/hooks";

export function DashboardPage() {
  const currency = useSessionStore((state) => state.preferredCurrency);
  const summaryQuery = usePortfolioSummary(currency);
  const portfolioQuery = usePortfolio(currency);
  const performanceQuery = usePortfolioPerformance(currency);

  if (summaryQuery.isLoading || portfolioQuery.isLoading || performanceQuery.isLoading) {
    return <Card>Loading dashboard...</Card>;
  }

  if (summaryQuery.error || portfolioQuery.error || performanceQuery.error || !summaryQuery.data || !portfolioQuery.data || !performanceQuery.data) {
    return <Card>Unable to load dashboard data.</Card>;
  }

  return (
    <div className="space-y-6">
      <Card>
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p className="eyebrow">Overview</p>
            <h2 className="mt-2 text-2xl font-semibold">Your portfolio command center</h2>
            <p className="mt-2 text-sm text-slate-400">
              Start by buying a stock or crypto asset from the Transactions page.
            </p>
          </div>
          <div className="flex flex-wrap gap-3">
            <Link to="/transactions">
              <Button>Add order</Button>
            </Link>
            <Link to="/portfolio">
              <SecondaryButton>View positions</SecondaryButton>
            </Link>
          </div>
        </div>
      </Card>
      <SummaryCards summary={summaryQuery.data} />
      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <Card>
          <div className="mb-4">
            <h2 className="text-xl font-semibold">Portfolio evolution</h2>
            <p className="text-sm text-slate-400">Historical invested capital over time.</p>
          </div>
          <PerformanceChart points={performanceQuery.data.data} currency={currency} />
        </Card>
        <Card>
          <div className="mb-4">
            <h2 className="text-xl font-semibold">Allocation</h2>
            <p className="text-sm text-slate-400">Current allocation by asset.</p>
          </div>
          <AllocationChart positions={portfolioQuery.data.positions} />
        </Card>
      </div>
      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <TopMovers positions={portfolioQuery.data.positions} currency={currency} />
        <IncomeCard summary={summaryQuery.data} />
      </div>
    </div>
  );
}
