# VPN MVP — Proje Planı ve Dosyalama Yapısı

Bu doküman, VPN MVP projesi için **tam kapsamlı** klasör/ dosya hiyerarşisini, adlandırma kurallarını, standardları, şablonları ve araçları açıklar. Amaç: hızlı onboarding, öngörülebilir yapı, düşük operasyon maliyeti.

> Kapsam: Monorepo, FE/BE ayrı; Node Agent; Infra; CI/CD; Observability; Security; Docs; Templates; Runbooks.

---

## 1) Üst Düzey Depo Ağacı

```
repo-root/
  .github/
    ISSUE_TEMPLATE/
      bug_report.md
      feature_request.md
      tech_debt.md
    workflows/
      ci-backend.yaml
      ci-frontend.yaml
      ci-agent.yaml
      security-scan.yaml
      release.yaml
  .vscode/
    settings.json
    extensions.json
  .devcontainer/
    devcontainer.json
    docker-compose.dev.yaml
  .editorconfig
  .gitattributes
  .gitignore
  CODEOWNERS
  CONTRIBUTING.md
  SECURITY.md
  GOVERNANCE.md               # küçük ekiplerde opsiyonel
  LICENSE                     # tescilli/OSS
  README.md                   # kök: kısa yönlendirme

  docs/
    architecture.md
    privacy.md
    adr/
      0001-architecture-decisions.md
      0002-why-go-wgctrl.md
    runbooks/
      incident-node-down.md
      rotate-wg-keys.md
      restore-db.md
    api/
      openapi.yaml            # REST şeması
      grpc/
        peers.proto
        regions.proto
      examples/
        peers.http            # REST örnek istekler (REST Client)
    product/
      prd.md                  # PRD’nin kaynak hali
      roadmap.md
      metrics-kpis.md

  scripts/
    bootstrap.sh              # repo init, pre-commit, hooks
    leaktest.sh               # DNS/IPv6 leak testi
    loadtest.sh               # iperf3 vb.
    gen-config-url.sh         # tek‑kullanımlık link üretim testi
    db/
      migrate.sh
      seed.sh

  tools/
    sbom/
      generate-sbom.sh
    lint/
      golangci.yml
      eslint.config.mjs
      stylelint.config.cjs

  deploy/
    docker/
      backend.Dockerfile
      agent.Dockerfile
      frontend.Dockerfile
      compose.node.yaml
    k8s/                      # v1.0+
      base/
        backend-deployment.yaml
        backend-service.yaml
        agent-daemonset.yaml
        prometheus.yaml
        grafana.yaml
        loki.yaml
      overlays/
        staging/
          kustomization.yaml
        prod/
          kustomization.yaml

  infra/
    terraform/
      modules/
        vpc/
          main.tf
          variables.tf
          outputs.tf
        vm/
          main.tf
          variables.tf
          outputs.tf
        dns/
        monitoring/
      envs/
        staging/
          backend.tfvars
          main.tf
        prod/
          backend.tfvars
          main.tf
    ansible/
      inventories/
        staging/hosts.yml
        prod/hosts.yml
      roles/
        common/
          tasks/main.yml
        wireguard/
          tasks/main.yml
          templates/wg0.conf.j2
        agent/
          tasks/main.yml
          templates/agent.env.j2
      playbooks/
        node.yaml
        rotate-keys.yaml

  backend/
    cmd/
      api/
        main.go
    internal/
      auth/
        jwt.go
        totp.go
      billing/
        stripe.go
        iyzico.go
        webhooks.go
      peers/
        service.go
        handler.go
        repository.go
      regions/
        service.go
        handler.go
      nodes/
        service.go
        handler.go
        scheduler.go          # kapasite tabanlı yerleştirme
      storage/
        postgres/
          db.go
          migrations/
            0001_init.sql
            0002_peers.sql
        redis/
          cache.go
      wg/
        ctrl.go               # wgctrl sarmalayıcı
        config.go             # conf üretimi
      api/
        middleware/
          auth.go
          ratelimit.go
        router.go
      observability/
        metrics.go
        logging.go
    pkg/
      types/
        models.go
        dto.go
      utils/
        config.go
        errors.go
    test/
      e2e/
        peers_e2e_test.go
      unit/
        peers_service_test.go
    Makefile
    go.mod
    go.sum

  node-agent/
    cmd/agent/main.go
    internal/
      wg/
        iface.go
        peer.go
      health/
        reporter.go
      rpc/
        client.go             # mTLS
      config/
        env.go
    pkg/
      logger/logger.go
    test/
      unit/
        wg_iface_test.go
    Makefile
    go.mod
    go.sum

  frontend/
    app/
      layout.tsx
      page.tsx
      (auth)/
        login/page.tsx
        signup/page.tsx
      (billing)/
        plans/page.tsx
        success/page.tsx
      (devices)/
        page.tsx
        [peerId]/config/page.tsx
      (regions)/
        page.tsx
      (account)/
        page.tsx
    components/
      ui/*                     # shadcn
      forms/*
      charts/*
    lib/
      api.ts
      auth.ts
      stripe.ts
      iyzico.ts
    public/
      icons/
      images/
    styles/
      globals.css
    tests/
      unit/
        devices.test.ts
      e2e/
        devices.spec.ts       # Playwright
    package.json
    pnpm-lock.yaml
    next.config.ts
    tsconfig.json
    playwright.config.ts

  ops/
    grafana/
      dashboards/
        nodes-overview.json
        peers-overview.json
    prometheus/
      prometheus.yml
    loki/
      config.yml
    alertmanager/
      alertmanager.yml

  .pre-commit-config.yaml
  CHANGELOG.md
  RELEASE_NOTES.md
```

