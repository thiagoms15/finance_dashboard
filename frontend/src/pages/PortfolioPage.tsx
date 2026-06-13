import { Link } from "react-router-dom";

import { Card, EmptyState } from "../components/ui/primitives";
import { formatCurrency, formatPercent } from "../lib/format";
import { useSessionStore } from "../features/auth/store";
import { usePortfolio } from "../features/portfolio/hooks";

export function PortfolioPage() {
  const currency = useSessionStore((state) => state.preferredCurrency);
  const portfolioQuery = usePortfolio(currency);

  if (portfolioQuery.isLoading) {
    return <Card>Loading portfolio...</Card>;
  }

  if (portfolioQuery.error || !portfolioQuery.data) {
    return <Card>Unable to load portfolio.</Card>;
  }

  if (portfolioQuery.data.positions.length === 0) {
    return <EmptyState title="No positions" description="Create a BUY transaction to open your first position." />;
  }

  return (
    <Card>
      <div className="mb-4">
        <h2 className="text-xl font-semibold">Open positions</h2>
        <p className="text-sm text-slate-400">Current holdings and unrealized P/L.</p>
      </div>
      <div className="overflow-x-auto">
        <table className="min-w-full text-left text-sm">
          <thead className="text-slate-400">
            <tr>
              <th className="pb-3">Asset</th>
              <th className="pb-3">Exchange</th>
              <th className="pb-3">Quantity</th>
              <th className="pb-3">Avg cost</th>
              <th className="pb-3">Current price</th>
              <th className="pb-3">Current value</th>
              <th className="pb-3">P/L</th>
            </tr>
          </thead>
          <tbody>
            {portfolioQuery.data.positions.map((position) => (
              <tr key={position.asset.id} className="border-t border-slate-800/70">
                <td className="py-4">
                  <Link className="font-medium text-sky-300" to={`/portfolio/${position.asset.id}`}>
                    {position.asset.symbol}
                  </Link>
                  <p className="text-xs text-slate-500">{position.asset.name}</p>
                </td>
                <td className="py-4">{position.asset.exchange}</td>
                <td className="py-4">{position.quantity}</td>
                <td className="py-4">{formatCurrency(position.averageCost, currency)}</td>
                <td className="py-4">
                  {formatCurrency(position.currentPrice, position.currentCurrency || currency)}
                </td>
                <td className="py-4">{formatCurrency(position.currentValue, currency)}</td>
                <td className="py-4">
                  {formatCurrency(position.unrealizedPL, currency)}
                  <p className="text-xs text-slate-500">{formatPercent(position.unrealizedPLPct)}</p>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
