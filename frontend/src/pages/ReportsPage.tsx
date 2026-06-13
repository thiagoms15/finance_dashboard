import { Card } from "../components/ui/primitives";
import { formatCurrency, formatDate } from "../lib/format";
import { useSessionStore } from "../features/auth/store";
import { useDividends } from "../features/dividends/hooks";
import { usePortfolioPerformance, usePortfolioSummary } from "../features/portfolio/hooks";

export function ReportsPage() {
  const currency = useSessionStore((state) => state.preferredCurrency);
  const summaryQuery = usePortfolioSummary(currency);
  const performanceQuery = usePortfolioPerformance(currency);
  const dividendsQuery = useDividends();

  if (summaryQuery.isLoading || performanceQuery.isLoading || dividendsQuery.isLoading) {
    return <Card>Loading reports...</Card>;
  }

  if (!summaryQuery.data || !performanceQuery.data || !dividendsQuery.data) {
    return <Card>Unable to load reports.</Card>;
  }

  const latestPoint = performanceQuery.data.data.at(-1);

  return (
    <div className="grid gap-6 xl:grid-cols-2">
      <Card>
        <h2 className="text-xl font-semibold">Performance snapshot</h2>
        <div className="mt-4 space-y-3 text-sm text-slate-300">
          <p>Total invested: {formatCurrency(summaryQuery.data.totalInvested, currency)}</p>
          <p>Current value: {formatCurrency(summaryQuery.data.currentValue, currency)}</p>
          <p>Realized profit: {formatCurrency(summaryQuery.data.realizedProfit, currency)}</p>
          <p>Dividends received: {formatCurrency(summaryQuery.data.dividendsReceived, currency)}</p>
        </div>
      </Card>

      <Card>
        <h2 className="text-xl font-semibold">Latest performance point</h2>
        {latestPoint ? (
          <div className="mt-4 space-y-3 text-sm text-slate-300">
            <p>Date: {formatDate(latestPoint.date)}</p>
            <p>Estimated portfolio value: {formatCurrency(latestPoint.value, currency)}</p>
          </div>
        ) : (
          <p className="mt-4 text-sm text-slate-400">No performance points yet.</p>
        )}
      </Card>

      <Card className="xl:col-span-2">
        <h2 className="text-xl font-semibold">Dividend history</h2>
        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full text-left text-sm">
            <thead className="text-slate-400">
              <tr>
                <th className="pb-3">Amount</th>
                <th className="pb-3">Currency</th>
                <th className="pb-3">Payment date</th>
              </tr>
            </thead>
            <tbody>
              {dividendsQuery.data.data.map((item) => (
                <tr key={item.id} className="border-t border-slate-800/70">
                  <td className="py-4">{formatCurrency(item.amount, item.currency)}</td>
                  <td className="py-4">{item.currency}</td>
                  <td className="py-4">{formatDate(item.paymentDate)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
