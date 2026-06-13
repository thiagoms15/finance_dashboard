import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import type { PerformancePoint } from "../../types/api";
import { formatCurrency, formatDate } from "../../lib/format";
import { EmptyState } from "../ui/primitives";

export function PerformanceChart({
  points,
  currency
}: {
  points: PerformancePoint[];
  currency: string;
}) {
  if (points.length === 0) {
    return <EmptyState title="No performance data" description="Create trades to visualize portfolio evolution." />;
  }

  const groupedByDay = new Map<string, { date: string; value: number }>();
  for (const point of points) {
    const day = formatDate(point.date);
    // Keep the latest point of the day so intraday updates are represented.
    groupedByDay.set(day, {
      date: day,
      value: Number(point.value)
    });
  }
  const data = Array.from(groupedByDay.values());

  return (
    <div className="h-80">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data}>
          <defs>
            <linearGradient id="perfFill" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stopColor="#38bdf8" stopOpacity={0.5} />
              <stop offset="100%" stopColor="#38bdf8" stopOpacity={0.06} />
            </linearGradient>
          </defs>
          <CartesianGrid stroke="rgba(148,163,184,0.15)" vertical={false} />
          <XAxis dataKey="date" stroke="#94a3b8" />
          <YAxis stroke="#94a3b8" tickFormatter={(value) => formatCurrency(value, currency)} />
          <Tooltip formatter={(value) => formatCurrency(Number(value ?? 0), currency)} />
          <Area type="monotone" dataKey="value" stroke="#38bdf8" fill="url(#perfFill)" strokeWidth={2} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
