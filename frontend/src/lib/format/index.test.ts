import { describe, expect, it } from "vitest";

import { formatCurrency, formatDate, formatPercent } from "./index";

describe("format helpers", () => {
  it("formats currency with the selected code", () => {
    expect(formatCurrency("1234.5", "USD")).toContain("$1,234.50");
  });

  it("formats percents with sign", () => {
    expect(formatPercent("5.234")).toBe("+5.23%");
    expect(formatPercent("-1.2")).toBe("-1.20%");
  });

  it("formats dates for the UI", () => {
    expect(formatDate("2026-06-13T12:00:00Z")).toContain("2026");
  });
});
