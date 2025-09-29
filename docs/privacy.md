# TriDot VPN Privacy Notice (KVKK / GDPR)

_Last updated: 2024-05-13_

This notice explains how TriDot Teknoloji A.Ş. ("TriDot", "we") processes personal data for the TriDot VPN product. It covers operations under the Turkish Personal Data Protection Law (KVKK) and the EU General Data Protection Regulation (GDPR).

## 1. Data Controller
- TriDot Teknoloji A.Ş., İstanbul, Türkiye
- Contact: `privacy@tridot.dev`
- Data Protection Officer: delegated to the Operations team (ops@tridot.dev)

## 2. Personal Data We Process
| Category | Fields | Purpose |
| --- | --- | --- |
| Account | Name, surname, e-mail, password hash, two-factor secrets | Account registration, authentication, customer support |
| Billing | Plan selection, Stripe/Iyzico customer IDs, payment tokens, invoices | Subscription management, statutory accounting |
| Operations | Region preferences, device identifiers, WireGuard public keys, peer names | Provisioning VPN peers and device management |
| Telemetry (pseudonymised) | Last handshake timestamp, bytes up/down, agent health metrics | Capacity planning, abuse detection, incident response |
| Security | IP address (ephemeral), hCaptcha score, audit logs | Fraud prevention, service protection |

We do **not** log traffic content, DNS queries, or browsing history. Session metrics are stored as aggregates associated with peer identifiers.

## 3. Legal Bases and Processing Purposes
| Purpose | Legal Basis |
| --- | --- |
| Provide VPN connectivity and manage subscriptions | GDPR Art. 6(1)(b), KVKK Art. 5(2)(c) (performance of contract) |
| Compliance with tax and accounting obligations | GDPR Art. 6(1)(c), KVKK Art. 5(2)(ç) |
| Fraud detection, network security, abuse prevention | GDPR Art. 6(1)(f), KVKK Art. 5(2)(f) (legitimate interest) |
| Operational communications (incidents, service updates) | GDPR Art. 6(1)(b)/(f), KVKK Art. 5(2)(c)/(f) |
| Marketing communications (optional) | GDPR Art. 6(1)(a), KVKK Art. 5(1) (explicit consent) |

## 4. Retention Schedule
| Data | Retention |
| --- | --- |
| Account details | Contract term + 24 months |
| Billing records | 10 years (tax law) |
| VPN session metrics | 30 days |
| IP addresses & audit logs | 7 days unless required for investigations |
| Support tickets | Contract term + 12 months |

## 5. Processors and Transfers
- Stripe, Inc. (EU data residency) – payment processing
- Iyzico Ödeme Hizmetleri A.Ş. – local card processing and 3D Secure
- Amazon Web Services EMEA SARL – infrastructure hosting (Frankfurt region)
- HelpScout Inc. – customer support mailbox (EU data center)
- Open-source observability stack (Prometheus, Loki, Grafana) hosted on our own infrastructure

International transfers outside Türkiye/EU use standard contractual clauses (GDPR Art. 46) or KVKK Art. 9 approvals. Copies of safeguards are available upon request.

## 6. Your Rights
Data subjects can exercise the following rights:
- Access, rectification, erasure, restriction (KVKK Art. 11; GDPR Arts. 15-18)
- Data portability (GDPR Art. 20)
- Object to processing based on legitimate interests (GDPR Art. 21; KVKK Art. 11/1-ğ)
- Withdraw explicit consent for marketing at any time without affecting core service

Send requests to `privacy@tridot.dev`. We reply within 30 days (KVKK) / 1 month (GDPR).

## 7. Consent Withdrawal & Preferences
- Use the account portal to manage marketing preferences.
- For manual withdrawal, e-mail `privacy@tridot.dev` or open a support ticket.
- We log consent timestamps linked to user ID for auditing.

## 8. Security Measures
- Mutual TLS between control plane and node agents
- Encrypted secrets storage via Vault/SOPS
- Role-based access control and just-in-time SSH
- Continuous security scanning (SAST, dependency scans, container image signing)

## 9. Automated Decision Making
We do not perform automated decision making or profiling with legal effects.

## 10. Updates
We version and timestamp this notice in Git. Substantive changes are announced via e-mail and the status page. The latest version is available at `https://vpn.tridot.dev/privacy`.

---
If you believe your rights have been violated, contact us first. You also have the right to lodge complaints with the Turkish Data Protection Authority (KVKK) or your local EU supervisory authority.
