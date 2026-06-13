import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import { z } from "zod";

import type { Asset, Transaction } from "../../types/api";
import { Button, Field, Input, SecondaryButton, Select } from "../ui/primitives";

const transactionSchema = z.object({
  assetId: z.string().uuid("Choose an asset"),
  type: z.enum(["BUY", "SELL"]),
  quantity: z.string().min(1, "Quantity is required"),
  price: z.string().min(1, "Price is required"),
  fees: z.string().min(1, "Fees are required"),
  currency: z.string().length(3, "Currency must be 3 letters"),
  transactionDate: z.string().min(1, "Date is required"),
  notes: z.string()
});

export type TransactionFormValues = z.infer<typeof transactionSchema>;

export function TransactionForm({
  assets,
  initial,
  onSubmit,
  submitLabel = "Save transaction",
  defaultType = "BUY",
  preferredAssetId,
  onCancel
}: {
  assets: Asset[];
  initial?: Partial<Transaction>;
  onSubmit: (values: TransactionFormValues) => Promise<void> | void;
  submitLabel?: string;
  defaultType?: "BUY" | "SELL";
  preferredAssetId?: string;
  onCancel?: () => void;
}) {
  const defaultValues = {
    assetId: initial?.assetId ?? "",
    type: initial?.type === "SELL" ? "SELL" : defaultType,
    quantity: initial?.quantity ?? "",
    price: initial?.price ?? "",
    fees: initial?.fees ?? "0",
    currency: initial?.currency ?? "USD",
    transactionDate: initial?.transactionDate?.slice(0, 16) ?? new Date().toISOString().slice(0, 16),
    notes: initial?.notes ?? ""
  } satisfies TransactionFormValues;

  const form = useForm<TransactionFormValues>({
    resolver: zodResolver(transactionSchema),
    defaultValues
  });
  const selectedAssetID = useWatch({ control: form.control, name: "assetId" });

  useEffect(() => {
    form.reset(defaultValues);
  }, [defaultType, form, initial?.assetId, initial?.currency, initial?.fees, initial?.notes, initial?.price, initial?.quantity, initial?.transactionDate, initial?.type]);

  useEffect(() => {
    if (preferredAssetId) {
      form.setValue("assetId", preferredAssetId, { shouldValidate: true });
    }
  }, [form, preferredAssetId]);

  useEffect(() => {
    const selectedAsset = assets.find((asset) => asset.id === selectedAssetID);
    if (selectedAsset) {
      form.setValue("currency", selectedAsset.currency, { shouldValidate: true });
    }
  }, [assets, form, selectedAssetID]);

  return (
    <form
      className="grid gap-3 md:grid-cols-2"
      onSubmit={form.handleSubmit(async (values) => {
        await onSubmit(values);
        if (!initial) {
          form.reset(defaultValues);
        }
      })}
    >
      <Field label="Asset" error={form.formState.errors.assetId?.message}>
        <Select aria-label="Asset" {...form.register("assetId")}>
          <option value="">Select an asset</option>
          {assets.map((asset) => (
            <option key={asset.id} value={asset.id}>
              {asset.symbol} · {asset.exchange}
            </option>
          ))}
        </Select>
      </Field>
      <Field label="Type" error={form.formState.errors.type?.message}>
        <Select aria-label="Type" {...form.register("type")}>
          <option value="BUY">BUY</option>
          <option value="SELL">SELL</option>
        </Select>
      </Field>
      <Field label="Quantity" error={form.formState.errors.quantity?.message}>
        <Input aria-label="Quantity" {...form.register("quantity")} />
      </Field>
      <Field label="Price" error={form.formState.errors.price?.message}>
        <Input aria-label="Price" {...form.register("price")} />
      </Field>
      <Field label="Fees" error={form.formState.errors.fees?.message}>
        <Input aria-label="Fees" {...form.register("fees")} />
      </Field>
      <Field label="Currency" error={form.formState.errors.currency?.message}>
        <Select aria-label="Currency" {...form.register("currency")}>
          <option value="USD">USD</option>
          <option value="BRL">BRL</option>
        </Select>
      </Field>
      <Field label="Transaction date" error={form.formState.errors.transactionDate?.message}>
        <Input aria-label="Transaction date" type="datetime-local" {...form.register("transactionDate")} />
      </Field>
      <Field label="Notes" error={form.formState.errors.notes?.message}>
        <Input aria-label="Notes" {...form.register("notes")} />
      </Field>
      <div className="md:col-span-2">
        <div className="flex flex-wrap gap-3">
          <Button type="submit">{submitLabel}</Button>
          {onCancel ? (
            <SecondaryButton type="button" onClick={onCancel}>
              Cancel
            </SecondaryButton>
          ) : null}
        </div>
      </div>
    </form>
  );
}
