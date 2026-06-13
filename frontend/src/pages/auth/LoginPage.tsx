import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";

import { useLogin } from "../../features/auth/hooks";
import { useSessionStore } from "../../features/auth/store";
import { Button, Card, Field, Input, Select } from "../../components/ui/primitives";

const schema = z.object({
  email: z.string().email("Enter a valid email"),
  password: z.string().min(8, "Password must have at least 8 characters")
});

type FormValues = z.infer<typeof schema>;

export function LoginPage() {
  const navigate = useNavigate();
  const login = useLogin();
  const theme = useSessionStore((state) => state.theme);
  const setTheme = useSessionStore((state) => state.setTheme);
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "" }
  });

  return (
    <div className="flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-md">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.3em] text-sky-300">Finance</p>
          </div>
          <div className="w-32">
            <Field label="Theme">
              <Select aria-label="Theme" value={theme} onChange={(event) => setTheme(event.target.value as "dark" | "light" | "neon")}>
                <option value="dark">Dark</option>
                <option value="light">Light</option>
                <option value="neon">Neon</option>
              </Select>
            </Field>
          </div>
        </div>
        <h1 className="mt-3 text-3xl font-semibold">Welcome</h1>
        <p className="mt-2 text-sm text-slate-400">Sign in to monitor your portfolio across markets.</p>

        <form
          className="mt-6 space-y-4"
          onSubmit={form.handleSubmit(async (values) => {
            await login.mutateAsync(values);
            navigate("/");
          })}
        >
          <Field label="Email" error={form.formState.errors.email?.message}>
            <Input aria-label="Email" {...form.register("email")} />
          </Field>
          <Field label="Password" error={form.formState.errors.password?.message}>
            <Input aria-label="Password" type="password" {...form.register("password")} />
          </Field>
          {login.error ? <p className="text-sm text-rose-300">{login.error.message}</p> : null}
          <Button className="w-full" type="submit" disabled={login.isPending}>
            {login.isPending ? "Signing in..." : "Sign in"}
          </Button>
        </form>

        <div className="mt-6 flex justify-between text-sm text-slate-400">
          <Link to="/register">Create account</Link>
          <Link to="/reset-password">Reset password</Link>
        </div>
      </Card>
    </div>
  );
}
