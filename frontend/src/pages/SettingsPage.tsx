import { Card, Field, Select } from "../components/ui/primitives";
import { useSessionStore, type Theme } from "../features/auth/store";

export function SettingsPage() {
  const preferredCurrency = useSessionStore((state) => state.preferredCurrency);
  const setPreferredCurrency = useSessionStore((state) => state.setPreferredCurrency);
  const theme = useSessionStore((state) => state.theme);
  const setTheme = useSessionStore((state) => state.setTheme);

  return (
    <Card>
      <h2 className="text-xl font-semibold">Settings</h2>
      <p className="mt-1 text-sm text-slate-400">Manage presentation and portfolio preferences.</p>

      <div className="mt-6 grid gap-4 md:grid-cols-2">
        <Field label="Preferred currency">
          <Select
            aria-label="Preferred currency"
            value={preferredCurrency}
            onChange={(event) => setPreferredCurrency(event.target.value as "USD" | "BRL")}
          >
            <option value="USD">USD</option>
            <option value="BRL">BRL</option>
          </Select>
        </Field>

        <Field label="Theme">
          <Select
            aria-label="Theme"
            value={theme}
            onChange={(event) => setTheme(event.target.value as Theme)}
          >
            <option value="dark">Dark</option>
            <option value="light">Light</option>
            <option value="neon">Neon</option>
          </Select>
        </Field>
      </div>
    </Card>
  );
}
