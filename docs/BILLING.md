# Billing & Payments

## Providers

* **Stripe**: Production-ready. Checkout sessions created via `POST /api/v1/checkout`, webhook listener at `POST /api/v1/webhooks/stripe`.
* **Iyzico**: Placeholder implementation. API keys/URL present in `.env` templates for future integration.

## API Endpoints

### `GET /api/v1/plans`
Returns active subscription plans including price, currency, device limits.

### `POST /api/v1/checkout`
Requires JWT auth. Accepts `{ "plan_code": "vpn-monthly", "provider": "stripe" }`. Responds with checkout session URL and provider session id.

### `POST /api/v1/webhooks/stripe`
Validates incoming Stripe signatures. Supported events:

* `checkout.session.completed`
* `customer.subscription.updated`
* `customer.subscription.deleted`
* `invoice.payment_succeeded`

Events are normalized into subscription/payment updates with idempotent upserts.

### `GET /api/v1/account/payments`
JWT-protected endpoint returning historic payment list for the current user.

## Database Entities

* `plans`: code, name, description, price (cents), currency, billing period, interval, device limit.
* `subscriptions`: user, plan, status (`trialing|active|past_due|canceled`), provider identifiers, period start/end.
* `payments`: subscription, provider payment id, amount, currency, paid/refunded timestamps, metadata.

## Seed Data

`billing.Service.SeedDefaultPlans` inserts initial plans:

| Code           | Name     | Period | Price (TRY) |
| -------------- | -------- | ------ | ----------- |
| vpn-monthly    | Aylık    | month  | 149.00      |
| vpn-quarterly  | 3 Aylık  | month* | 399.00      |
| vpn-annual     | Yıllık   | year   | 1,299.00    |

> *Quarterly plan uses interval count 3.

## Environment Variables

Refer to `backend/.env.example` for full list:

* `STRIPE_SECRET`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_SUCCESS_URL`, `STRIPE_CANCEL_URL`
* `IYZICO_API_KEY`, `IYZICO_SECRET_KEY`, `IYZICO_BASE_URL`
* `BILLING_DEFAULT_CURRENCY`

## Future Work

* Implement full Iyzico checkout/webhook pipeline.
* Tie subscription status to device quota enforcement.
* Expand reporting (refunds, invoices) for support tooling.
