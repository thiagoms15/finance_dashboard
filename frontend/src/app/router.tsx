import type { ReactElement } from "react";
import { createBrowserRouter, Navigate } from "react-router-dom";

import { AppShell } from "../components/layout/AppShell";
import { useSessionStore } from "../features/auth/store";
import { AssetDetailsPage } from "../pages/AssetDetailsPage";
import { DashboardPage } from "../pages/DashboardPage";
import { ErrorPage } from "../pages/ErrorPage";
import { PortfolioPage } from "../pages/PortfolioPage";
import { ReportsPage } from "../pages/ReportsPage";
import { SettingsPage } from "../pages/SettingsPage";
import { TransactionsPage } from "../pages/TransactionsPage";
import { LoginPage } from "../pages/auth/LoginPage";
import { RegisterPage } from "../pages/auth/RegisterPage";
import { ResetPasswordPage } from "../pages/auth/ResetPasswordPage";

function ProtectedLayout() {
  const token = useSessionStore((state) => state.token);
  return token ? <AppShell /> : <Navigate to="/login" replace />;
}

function PublicOnly({ children }: { children: ReactElement }) {
  const token = useSessionStore((state) => state.token);
  return token ? <Navigate to="/" replace /> : children;
}

export const router = createBrowserRouter([
  {
    path: "/login",
    errorElement: <ErrorPage />,
    element: (
      <PublicOnly>
        <LoginPage />
      </PublicOnly>
    )
  },
  {
    path: "/register",
    errorElement: <ErrorPage />,
    element: (
      <PublicOnly>
        <RegisterPage />
      </PublicOnly>
    )
  },
  {
    path: "/reset-password",
    errorElement: <ErrorPage />,
    element: (
      <PublicOnly>
        <ResetPasswordPage />
      </PublicOnly>
    )
  },
  {
    path: "/",
    errorElement: <ErrorPage />,
    element: <ProtectedLayout />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: "portfolio", element: <PortfolioPage /> },
      { path: "portfolio/:assetId", element: <AssetDetailsPage /> },
      { path: "transactions", element: <TransactionsPage /> },
      { path: "reports", element: <ReportsPage /> },
      { path: "settings", element: <SettingsPage /> }
    ]
  }
]);
