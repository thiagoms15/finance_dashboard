import { useState } from "react";

import { Card, DangerButton, EmptyState, SecondaryButton } from "../components/ui/primitives";
import { AssetForm } from "../components/forms/AssetForm";
import { DividendForm } from "../components/forms/DividendForm";
import { TransactionForm } from "../components/forms/TransactionForm";
import { formatCurrency, formatDate } from "../lib/format";
import { useAssetMutations, useAssets } from "../features/assets/hooks";
import { useDividendMutations, useDividends } from "../features/dividends/hooks";
import { useTransactionMutations, useTransactions } from "../features/transactions/hooks";
import type { Transaction } from "../types/api";

const actionConfig = {
  BUY: {
    title: "Buy stock or crypto",
    description: "Register a new purchase. This is the clearest way to add a position to your portfolio.",
    submitLabel: "Add order"
  },
  SELL: {
    title: "Sell position",
    description: "Register a partial or full exit. Realized profit is recalculated on the backend.",
    submitLabel: "Add sell order"
  },
  DIVIDEND: {
    title: "Add dividend",
    description: "Track cash income separately from price appreciation.",
    submitLabel: "Add dividend"
  }
} as const;

export function TransactionsPage() {
  const [mode, setMode] = useState<"BUY" | "SELL" | "DIVIDEND">("BUY");
  const [preferredAssetId, setPreferredAssetId] = useState<string>("");
  const [editingTransaction, setEditingTransaction] = useState<Transaction | null>(null);
  const [pendingDelete, setPendingDelete] = useState<{
    id: string;
    kind: "transaction" | "dividend";
    label: string;
  } | null>(null);
  const assetsQuery = useAssets();
  const assetMutations = useAssetMutations();
  const transactionsQuery = useTransactions();
  const dividendsQuery = useDividends();
  const transactionMutations = useTransactionMutations();
  const dividendMutations = useDividendMutations();

  if (assetsQuery.isLoading || transactionsQuery.isLoading || dividendsQuery.isLoading) {
    return <Card>Loading transactions...</Card>;
  }

  if (!assetsQuery.data || !transactionsQuery.data || !dividendsQuery.data) {
    return <Card>Unable to load transaction data.</Card>;
  }

  const assets = Array.isArray(assetsQuery.data.data) ? assetsQuery.data.data : [];
  const transactions = Array.isArray(transactionsQuery.data.data) ? transactionsQuery.data.data : [];
  const dividends = Array.isArray(dividendsQuery.data.data) ? dividendsQuery.data.data : [];
  const isDeleting =
    pendingDelete?.kind === "transaction" ? transactionMutations.remove.isPending : dividendMutations.remove.isPending;

  const confirmDelete = async () => {
    if (!pendingDelete) {
      return;
    }

    if (pendingDelete.kind === "transaction") {
      await transactionMutations.remove.mutateAsync(pendingDelete.id);
      if (editingTransaction?.id === pendingDelete.id) {
        setEditingTransaction(null);
      }
    } else {
      await dividendMutations.remove.mutateAsync(pendingDelete.id);
    }

    setPendingDelete(null);
  };

  const openCreateMode = (nextMode: "BUY" | "SELL" | "DIVIDEND") => {
    setEditingTransaction(null);
    setPreferredAssetId("");
    setMode(nextMode);
  };

  const openEditMode = (transaction: Transaction) => {
    setPendingDelete(null);
    setPreferredAssetId("");
    setEditingTransaction(transaction);
    setMode(transaction.type === "SELL" ? "SELL" : "BUY");
  };

  return (
    <div className="space-y-6">
      {pendingDelete ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 px-4 backdrop-blur-sm">
          <Card className="w-full max-w-md">
            <p className="eyebrow">Confirm delete</p>
            <h3 className="mt-2 text-xl font-semibold">Delete this entry?</h3>
            <p className="mt-3 text-sm text-slate-400">
              You are about to delete {pendingDelete.label}. This action cannot be undone.
            </p>
            <div className="mt-6 flex justify-end gap-3">
              <SecondaryButton onClick={() => setPendingDelete(null)} disabled={isDeleting}>
                Cancel
              </SecondaryButton>
              <DangerButton onClick={() => void confirmDelete()} disabled={isDeleting}>
                {isDeleting ? "Deleting..." : "Delete"}
              </DangerButton>
            </div>
          </Card>
        </div>
      ) : null}

      <Card>
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p className="eyebrow">Trading Desk</p>
            <h2 className="mt-2 text-2xl font-semibold">Add trades with explicit buy and sell actions</h2>
            <p className="mt-2 text-sm text-slate-400">
              Use the buttons below to choose the action you want. Buying an asset starts here.
            </p>
          </div>
          <div className="grid gap-2 sm:grid-cols-3">
            {(["BUY", "SELL", "DIVIDEND"] as const).map((action) => (
              <button
                key={action}
                className={`rounded-2xl border px-4 py-3 text-sm font-medium ${
                  mode === action ? "btn-primary" : "btn-secondary"
                }`}
                onClick={() => openCreateMode(action)}
                type="button"
              >
                {action === "BUY" ? "Buy" : action === "SELL" ? "Sell" : "Dividend"}
              </button>
            ))}
          </div>
        </div>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <Card>
          <h2 className="text-xl font-semibold">
            {editingTransaction ? "Edit transaction" : actionConfig[mode].title}
          </h2>
          <p className="mt-1 text-sm text-slate-400">
            {editingTransaction
              ? "Update the selected trade below and press Update transaction to save your changes."
              : actionConfig[mode].description}
          </p>
          <div className="mt-4">
            {mode === "DIVIDEND" ? (
              <DividendForm
                assets={assets}
                submitLabel={actionConfig[mode].submitLabel}
                onSubmit={async (values) => {
                  await dividendMutations.create.mutateAsync({
                    ...values,
                    paymentDate: new Date(values.paymentDate).toISOString()
                  });
                }}
              />
            ) : (
              <TransactionForm
                key={`${mode}-${editingTransaction?.id ?? "new"}`}
                assets={assets}
                initial={editingTransaction ?? undefined}
                defaultType={mode}
                preferredAssetId={editingTransaction ? undefined : preferredAssetId}
                submitLabel={editingTransaction ? "Update transaction" : actionConfig[mode].submitLabel}
                onCancel={editingTransaction ? () => setEditingTransaction(null) : undefined}
                onSubmit={async (values) => {
                  const body = {
                    ...values,
                    type: values.type,
                    transactionDate: new Date(values.transactionDate).toISOString()
                  };

                  if (editingTransaction) {
                    await transactionMutations.update.mutateAsync({
                      id: editingTransaction.id,
                      body
                    });
                    setEditingTransaction(null);
                    setMode(values.type);
                    return;
                  }

                  await transactionMutations.create.mutateAsync(body);
                }}
              />
            )}
          </div>
        </Card>

        <Card>
          <h2 className="text-xl font-semibold">Create a new symbol</h2>
          <div className="mt-4 space-y-3 text-sm text-slate-400">
            <div className="surface-soft rounded-2xl p-4">
              <p className="font-medium text-slate-100">Need a symbol that is not listed yet?</p>
              <p className="mt-2">
                Create it here, then the buy or sell form will auto-select the new asset for you.
              </p>
            </div>
            <div className="surface-soft rounded-2xl p-4">
              <AssetForm
                onSubmit={async (values) => {
                  const asset = await assetMutations.create.mutateAsync({
                    symbol: values.symbol,
                    name: values.name ?? "",
                    exchange: values.exchange,
                    currency: values.currency,
                    sector: values.sector ?? ""
                  });
                  setPreferredAssetId(asset.id);
                  setEditingTransaction(null);
                  setMode("BUY");
                }}
              />
            </div>
            <div className="surface-soft rounded-2xl p-4">
              <p className="font-medium text-slate-100">Quick navigation</p>
              <div className="mt-3 flex flex-wrap gap-2">
                <SecondaryButton onClick={() => openCreateMode("BUY")}>Open buy form</SecondaryButton>
                <SecondaryButton onClick={() => openCreateMode("SELL")}>Open sell form</SecondaryButton>
                <SecondaryButton onClick={() => openCreateMode("DIVIDEND")}>Open dividend form</SecondaryButton>
              </div>
            </div>
          </div>
        </Card>
      </div>

      <Card>
        <h3 className="text-xl font-semibold">Transactions</h3>
        {transactions.length === 0 ? (
          <EmptyState title="No trades yet" description="Press Buy above to add your first stock or crypto purchase." />
        ) : (
          <div className="mt-4 overflow-x-auto">
            <table className="min-w-full text-left text-sm">
              <thead className="text-slate-400">
                <tr>
                  <th className="pb-3">Type</th>
                  <th className="pb-3">Asset</th>
                  <th className="pb-3">Quantity</th>
                  <th className="pb-3">Price</th>
                  <th className="pb-3">Date</th>
                  <th className="pb-3">Action</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((item) => (
                  <tr key={item.id} className="border-t border-slate-800/70">
                    <td className="py-4">{item.type}</td>
                    <td className="py-4">{assets.find((asset) => asset.id === item.assetId)?.symbol ?? item.assetId}</td>
                    <td className="py-4">{item.quantity}</td>
                    <td className="py-4">{formatCurrency(item.price, item.currency)}</td>
                    <td className="py-4">{formatDate(item.transactionDate)}</td>
                    <td className="py-4">
                      <div className="flex flex-wrap gap-2">
                        <SecondaryButton className="px-3 py-1 text-sm" onClick={() => openEditMode(item)} type="button">
                          Edit
                        </SecondaryButton>
                        <DangerButton
                          className="px-3 py-1 text-sm"
                          onClick={() => {
                            const symbol = assets.find((asset) => asset.id === item.assetId)?.symbol ?? "transaction";
                            setPendingDelete({ id: item.id, kind: "transaction", label: `${item.type} for ${symbol}` });
                          }}
                          type="button"
                        >
                          Delete
                        </DangerButton>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <Card>
        <h3 className="text-xl font-semibold">Dividends</h3>
        {dividends.length === 0 ? (
          <EmptyState title="No dividends yet" description="Use the Dividend mode above when you receive income." />
        ) : (
          <div className="mt-4 overflow-x-auto">
            <table className="min-w-full text-left text-sm">
              <thead className="text-slate-400">
                <tr>
                  <th className="pb-3">Asset</th>
                  <th className="pb-3">Amount</th>
                  <th className="pb-3">Date</th>
                  <th className="pb-3">Action</th>
                </tr>
              </thead>
              <tbody>
                {dividends.map((item) => (
                  <tr key={item.id} className="border-t border-slate-800/70">
                    <td className="py-4">{assets.find((asset) => asset.id === item.assetId)?.symbol ?? item.assetId}</td>
                    <td className="py-4">{formatCurrency(item.amount, item.currency)}</td>
                    <td className="py-4">{formatDate(item.paymentDate)}</td>
                    <td className="py-4">
                      <DangerButton
                        className="px-3 py-1 text-sm"
                        onClick={() => {
                          const symbol = assets.find((asset) => asset.id === item.assetId)?.symbol ?? "dividend";
                          setPendingDelete({ id: item.id, kind: "dividend", label: `dividend for ${symbol}` });
                        }}
                        type="button"
                      >
                        Delete
                      </DangerButton>
                    </td>
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
