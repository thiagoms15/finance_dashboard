import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { TransactionForm } from "./TransactionForm";

const assets = [
  {
    id: "550e8400-e29b-41d4-a716-446655440000",
    symbol: "AAPL",
    name: "Apple",
    exchange: "NASDAQ",
    currency: "USD",
    sector: "Technology",
    createdAt: "",
    updatedAt: ""
  }
];

describe("TransactionForm", () => {
  it("submits normalized form values", async () => {
    const onSubmit = vi.fn();

    render(<TransactionForm assets={assets} onSubmit={onSubmit} />);

    fireEvent.change(screen.getByLabelText("Asset"), {
      target: { value: assets[0].id }
    });
    fireEvent.change(screen.getByLabelText("Quantity"), {
      target: { value: "10" }
    });
    fireEvent.change(screen.getByLabelText("Price"), {
      target: { value: "100.50" }
    });
    fireEvent.change(screen.getByLabelText("Fees"), {
      target: { value: "1.25" }
    });
    fireEvent.change(screen.getByLabelText("Transaction date"), {
      target: { value: "2026-06-13T11:00" }
    });

    fireEvent.click(screen.getByRole("button", { name: "Save transaction" }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          assetId: assets[0].id,
          quantity: "10",
          price: "100.50",
          fees: "1.25"
        })
      );
    });
  });
});
