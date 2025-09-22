# AGENTS.md — VPN‑MVP

This file guides coding agents (Codex, etc.) on how to work in this repo. Keep it in the repo root.

## Goals

* Ship MVP in 6–8 weeks.
* WireGuard data‑plane, Go control‑plane, Next.js panel.
* Full auto provisioning with Terraform + Ansible + Node Agent.
* Privacy by design. No traffic/content logs.

## Project Map

```
backend/       Go API (REST ext, gRPC int)
node-agent/    Go daemon on VPN nodes (wgctrl, health)
frontend/      Next.js 15 panel (TS, Tailwind v4, shadcn)
infra/         Terraform + Ansible
deploy/        Dockerfiles, compose, k8s (v1+)
ops/           Prometheus, Grafana, Loki, Alertmanager
scripts/       leak/load tests, db utils
docs/          PRD, architecture, runbooks, privacy, ADRs
```

## Setup

* Prereqs: Go 1.22+, Node 22+, pnpm, Docker 24+, Postgres 16+, Redis 7+, Terraform 1.8+, Ansible 2.16+.
* Secrets via SOPS/Vault. Never commit `.env`.

### Install

```bash
pnpm -C frontend i
(cd backend && go mod download)
```

### Dev stack

```bash
# DB + Redis for local dev (provide your own file or compose)
# then run services:
(cd backend && go run ./cmd/api)
(cd frontend && pnpm dev)
```

## Commands

* **Backend**

  * Build: `cd backend && go build ./cmd/api`
  * Test: `cd backend && go test ./...`
  * Migrate: `GOOSE_DRIVER=postgres GOOSE_DBSTRING=$POSTGRES_DSN goose up`
* **Agent**

  * Build: `cd node-agent && go build ./cmd/agent`
  * Test: `cd node-agent && go test ./...`
* **Frontend**

  * Dev: `pnpm -C frontend dev`
  * Build: `pnpm -C frontend build`
  * Test: `pnpm -C frontend test && pnpm -C frontend exec playwright test`
* **Infra (staging)**

  * `terraform -chdir=infra/terraform/envs/staging init && terraform -chdir=infra/terraform/envs/staging apply -auto-approve`
  * `ansible-playbook -i infra/ansible/inventories/staging/hosts.yml infra/ansible/playbooks/node.yaml`
* **Checks**

  * Lint: `golangci-lint run ./...` and `pnpm -C frontend lint`
  * SBOM: `syft packages . -o json > tools/sbom/sbom.json`
  * Image scan: `grype dir:. --fail-on high`
  * Secrets: `gitleaks detect --no-git -v`

## Tests

* **Unit**: backend auth/webhooks/peers; agent wg/health; frontend components/hooks.
* **E2E**: backend `test/e2e` peer flow; frontend Playwright devices flow.
* **Leak/Load**: `scripts/leaktest.sh`, `scripts/loadtest.sh`, `iperf3` target ≥ 2 Gbps agg/node.

## Security Guards

* Confirm before running shell commands or package installs.
* Zero Trust policy. 
* Read‑only: `docs/**`, `backend/internal/storage/postgres/migrations/**`.
* FS scope: `backend/`, `frontend/`, `node-agent/`, `infra/`, `deploy/`, `ops/`, `scripts/`, `docs/`.
* Network allowlist: `api.github.com`, `registry.npmjs.org`, `pypi.org`, `files.pythonhosted.org`, `dl.google.com`, `objects.githubusercontent.com`. Block others by default.
* Secrets required (from env or Vault/SOPS): `POSTGRES_DSN`, `REDIS_URL`, `STRIPE_SECRET`, `STRIPE_WEBHOOK_SECRET`, `IYZICO_API_KEY`, `IYZICO_SECRET_KEY`, `JWT_SECRET`, `MTLS_CA_PEM`, `MTLS_SERVER_CERT`, `MTLS_SERVER_KEY`.

## Conventions

* Conventional Commits; SemVer; protected `main`.
* Code style: golangci-lint defaults; ESLint + TS strict; Tailwind v4.
* API breaking changes require version bump and doc updates.

## Release Checklist

* Tests green; leak/load checks pass.
* SBOM generated; `grype` high/crit = 0.
* Secrets mounted via Vault/SOPS; no `.env`.
* Staging canary OK; then prod rollout.
* Update `CHANGELOG.md`, `SECURITY.md` if relevant.

## Runbooks

* `docs/runbooks/incident-node-down.md`
* `docs/runbooks/rotate-wg-keys.md`
* `docs/runbooks/restore-db.md`

## Knowledge

* PRD: `docs/product/prd.md`
* Architecture: `docs/architecture.md`
* Project plan: `docs/PROJECT_PLAN.md`
* Structure: `docs/PROJECT_STRUCTURE.md`

---

> Keep this file short, actionable, and always accurate. Agents read this first.
