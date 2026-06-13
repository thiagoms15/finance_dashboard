import { Pie, PieChart, ResponsiveContainer, Tooltip, Cell } from "recharts";

import type { PortfolioPosition } from "../../types/api";
import { EmptyState } from "../ui/primitives";

const COLORS = ["#38bdf8", "#22c55e", "#f59e0b", "#a855f7", "#ef4444", "#06b6d4"];

export function AllocationChart({ positions }: { positions: PortfolioPosition[] }) {
  if (positions.length === 0) {
    return <EmptyState title="No allocation yet" description="Add transactions to generate allocation data." />;
  }

  const data = positions.map((position) => ({
    name: position.asset.symbol,
    value: Number(position.currentValue)
  }));

  return (
    <div className="h-72">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie data={data} dataKey="value" nameKey="name" outerRadius={100} innerRadius={55}>
            {data.map((entry, index) => (
              <Cell key={entry.name} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
