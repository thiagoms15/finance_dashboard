import type { PortfolioSummary } from "../../types/api";
import { formatCurrency } from "../../lib/format";
import { Card } from "../ui/primitives";

export function IncomeCard({ summary }: { summary: PortfolioSummary }) {
  return (
    <Card>
      <p className="text-sm text-slate-400">Dividend income</p>
      <p className="mt-3 text-2xl font-semibold">
        {formatCurrency(summary.dividendsReceived, summary.preferredCurrency)}
      </p>
      <p className="mt-2 text-sm text-slate-400">
        Realized profit: {formatCurrency(summary.realizedProfit, summary.preferredCurrency)}
      </p>
    </Card>
  );
}
