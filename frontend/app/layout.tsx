import type { ReactNode } from "react";
import "../styles/globals.css";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="tr">
      <body className="min-h-screen bg-slate-950 text-slate-50">
        {children}
      </body>
    </html>
  );
}
