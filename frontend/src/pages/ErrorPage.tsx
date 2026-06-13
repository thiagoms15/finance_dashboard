import { isRouteErrorResponse, Link, useRouteError } from "react-router-dom";

import { Button, Card, SecondaryButton } from "../components/ui/primitives";

export function ErrorPage() {
  const error = useRouteError();

  const message = isRouteErrorResponse(error)
    ? `${error.status} ${error.statusText}`
    : error instanceof Error
      ? error.message
      : "Something unexpected happened.";

  return (
    <div className="flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-xl text-center">
        <p className="eyebrow">App Error</p>
        <h1 className="mt-3 text-3xl font-semibold">Something went wrong</h1>
        <p className="mt-3 text-muted">{message}</p>
        <div className="mt-6 flex justify-center gap-3">
          <Link to="/">
            <Button>Back to dashboard</Button>
          </Link>
          <SecondaryButton onClick={() => window.location.reload()}>
            Reload app
          </SecondaryButton>
        </div>
      </Card>
    </div>
  );
}
