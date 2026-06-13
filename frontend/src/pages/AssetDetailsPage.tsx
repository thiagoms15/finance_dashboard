import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";

import { PerformanceChart } from "../components/charts/PerformanceChart";
import { Card, EmptyState, SecondaryButton } from "../components/ui/primitives";
import { useAssetIcon } from "../features/assets/hooks";
import { formatCurrency, formatDate, formatPercent } from "../lib/format";
import { useSessionStore } from "../features/auth/store";
import { usePortfolio, usePortfolioPerformance } from "../features/portfolio/hooks";
import { useTransactions } from "../features/transactions/hooks";

export function AssetDetailsPage() {
  const { assetId } = useParams();
  const currency = useSessionStore((state) => state.preferredCurrency);
  const portfolioQuery = usePortfolio(currency);
  const performanceQuery = usePortfolioPerformance(currency);
  const transactionsQuery = useTransactions();
  const iconQuery = useAssetIcon(assetId);
  const [iconURL, setIconURL] = useState<string | null>(null);

  const position = useMemo(
    () => portfolioQuery.data?.positions.find((item) => item.asset.id === assetId),
    [assetId, portfolioQuery.data]
  );

  const assetTransactions = useMemo(
    () => transactionsQuery.data?.data.filter((item) => item.assetId === assetId) ?? [],
    [assetId, transactionsQuery.data]
  );

  useEffect(() => {
    if (!iconQuery.data) {
      setIconURL(null);
      return;
    }

    let active = true;
    const reader = new FileReader();
    reader.onloadend = () => {
      if (active) {
        setIconURL(typeof reader.result === "string" ? reader.result : null);
      }
    };
    reader.readAsDataURL(iconQuery.data);

    return () => {
      active = false;
    };
  }, [iconQuery.data]);

  if (portfolioQuery.isLoading || performanceQuery.isLoading || transactionsQuery.isLoading) {
    return <Card>Loading asset details...</Card>;
  }

  if (!position) {
    return <EmptyState title="Asset not found" description="Pick an asset from the portfolio table." />;
  }

  return (
    <div className="space-y-6">
      <Card>
        <Link to="/portfolio">
          <SecondaryButton>Back to portfolio</SecondaryButton>
        </Link>
        <div className="mt-4 flex items-center gap-4">
          {iconURL ? (
            <img
              alt={`${position.asset.symbol} icon`}
              className="h-14 w-14 rounded-2xl bg-white/90 p-2 object-contain"
              src={iconURL}
            />
          ) : (
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl border border-slate-700 bg-slate-900 text-lg font-semibold text-slate-200">
              {position.asset.symbol.slice(0, 2)}
            </div>
          )}
          <div>
            <h2 className="text-2xl font-semibold">{position.asset.symbol}</h2>
            <p className="mt-1 text-slate-400">
              {position.asset.name} • {position.asset.exchange} • {position.asset.currency}
            </p>
          </div>
        </div>
        <div className="mt-5 grid gap-4 md:grid-cols-4">
          <div>
            <p className="text-sm text-slate-400">Open quantity</p>
            <p className="mt-2 text-xl font-semibold">{position.quantity}</p>
          </div>
          <div>
            <p className="text-sm text-slate-400">Average cost</p>
            <p className="mt-2 text-xl font-semibold">{formatCurrency(position.averageCost, currency)}</p>
          </div>
          <div>
            <p className="text-sm text-slate-400">Current price</p>
            <p className="mt-2 text-xl font-semibold">{formatCurrency(position.currentPrice, position.currentCurrency || currency)}</p>
          </div>
          <div>
            <p className="text-sm text-slate-400">Unrealized P/L</p>
            <p className="mt-2 text-xl font-semibold">
              {formatCurrency(position.unrealizedPL, currency)} ({formatPercent(position.unrealizedPLPct)})
            </p>
          </div>
        </div>
      </Card>

      <Card>
        <h3 className="text-xl font-semibold">Portfolio evolution context</h3>
        <p className="mt-1 text-sm text-slate-400">The backend currently exposes portfolio-level performance, reused here for context.</p>
        <div className="mt-4">
          <PerformanceChart points={performanceQuery.data?.data ?? []} currency={currency} />
        </div>
      </Card>

      <Card>
        <h3 className="text-xl font-semibold">Trade history</h3>
        {assetTransactions.length === 0 ? (
          <p className="mt-3 text-sm text-slate-400">No trades for this asset yet.</p>
        ) : (
          <div className="mt-4 overflow-x-auto">
            <table className="min-w-full text-left text-sm">
              <thead className="text-slate-400">
                <tr>
                  <th className="pb-3">Type</th>
                  <th className="pb-3">Quantity</th>
                  <th className="pb-3">Price</th>
                  <th className="pb-3">Fees</th>
                  <th className="pb-3">Date</th>
                </tr>
              </thead>
              <tbody>
                {assetTransactions.map((transaction) => (
                  <tr key={transaction.id} className="border-t border-slate-800/70">
                    <td className="py-4">{transaction.type}</td>
                    <td className="py-4">{transaction.quantity}</td>
                    <td className="py-4">{formatCurrency(transaction.price, transaction.currency)}</td>
                    <td className="py-4">{formatCurrency(transaction.fees, transaction.currency)}</td>
                    <td className="py-4">{formatDate(transaction.transactionDate)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}
