import { env } from "@/lib/env";

export type CheckoutResponse = {
  provider: "stripe" | "iyzico";
  checkoutUrl: string;
};

export async function createCheckoutSession(planCode: string): Promise<CheckoutResponse> {
  const response = await fetch("/api/checkout", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ planCode }),
  });

  if (!response.ok) {
    throw new Error("Checkout oturumu oluşturulamadı");
  }
  return response.json();
}

export const stripeEnv = {
  publishableKey: env.stripePublishableKey,
};
