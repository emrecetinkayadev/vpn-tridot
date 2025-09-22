import { loadStripe, Stripe } from "@stripe/stripe-js";

let stripePromise: Promise<Stripe | null> | null = null;

export function getStripe(): Promise<Stripe | null> {
  if (!stripePromise) {
    const publicKey = process.env.NEXT_PUBLIC_STRIPE_PK;
    if (!publicKey) {
      return Promise.resolve(null);
    }

    stripePromise = loadStripe(publicKey);
  }

  return stripePromise;
}
