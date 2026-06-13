import { NavLink, Outlet } from "react-router-dom";

import { Button } from "../ui/primitives";
import { useSessionStore } from "../../features/auth/store";

const navItems = [
  { to: "/", label: "Dashboard" },
  { to: "/portfolio", label: "Portfolio" },
  { to: "/transactions", label: "Transactions" },
  { to: "/reports", label: "Reports" },
  { to: "/settings", label: "Settings" }
];

function displayName(name: string | undefined, email: string | undefined) {
  if (name?.trim()) {
    return name.trim();
  }

  const localPart = email?.split("@")[0]?.trim();
  if (!localPart) {
    return "Investor";
  }

  return localPart
    .replace(/[._-]+/g, " ")
    .split(" ")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function AppShell() {
  const user = useSessionStore((state) => state.user);
  const clearAuth = useSessionStore((state) => state.clearAuth);
  const name = displayName(user?.name, user?.email);

  return (
    <div className="min-h-screen px-4 py-6 md:px-6">
      <div className="mx-auto grid max-w-7xl gap-6 lg:grid-cols-[260px_1fr]">
        <aside className="glass rounded-3xl p-5">
          <div>
            <p className="eyebrow">Finance</p>
            <h1 className="mt-2 text-2xl font-semibold">Portfolio Manager</h1>
            <p className="mt-2 text-sm text-slate-200">{name}</p>
            <p className="mt-1 text-sm text-slate-400">{user?.email}</p>
            <div className="theme-chip mt-4 rounded-2xl px-3 py-3 text-sm">
              Cross-market tracking for B3, NASDAQ, and crypto in one cockpit.
            </div>
          </div>
          <nav className="mt-8 space-y-2">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `block rounded-2xl px-4 py-3 text-sm transition ${
                    isActive ? "bg-sky-500 text-slate-950" : "text-slate-300 hover:bg-slate-900/70"
                  }`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
          <div className="mt-8 flex flex-col gap-3">
            <Button className="bg-rose-400 hover:bg-rose-300" onClick={clearAuth}>
              Sign out
            </Button>
          </div>
        </aside>

        <main className="space-y-6">
          <header className="glass rounded-3xl p-5">
            <div>
              <p className="eyebrow">Command Center</p>
              <h2 className="mt-2 text-3xl font-semibold">Welcome back, {name}.</h2>
              <p className="mt-2 text-sm text-slate-400">
                Your markets are ready. Check performance, review positions, and keep your next move organized.
              </p>
            </div>
          </header>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
