import type { ButtonHTMLAttributes, HTMLAttributes, InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from "react";

type CommonProps = {
  className?: string;
  children?: ReactNode;
};

export function Card({ className = "", children, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={`glass rounded-3xl p-5 text-slate-100 ${className}`}
      {...props}
    >
      {children}
    </div>
  );
}

export function Button({ className = "", children, type = "button", ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type={type}
      className={`btn-primary inline-flex items-center justify-center rounded-xl border px-4 py-2 font-medium transition disabled:cursor-not-allowed disabled:opacity-60 ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

export function SecondaryButton({ className = "", children, type = "button", ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type={type}
      className={`btn-secondary inline-flex items-center justify-center rounded-xl border px-4 py-2 font-medium transition ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

export function DangerButton({ className = "", children, type = "button", ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type={type}
      className={`btn-danger inline-flex items-center justify-center rounded-xl border px-4 py-2 font-medium transition disabled:cursor-not-allowed disabled:opacity-60 ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

export function Input({ className = "", ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={`form-control w-full rounded-xl border px-3 py-2 outline-none transition ${className}`}
      {...props}
    />
  );
}

export function Select({ className = "", children, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={`form-control w-full rounded-xl border px-3 py-2 outline-none transition ${className}`}
      {...props}
    >
      {children}
    </select>
  );
}

export function Field({ label, error, children }: CommonProps & { label: string; error?: string }) {
  return (
    <label className="block space-y-2">
      <span className="text-sm text-slate-300">{label}</span>
      {children}
      {error ? <p className="text-sm text-rose-300">{error}</p> : null}
    </label>
  );
}

export function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <Card className="text-center">
      <h3 className="text-lg font-semibold">{title}</h3>
      <p className="mt-2 text-sm text-slate-400">{description}</p>
    </Card>
  );
}
