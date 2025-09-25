"use client";
import { useState } from "react";

import { createCheckoutSession } from "@/lib/stripe";

const PLANS = [
  {
    name: "Monthly",
    price: "₺149",
    description: "Billed every month, includes 5 devices",
    features: ["5 peers", "All regions", "Email support"],
    code: "vpn-monthly",
  },
  {
    name: "Quarterly",
    price: "₺399",
    description: "Save 10% with quarterly billing",
    features: ["5 peers", "Priority support", "Early access regions"],
    code: "vpn-quarterly",
  },
  {
    name: "Annual",
    price: "₺1.399",
    description: "Best value, two months free",
    features: ["5 peers", "Dedicated success", "Usage analytics"],
    code: "vpn-annual",
  },
];

export default function PlansPage() {
  const [isLoading, setIsLoading] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const onCheckout = async (planCode: string) => {
    try {
      setIsLoading(planCode);
      setError(null);
      const session = await createCheckoutSession(planCode);
      window.location.href = session.checkoutUrl;
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsLoading(null);
    }
  };

  return (
    <div className="space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Subscription plans</h1>
        <p className="text-sm text-slate-400">
          Stripe + Iyzico checkout flows will plug in here. For now we expose the pricing skeleton to
          align on copy and layout.
        </p>
        {error ? <p className="text-sm text-rose-300">{error}</p> : null}
      </header>

      <section className="grid gap-4 md:grid-cols-3">
        {PLANS.map((plan) => (
          <div
            key={plan.name}
            className="flex h-full flex-col rounded-xl border border-slate-800 bg-slate-900/70 p-6"
          >
            <div className="space-y-2">
              <h2 className="text-lg font-semibold text-slate-100">{plan.name}</h2>
              <p className="text-3xl font-semibold text-slate-50">{plan.price}</p>
              <p className="text-xs text-slate-400">{plan.description}</p>
            </div>
            <ul className="mt-4 space-y-2 text-sm text-slate-300">
              {plan.features.map((feature) => (
                <li key={feature} className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-sky-400"></span>
                  {feature}
                </li>
              ))}
            </ul>
            <button
              type="button"
              className="mt-6 inline-flex items-center justify-center rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm font-medium text-slate-100 transition hover:border-slate-600 hover:bg-slate-700"
              onClick={() => onCheckout(plan.code)}
              disabled={Boolean(isLoading)}
            >
              {isLoading === plan.code ? "Yönlendiriliyor..." : "Planı seç"}
            </button>
          </div>
        ))}
      </section>
    </div>
  );
}