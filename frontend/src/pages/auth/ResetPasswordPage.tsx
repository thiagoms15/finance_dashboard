import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Link } from "react-router-dom";
import { z } from "zod";

import { useConfirmPasswordReset, useRequestPasswordReset } from "../../features/auth/hooks";
import { Button, Card, Field, Input } from "../../components/ui/primitives";

const requestSchema = z.object({
  email: z.string().email("Enter a valid email")
});

const confirmSchema = z.object({
  token: z.string().min(1, "Reset token is required"),
  newPassword: z.string().min(8, "Password must have at least 8 characters")
});

type RequestValues = z.infer<typeof requestSchema>;
type ConfirmValues = z.infer<typeof confirmSchema>;

export function ResetPasswordPage() {
  const requestReset = useRequestPasswordReset();
  const confirmReset = useConfirmPasswordReset();
  const requestForm = useForm<RequestValues>({
    resolver: zodResolver(requestSchema),
    defaultValues: { email: "" }
  });
  const confirmForm = useForm<ConfirmValues>({
    resolver: zodResolver(confirmSchema),
    defaultValues: { token: "", newPassword: "" }
  });

  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-8">
      <div className="grid w-full max-w-4xl gap-6 lg:grid-cols-2">
        <Card>
          <h1 className="text-2xl font-semibold">Request reset</h1>
          <p className="mt-2 text-sm text-slate-400">In development, the backend returns the reset token so you can complete the flow locally.</p>
          <form
            className="mt-6 space-y-4"
            onSubmit={requestForm.handleSubmit(async (values) => {
              await requestReset.mutateAsync(values);
            })}
          >
            <Field label="Email" error={requestForm.formState.errors.email?.message}>
              <Input aria-label="Reset email" {...requestForm.register("email")} />
            </Field>
            <Button className="w-full" type="submit" disabled={requestReset.isPending}>
              Request reset
            </Button>
            {"data" in requestReset && requestReset.data?.resetToken ? (
              <p className="text-sm text-emerald-300">Reset token: {requestReset.data.resetToken}</p>
            ) : null}
          </form>
        </Card>

        <Card>
          <h2 className="text-2xl font-semibold">Confirm reset</h2>
          <p className="mt-2 text-sm text-slate-400">Paste the token from the request step and choose a new password.</p>
          <form
            className="mt-6 space-y-4"
            onSubmit={confirmForm.handleSubmit(async (values) => {
              await confirmReset.mutateAsync(values);
            })}
          >
            <Field label="Reset token" error={confirmForm.formState.errors.token?.message}>
              <Input aria-label="Reset token" {...confirmForm.register("token")} />
            </Field>
            <Field label="New password" error={confirmForm.formState.errors.newPassword?.message}>
              <Input aria-label="New password" type="password" {...confirmForm.register("newPassword")} />
            </Field>
            <Button className="w-full" type="submit" disabled={confirmReset.isPending}>
              Confirm reset
            </Button>
            {confirmReset.isSuccess ? <p className="text-sm text-emerald-300">Password reset complete.</p> : null}
          </form>
          <div className="mt-6 text-sm text-slate-400">
            <Link to="/login">Back to sign in</Link>
          </div>
        </Card>
      </div>
    </div>
  );
}