---

## 2) Adlandırma ve Standartlar

* **Klasörler:** kebab-case (`node-agent`, `infra`, `runbooks`).
* **Go paketleri:** lower\_snake\_case değil, **paket ismi kısa** (`wg`, `auth`, `rpc`).
* **TypeScript dosyaları:** `camelCase.ts` bileşenler `PascalCase.tsx`.
* **Migrasyon dosyaları:** artan numara + açıklama (`0003_add_sessions.sql`).
* **Branch stratejisi:** `main` korumalı, `develop` opsiyonel; feature → `feat/<scope>-<short>`.
* **Commit mesajları:** Conventional Commits (`feat(peers): add server-side key gen`).
* **Versiyonlama:** SemVer. API kırıcı değişiklik → minor/major.

---

## 3) Ortamlar ve Konfig Yönetimi

* **Ortamlar:** `local`, `staging`, `prod`.
* **Gizli bilgiler:** SOPS/age veya Vault. `.env` commit edilmez.
* **Konfig yükleme:**

  * Backend: `APP_ENV`, `POSTGRES_DSN`, `REDIS_URL`, ödeme anahtarları, mTLS materyalleri.
  * Agent: `CONTROL_PLANE_URL`, `AGENT_TOKEN`, mTLS client cert.
  * Frontend: `NEXT_PUBLIC_API_BASE`, `NEXT_PUBLIC_STRIPE_PK`.
* **Config URL TTL:** 24h, tek kullanım.

---

## 4) CI/CD Boru Hatları

* **ci-backend.yaml**

  * Setup Go → build → unit test → vuln scan (gosec, grype) → cache.
* **ci-frontend.yaml**

  * Setup Node → pnpm i → typecheck → unit test → Playwright e2e (staging).
* **ci-agent.yaml**

  * Build → unit test → container image.
* **security-scan.yaml**

  * secrets scanning (gitleaks), SBOM (syft), image scan (grype).
* **release.yaml**

  * Tag → changelog → images push → GitHub Release.

---

## 5) İş Akışları ve Kanban

* **Epics:** Auth & Billing, Peer Lifecycle, Infra Automation, Observability, Security Hardening, Support.
* **Swimlanes:** Feature, Bugs, Tech Debt, Ops.
* **WIP limitleri:** Feature ≤ 3, Bugs ≤ 5.
* **Definition of Ready:** Kullanıcı senaryosu, kabul kriteri, ölçütler tanımlı.
* **Definition of Done:** Testler geçer, dokümantasyon güncel, izleme panoları güncel, runbook varsa eklendi.

---

## 6) Test Düzeni

* **Backend**

  * `internal/*` için unit test oranı ≥ %70
  * `test/e2e` içinde uçtan uca akış: signup → checkout (mock) → peer create → config fetch
