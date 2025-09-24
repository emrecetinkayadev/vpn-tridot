"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ReactNode, useMemo } from "react";

type NavItem = {
  href: string;
  label: string;
  hint?: string;
};

type NavSection = {
  title: string;
  items: NavItem[];
};

const NAV_SECTIONS: NavSection[] = [
  {
    title: "Overview",
    items: [
      {
        href: "/",
        label: "Dashboard",
        hint: "Traffic, billing, and fleet health",
      },
    ],
  },
  {
    title: "Control",
    items: [
      { href: "/devices", label: "Devices" },
      { href: "/regions", label: "Regions" },
      { href: "/plans", label: "Plans" },
    ],
  },
  {
    title: "Account",
    items: [
      { href: "/account", label: "Account" },
      { href: "/login", label: "Login" },
      { href: "/signup", label: "Signup" },
    ],
  },
];

export function AppShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();

  const environmentLabel = useMemo(() => {
    const env = process.env.NEXT_PUBLIC_DEPLOY_ENV ?? "local";
    return env.toUpperCase();
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-900 text-slate-50">
      <div className="mx-auto flex min-h-screen w-full max-w-7xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
        <header className="flex flex-col gap-4 rounded-2xl border border-slate-800/60 bg-slate-900/60 px-6 py-5 backdrop-blur">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-sky-500/10 text-sky-300">
                <span className="text-lg font-semibold">VPN</span>
              </div>
              <div>
                <p className="text-lg font-semibold tracking-tight">TriDot Control</p>
                <p className="text-xs text-slate-400">Operator panel for the VPN MVP</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <span className="rounded-full border border-slate-700/80 bg-slate-800/70 px-3 py-1 text-xs font-medium uppercase tracking-wide text-slate-300">
                {environmentLabel}
              </span>
              <button
                type="button"
                className="rounded-full border border-slate-700/70 bg-slate-900/80 px-3 py-1 text-xs font-medium text-slate-300 shadow-md transition hover:border-slate-600 hover:text-slate-100"
              >
                Contact Ops
              </button>
            </div>
          </div>
          <nav className="grid gap-3 sm:grid-cols-3">
            {NAV_SECTIONS.map((section) => (
              <div
                key={section.title}
                className="rounded-xl border border-slate-800/50 bg-slate-900/40 p-3"
              >
                <p className="px-1 text-xs font-semibold uppercase tracking-wide text-slate-500">
                  {section.title}
                </p>
                <ul className="mt-2 space-y-1">
                  {section.items.map((item) => {
                    const isActive =
                      pathname === item.href || pathname.startsWith(`${item.href}/`);
                    return (
                      <li key={item.href}>
                        <Link
                          href={item.href}
                          className={[
                            "group flex flex-col rounded-lg px-3 py-2 transition",
                            isActive
                              ? "bg-slate-800/80 text-slate-100 shadow"
                              : "text-slate-300 hover:bg-slate-800/40 hover:text-slate-100",
                          ].join(" ")}
                        >
                          <span className="text-sm font-medium">{item.label}</span>
                          {item.hint ? (
                            <span className="text-xs text-slate-400 group-hover:text-slate-300">
                              {item.hint}
                            </span>
                          ) : null}
                        </Link>
                      </li>
                    );
                  })}
                </ul>
              </div>
            ))}
          </nav>
        </header>

        <main className="flex-1 rounded-2xl border border-slate-800/60 bg-slate-900/60 px-6 py-8 shadow-inner backdrop-blur">
          {children}
        </main>

        <footer className="rounded-2xl border border-slate-800/60 bg-slate-900/50 px-6 py-4 text-xs text-slate-500">
          © {new Date().getFullYear()} TriDot VPN — Privacy-first network. All rights reserved.
        </footer>
      </div>
    </div>
  );
}
