import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";

import { useRegister } from "../../features/auth/hooks";
import { Button, Card, Field, Input } from "../../components/ui/primitives";

const schema = z
  .object({
    name: z.string().min(2, "Enter your name"),
    email: z.string().email("Enter a valid email"),
    password: z.string().min(8, "Password must have at least 8 characters"),
    confirmPassword: z.string().min(8, "Confirm your password")
  })
  .refine((values) => values.password === values.confirmPassword, {
    message: "Passwords must match",
    path: ["confirmPassword"]
  });

type FormValues = z.infer<typeof schema>;

export function RegisterPage() {
  const navigate = useNavigate();
  const register = useRegister();
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", email: "", password: "", confirmPassword: "" }
  });

  return (
    <div className="flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <p className="text-xs uppercase tracking-[0.3em] text-sky-300">Finance</p>
        <h1 className="mt-3 text-3xl font-semibold">Create account</h1>
        <p className="mt-2 text-sm text-slate-400">Start tracking B3, NASDAQ, and crypto positions in one place.</p>

        <form
          className="mt-6 space-y-4"
          onSubmit={form.handleSubmit(async (values) => {
            await register.mutateAsync({
              name: values.name,
              email: values.email,
              password: values.password
            });
            navigate("/");
          })}
        >
          <Field label="Name" error={form.formState.errors.name?.message}>
            <Input aria-label="Name" {...form.register("name")} />
          </Field>
          <Field label="Email" error={form.formState.errors.email?.message}>
            <Input aria-label="Email" {...form.register("email")} />
          </Field>
          <Field label="Password" error={form.formState.errors.password?.message}>
            <Input aria-label="Password" type="password" {...form.register("password")} />
          </Field>
          <Field label="Confirm password" error={form.formState.errors.confirmPassword?.message}>
            <Input aria-label="Confirm password" type="password" {...form.register("confirmPassword")} />
          </Field>
          {register.error ? <p className="text-sm text-rose-300">{register.error.message}</p> : null}
          <Button className="w-full" type="submit" disabled={register.isPending}>
            {register.isPending ? "Creating account..." : "Create account"}
          </Button>
        </form>

        <div className="mt-6 text-sm text-slate-400">
          Already registered? <Link to="/login">Sign in</Link>
        </div>
      </Card>
    </div>
  );
}
