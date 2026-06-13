const numberFormatter = (currency: string) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: currency === "BTC" ? 8 : 2
  });

export function formatCurrency(value: string | number, currency = "USD") {
  const numeric = typeof value === "string" ? Number(value) : value;
  return numberFormatter(currency).format(Number.isFinite(numeric) ? numeric : 0);
}

export function formatPercent(value: string | number) {
  const numeric = typeof value === "string" ? Number(value) : value;
  return `${numeric >= 0 ? "+" : ""}${numeric.toFixed(2)}%`;
}

export function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric"
  }).format(new Date(value));
}

export function toNumber(value: string | number) {
  return typeof value === "number" ? value : Number(value);
}
