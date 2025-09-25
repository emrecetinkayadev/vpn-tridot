# Container İmzalama Runbook (cosign)

## Amaç
TriDot VPN container imajlarının cosign ile imzalanmasını ve doğrulanmasını otomatikleştirmek.

## Gereksinimler
- GitHub Actions `release` workflow’unda cosign installer step’i mevcut.
- `COSIGN_PRIVATE_KEY`, `COSIGN_PASSWORD` ve `COSIGN_PUBLIC_KEY` secretları Vault/SOPS üzerinden sağlanır.

## Adımlar
1. **Anahtar Üretimi**
   ```bash
   cosign generate-key-pair
   ```
   - `cosign.key` ve `cosign.pub` dosyalarını Vault’a ekleyin.
2. **CI Pipeline Güncelleme**
   - `release.yaml` içinde `cosign sign-blob` adımı zaten ikilileri imzalar.
   - Container imajları için build step’i sonrası:
     ```bash
     cosign sign --key "$COSIGN_PRIVATE_KEY" ghcr.io/tridot/backend:${GITHUB_SHA}
     cosign sign --key "$COSIGN_PRIVATE_KEY" ghcr.io/tridot/node-agent:${GITHUB_SHA}
     ```
3. **Doğrulama**
   ```bash
   cosign verify --key cosign.pub ghcr.io/tridot/backend:${GITHUB_SHA}
   ```
4. **Policy**
   - Kubernetes admission controller (kyverno/gatekeeper) ile imzasız imajları reddedin.

## Rollback
- Cosign imzası bozulursa eski imaj tag’ı kullanılabilir.

## Notlar
- Cosign key’leri rotasyon runbook’u ekleyin.
- GitOps pipeline’larında doğrulama step’i şart.
