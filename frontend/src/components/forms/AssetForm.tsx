import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button, Field, Input, Select } from "../ui/primitives";

const assetSchema = z.object({
  symbol: z.string().min(1, "Symbol is required"),
  name: z.string(),
  exchange: z.string().min(1, "Exchange is required"),
  currency: z.string().length(3, "Currency must be 3 letters"),
  sector: z.string()
});

export type AssetFormValues = z.infer<typeof assetSchema>;

export function AssetForm({
  onSubmit
}: {
  onSubmit: (values: AssetFormValues) => Promise<void> | void;
}) {
  const form = useForm<AssetFormValues>({
    resolver: zodResolver(assetSchema),
    defaultValues: {
      symbol: "",
      name: "",
      exchange: "B3",
      currency: "BRL",
      sector: ""
    }
  });

  return (
    <form
      className="grid gap-3 md:grid-cols-2"
      onSubmit={form.handleSubmit(async (values) => {
        await onSubmit(values);
        form.reset({
          symbol: "",
          name: "",
          exchange: values.exchange,
          currency: values.currency,
          sector: ""
        });
      })}
    >
      <Field label="Symbol" error={form.formState.errors.symbol?.message}>
        <Input aria-label="Symbol" placeholder="ITUB4" {...form.register("symbol")} />
      </Field>
      <Field label="Name (optional)" error={form.formState.errors.name?.message}>
        <Input aria-label="Asset name" placeholder="Itaú Unibanco" {...form.register("name")} />
      </Field>
      <Field label="Exchange" error={form.formState.errors.exchange?.message}>
        <Select aria-label="Exchange" {...form.register("exchange")}>
          <option value="B3">B3</option>
          <option value="NASDAQ">NASDAQ</option>
          <option value="CRYPTO">CRYPTO</option>
          <option value="NYSE">NYSE</option>
        </Select>
      </Field>
      <Field label="Currency" error={form.formState.errors.currency?.message}>
        <Select aria-label="Asset currency" {...form.register("currency")}>
          <option value="BRL">BRL</option>
          <option value="USD">USD</option>
        </Select>
      </Field>
      <Field label="Sector (optional)" error={form.formState.errors.sector?.message}>
        <Input aria-label="Sector" placeholder="Financials" {...form.register("sector")} />
      </Field>
      <div className="md:col-span-2">
        <Button type="submit">Create symbol</Button>
      </div>
    </form>
  );
}
