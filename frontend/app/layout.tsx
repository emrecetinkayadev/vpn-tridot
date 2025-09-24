import type { ReactNode } from "react";
import { AppShell } from "../components/layout/AppShell";
import "../styles/globals.css";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="tr">
      <body className="bg-slate-950 font-sans text-slate-50 antialiased">
        <AppShell>{children}</AppShell>
      </body>
    </html>
  );
}