* **Agent**

  * `wg` soyutlamaları için birim testleri, `netlink` mocked.
* **Frontend**

  * Unit: bileşenler ve hooks
  * E2E: Playwright ile `devices` ve `plans` akışı
* **Load/Leak**

  * `scripts/loadtest.sh` ve `scripts/leaktest.sh`

---

## 7) Güvenlik ve Uyumluluk

* **Policy dosyaları:** `SECURITY.md`, `privacy.md`.
* **Ratelimit & bot koruması:** hCaptcha, IP tabanlı limit.
* **mTLS:** CA yönetimi, rotasyon runbook’u.
* **SBOM:** `tools/sbom/generate-sbom.sh`
* **KVKK/GDPR:** Veri envanteri, saklama süreleri, silme talebi akışı `docs/privacy.md`.

---

## 8) Operasyon ve İzleme

* **Prometheus**: node ve uygulama metrikleri
* **Grafana**: `ops/grafana/dashboards/*`
* **Alertmanager**: uptime, error rate, webhook failure eşikleri
* **Loglama**: Loki, `ops/loki/config.yml`
* **Runbook’lar**: `docs/runbooks/*`

---

## 9) Örnek Makefile Hedefleri

**backend/Makefile**

```
.PHONY: build test migrate run
build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

migrate:
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(POSTGRES_DSN)" goose up

run:
	APP_ENV=local ./bin/api
```

**node-agent/Makefile**

```
.PHONY: build test run
build:
	go build -o bin/agent ./cmd/agent

test:
	go test ./...

run:
	./bin/agent
```

---

## 10) Şablonlar ve Örnekler

* **Issue Templates:** Bug/Feature/Tech Debt.
* **PR Template:** etki alanı, risk, test kapsamı, rollback planı.
* **HTTP Örnekleri:** `docs/api/examples/*.http` REST Client ile çağırmak için.
* **OpenAPI & Protobuf:** `docs/api/openapi.yaml`, `docs/api/grpc/*.proto`.

---

## 11) Yol Haritası ve Milestone’lar

* **M1 (Hafta 1–2):** Monorepo iskelet, auth iskelet, DB migrasyonları, Stripe sandbox.
* **M2 (Hafta 3–4):** Agent mTLS, peer CRUD, config/QR, Terraform+Ansible staging.
* **M3 (Hafta 5–6):** İzleme, alerting, SSS, leak/load testleri.
* **M4 (Hafta 7–8):** Kapalı beta, fiyatlandırma, dokümantasyon ve runbook’lar.

**Kabul Kriterleri:** PRD’deki MVP “Done” maddeleri + bu dosyadaki runbook ve panoların hazır olması.

---

## 12) RACI (Örnek)

* **Product (Emre):** R/A — kapsam, öncelik, kabul
* **Backend Lead:** R — API, DB, billing
* **Infra/DevOps:** R — Terraform, Ansible, CI/CD
* **Frontend:** R — panel, QR/CONF
* **Security:** C — mTLS, gizlilik, SBOM
* **Support:** C/I — biletleme, SSS

---

## 13) Sürümleme ve Dağıtım Stratejisi

* **Kanallar:** `staging` → `prod`.
* **Release:** `release.yaml` tetikler; tag → imajlar → helm/kustomize rollout.
* **Rollback:** önceki imaj etiketi, DB backward‑compatible migrasyon önceliği.

---

## 14) Kalite Kapıları

* Lint ve testler zorunlu
* Code review 1 onay
* Açık kritik güvenlik bulgusu yok
* Doküman güncellemesi kontrol listesi

---

## 15) Onboarding Check‑List

* Erişim: repo, container registry, cloud hesapları
* Araçlar: Go, Node, Docker, Terraform, Ansible
* Secrets: SOPS/age key, vault erişimi
* İlk görev: `peers` servisi için unit test ekle ve çalıştır

---

## 16) Ek Notlar

* Mobil uygulamalar (v1.1) için `mobile/` klasörü hazırlanabilir (Flutter/React Native).
* Obfuscation protokolleri (v1.2) için `transport/` modülü ve ayrı `gateway` hizmeti planlanır.

