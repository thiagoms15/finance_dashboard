import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import { z } from "zod";

import type { Asset, Dividend } from "../../types/api";
import { Button, Field, Input, Select } from "../ui/primitives";

const dividendSchema = z.object({
  assetId: z.string().uuid("Choose an asset"),
  amount: z.string().min(1, "Amount is required"),
  currency: z.string().length(3, "Currency must be 3 letters"),
  paymentDate: z.string().min(1, "Date is required")
});

export type DividendFormValues = z.infer<typeof dividendSchema>;

export function DividendForm({
  assets,
  initial,
  onSubmit,
  submitLabel = "Save dividend"
}: {
  assets: Asset[];
  initial?: Partial<Dividend>;
  onSubmit: (values: DividendFormValues) => Promise<void> | void;
  submitLabel?: string;
}) {
  const form = useForm<DividendFormValues>({
    resolver: zodResolver(dividendSchema),
    defaultValues: {
      assetId: initial?.assetId ?? "",
      amount: initial?.amount ?? "",
      currency: initial?.currency ?? "USD",
      paymentDate: initial?.paymentDate?.slice(0, 16) ?? new Date().toISOString().slice(0, 16)
    }
  });
  const selectedAssetID = useWatch({ control: form.control, name: "assetId" });

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
          form.reset({
            assetId: "",
            amount: "",
            currency: "USD",
            paymentDate: new Date().toISOString().slice(0, 16)
          });
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
      <Field label="Amount" error={form.formState.errors.amount?.message}>
        <Input aria-label="Amount" {...form.register("amount")} />
      </Field>
      <Field label="Currency" error={form.formState.errors.currency?.message}>
        <Select aria-label="Currency" {...form.register("currency")}>
          <option value="USD">USD</option>
          <option value="BRL">BRL</option>
        </Select>
      </Field>
      <Field label="Payment date" error={form.formState.errors.paymentDate?.message}>
        <Input aria-label="Payment date" type="datetime-local" {...form.register("paymentDate")} />
      </Field>
      <div className="md:col-span-2">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
