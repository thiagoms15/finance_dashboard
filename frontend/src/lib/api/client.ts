import { useSessionStore } from "../../features/auth/store";
import type {
  ApiListResponse,
  Asset,
  CreateAssetRequest,
  Dividend,
  LoginResponse,
  PerformancePoint,
  PortfolioSnapshot,
  PortfolioSummary,
  Transaction
} from "../../types/api";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api";

type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  body?: unknown;
  auth?: boolean;
  credentials?: RequestCredentials;
};

let refreshPromise: Promise<boolean> | null = null;

function listResponse<T>(payload: ApiListResponse<T> | null | undefined): ApiListResponse<T> {
  return {
    data: Array.isArray(payload?.data) ? payload.data : []
  };
}

async function executeRequest(path: string, options: RequestOptions = {}, tokenOverride?: string | null) {
  const token = tokenOverride ?? useSessionStore.getState().token;
  const headers: Record<string, string> = {
    "Content-Type": "application/json"
  };

  if (options.auth !== false && token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? "GET",
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined,
    credentials: options.credentials
  });

  return response;
}

async function refreshAccessToken() {
  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = (async () => {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "include"
    });

    if (!response.ok) {
      useSessionStore.getState().clearAuth();
      return false;
    }

    const payload = (await response.json()) as LoginResponse;
    useSessionStore.getState().setAuth(payload.accessToken, payload.user, payload.expiresIn);
    return true;
  })().finally(() => {
    refreshPromise = null;
  });

  return refreshPromise;
}

async function request<T>(path: string, options: RequestOptions = {}, allowRefresh = true): Promise<T> {
  let response = await executeRequest(path, options);

  if (response.status === 401 && options.auth !== false && allowRefresh) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      response = await executeRequest(path, options, useSessionStore.getState().token);
    }
  }

  if (response.status === 401) {
    useSessionStore.getState().clearAuth();
  }

  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as
      | { error?: { message?: string } }
      | null;
    throw new Error(payload?.error?.message ?? `Request failed with ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export const api = {
  login: (body: { email: string; password: string }) =>
    request<LoginResponse>("/auth/login", { method: "POST", body, auth: false, credentials: "include" }),
  register: (body: { name: string; email: string; password: string }) =>
    request<LoginResponse>("/auth/register", { method: "POST", body, auth: false, credentials: "include" }),
  logout: () =>
    request<{ ok: true }>("/auth/logout", { method: "POST", auth: false, credentials: "include" }, false),
  requestPasswordReset: (body: { email: string }) =>
    request<{ ok: true; resetToken?: string }>("/auth/password-reset/request", {
      method: "POST",
      body,
      auth: false
    }),
  confirmPasswordReset: (body: { token: string; newPassword: string }) =>
    request<{ ok: true }>("/auth/password-reset/confirm", {
      method: "POST",
      body,
      auth: false
    }),
  assets: async (search = "") =>
    listResponse(await request<ApiListResponse<Asset>>(`/assets${search ? `?search=${encodeURIComponent(search)}` : ""}`)),
  createAsset: (body: CreateAssetRequest) => request<Asset>("/assets", { method: "POST", body }),
  asset: (id: string) => request<Asset>(`/assets/${id}`),
  assetIcon: async (id: string) => {
    let response = await executeRequest(`/assets/${id}/icon`);

    if (response.status === 401) {
      const refreshed = await refreshAccessToken();
      if (refreshed) {
        response = await executeRequest(`/assets/${id}/icon`, {}, useSessionStore.getState().token);
      }
    }

    if (response.status === 401) {
      useSessionStore.getState().clearAuth();
    }

    if (!response.ok) {
      const payload = (await response.json().catch(() => null)) as
        | { error?: { message?: string } }
        | null;
      throw new Error(payload?.error?.message ?? `Request failed with ${response.status}`);
    }

    return response.blob();
  },
  transactions: async () => listResponse(await request<ApiListResponse<Transaction>>("/transactions")),
  createTransaction: (body: Record<string, string>) =>
    request<Transaction>("/transactions", { method: "POST", body }),
  updateTransaction: (id: string, body: Record<string, string>) =>
    request<Transaction>(`/transactions/${id}`, { method: "PUT", body }),
  deleteTransaction: (id: string) =>
    request<void>(`/transactions/${id}`, { method: "DELETE" }),
  dividends: async () => listResponse(await request<ApiListResponse<Dividend>>("/dividends")),
  createDividend: (body: Record<string, string>) =>
    request<Dividend>("/dividends", { method: "POST", body }),
  updateDividend: (id: string, body: Record<string, string>) =>
    request<Dividend>(`/dividends/${id}`, { method: "PUT", body }),
  deleteDividend: (id: string) => request<void>(`/dividends/${id}`, { method: "DELETE" }),
  portfolio: (currency: string) => request<PortfolioSnapshot>(`/portfolio?currency=${currency}`),
  portfolioSummary: (currency: string) =>
    request<PortfolioSummary>(`/portfolio/summary?currency=${currency}`),
  portfolioPerformance: async (currency: string) =>
    listResponse(await request<ApiListResponse<PerformancePoint>>(`/portfolio/performance?currency=${currency}`))
};
