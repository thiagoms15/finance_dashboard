import { create } from "zustand";

import type { User } from "../../types/api";

export type Theme = "dark" | "light" | "neon";
type Currency = "USD" | "BRL";

type SessionState = {
  token: string | null;
  user: User | null;
  sessionExpiresAt: number | null;
  theme: Theme;
  preferredCurrency: Currency;
  setAuth: (token: string, user: User, expiresInSeconds: number) => void;
  clearAuth: () => void;
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
  setPreferredCurrency: (currency: Currency) => void;
};

const SESSION_STORAGE_KEY = "session";

const themes: Theme[] = ["dark", "light", "neon"];

const isTheme = (value: string | null): value is Theme => value !== null && themes.includes(value as Theme);

const themeFromStorage = (): Theme => {
  if (typeof window === "undefined") {
    return "dark";
  }
  const stored = window.localStorage.getItem("theme");
  return isTheme(stored) ? stored : "dark";
};

const currencyFromStorage = (): Currency => {
  if (typeof window === "undefined") {
    return "USD";
  }
  return (window.localStorage.getItem("preferredCurrency") as Currency) || "USD";
};

const sessionFromStorage = (): { token: string | null; user: User | null; sessionExpiresAt: number | null } => {
  if (typeof window === "undefined") {
    return { token: null, user: null, sessionExpiresAt: null };
  }

  const raw = window.localStorage.getItem(SESSION_STORAGE_KEY);
  if (!raw) {
    return { token: null, user: null, sessionExpiresAt: null };
  }

  try {
    const parsed = JSON.parse(raw) as { token?: string; user?: User; sessionExpiresAt?: number };
    if (!parsed.token || !parsed.user || !parsed.sessionExpiresAt || parsed.sessionExpiresAt <= Date.now()) {
      window.localStorage.removeItem(SESSION_STORAGE_KEY);
      return { token: null, user: null, sessionExpiresAt: null };
    }

    return {
      token: parsed.token,
      user: parsed.user,
      sessionExpiresAt: parsed.sessionExpiresAt
    };
  } catch {
    window.localStorage.removeItem(SESSION_STORAGE_KEY);
    return { token: null, user: null, sessionExpiresAt: null };
  }
};

const persistSession = (token: string, user: User, sessionExpiresAt: number) => {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(
    SESSION_STORAGE_KEY,
    JSON.stringify({
      token,
      user,
      sessionExpiresAt
    })
  );
};

const clearStoredSession = () => {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(SESSION_STORAGE_KEY);
};

const initialSession = sessionFromStorage();

export const useSessionStore = create<SessionState>((set) => ({
  token: initialSession.token,
  user: initialSession.user,
  sessionExpiresAt: initialSession.sessionExpiresAt,
  theme: themeFromStorage(),
  preferredCurrency: currencyFromStorage(),
  setAuth: (token, user, expiresInSeconds) => {
    const sessionExpiresAt = Date.now() + expiresInSeconds * 1000;
    persistSession(token, user, sessionExpiresAt);
    set({ token, user, sessionExpiresAt });
  },
  clearAuth: () => {
    clearStoredSession();
    set({ token: null, user: null, sessionExpiresAt: null });
  },
  setTheme: (theme) => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem("theme", theme);
      document.documentElement.dataset.theme = theme;
    }
    set({ theme });
  },
  toggleTheme: () =>
    set((state) => {
      const currentIndex = themes.indexOf(state.theme);
      const nextTheme = themes[(currentIndex + 1) % themes.length];
      if (typeof window !== "undefined") {
        window.localStorage.setItem("theme", nextTheme);
        document.documentElement.dataset.theme = nextTheme;
      }
      return { theme: nextTheme };
    }),
  setPreferredCurrency: (preferredCurrency) => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem("preferredCurrency", preferredCurrency);
    }
    set({ preferredCurrency });
  }
}));
