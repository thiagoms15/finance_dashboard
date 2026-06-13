import { Card } from "../ui/primitives";
import { formatCurrency } from "../../lib/format";
import type { PortfolioSummary } from "../../types/api";

export function SummaryCards({ summary }: { summary: PortfolioSummary }) {
  const items = [
    { label: "Total Invested", value: summary.totalInvested },
    { label: "Current Value", value: summary.currentValue },
    { label: "Total P/L", value: summary.totalProfitLoss },
    { label: "Daily P/L", value: summary.dailyGainLoss }
  ];

  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      {items.map((item) => (
        <Card key={item.label}>
          <p className="text-sm text-slate-400">{item.label}</p>
          <p className="mt-3 text-2xl font-semibold">
            {formatCurrency(item.value, summary.preferredCurrency)}
          </p>
        </Card>
      ))}
    </div>
  );
}
