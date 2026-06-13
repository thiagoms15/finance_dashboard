import type { PortfolioPosition } from "../../types/api";
import { formatCurrency, formatPercent } from "../../lib/format";
import { Card, EmptyState } from "../ui/primitives";

export function TopMovers({
  positions,
  currency
}: {
  positions: PortfolioPosition[];
  currency: string;
}) {
  if (positions.length === 0) {
    return <EmptyState title="No movers yet" description="Open positions will appear here." />;
  }

  const sorted = [...positions].sort(
    (a, b) => Number(b.unrealizedPLPct) - Number(a.unrealizedPLPct)
  );
  const best = sorted[0];
  const worst = sorted[sorted.length - 1];

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {[{ label: "Best performer", item: best }, { label: "Worst performer", item: worst }].map(({ label, item }) => (
        <Card key={label}>
          <p className="text-sm text-slate-400">{label}</p>
          <h3 className="mt-3 text-xl font-semibold">{item.asset.symbol}</h3>
          <p className="mt-2 text-sm text-slate-300">
            {formatCurrency(item.unrealizedPL, currency)} • {formatPercent(item.unrealizedPLPct)}
          </p>
        </Card>
      ))}
    </div>
  );
}
