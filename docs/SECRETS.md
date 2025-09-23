# Secrets Yönetimi

Bu proje, gizli anahtarları çalışma zamanına taşımak için iki yönteme hazırdır: **SOPS dosyaları** ve **HashiCorp Vault (KV v2)**. Her iki yöntem de uygulama başlamadan önce environment değişkenlerini besler.

## SOPS

1. SOPS ile şifrelenmiş JSON/YAML dosyası oluşturun (ör. `secrets/backend.json`).
2. Çalışma zamanında dosyayı aşağıdaki değişkenlerle işaretleyin:
   ```env
   SOPS_SECRETS_ENABLED=true
   SOPS_SECRETS_PATH=/run/secrets/backend.json
   SOPS_SECRETS_FORMAT=json  # json|yaml|env desteklenir
   ```
3. Dosya deşifre edilmiş olarak konteynerde mevcut olmalıdır (ör. `sops -d ... > /run/secrets/backend.json`).
4. Dosyadaki anahtar–değer çiftleri environment değişkeni olarak uygulanır.

## HashiCorp Vault

1. KV v2 secret’ınızı aşağıdaki gibi yerleştirin:
   ```bash
   vault kv put kv/vpn-backend/prod JWT_SECRET=... STRIPE_SECRET=...
   ```
2. Backend için environment değişkenleri:
   ```env
   VAULT_ENABLED=true
   VAULT_ADDR=https://vault.internal.example.com
   VAULT_TOKEN=s.xxxxxx
   VAULT_PATH=kv/data/vpn-backend/prod
   VAULT_NAMESPACE= # opsiyonel
   VAULT_TIMEOUT=5s
   VAULT_TLS_SKIP_VERIFY=false
   ```
3. Uygulama TLS doğrulamasını devre dışı bırakmak için `VAULT_TLS_SKIP_VERIFY=true` seçeneğine sahiptir (yerel testler için).
4. Vault yanıtında dönen tüm değerler (KV v2 için `data.data`) environment değişkeni olarak uygulanır.

## Çalışma Zamanı Akışı

1. `config.Load()` ilk çağrıldığında secret kaynakları okunur.
2. Aktif bir kaynak varsa secrets manager bu değerleri `os.Setenv` ile uygular.
3. Konfigürasyon ikinci kez yüklenir; böylece Postgres DSN, Stripe anahtarları vb. güncel env değerleri kullanır.

### GitHub Actions

`ci-backend` ve `release` workflow’ları, aşağıdaki repository secret’ları opsiyonel olarak kullanarak aynı mekanizmanın CI/CD’de çalışmasını sağlar:

| Secret Adı | Açıklama |
|------------|----------|
| `SOPS_SECRETS_FILE_B64` | Base64 olarak kodlanmış şifreli SOPS dosyası |
| `SOPS_AGE_KEY` | SOPS dosyasını çözmek için age private key |
| `VAULT_TOKEN` | Vault erişim tokenı |
| `VAULT_ADDR` | Vault adresi (örn. https://vault.internal.example.com) |
| `VAULT_PATH` | KV v2 secret yolu (örn. kv/data/vpn-backend/ci) |
| `VAULT_NAMESPACE` | Opsiyonel Vault namespace |
| `COSIGN_PRIVATE_KEY` | Release artefactlarını imzalamak için cosign private key (PEM) |
| `COSIGN_PASSWORD` | Cosign key parolasını saklar (varsa) |
| `COSIGN_PUBLIC_KEY` | Release doğrulaması için cosign public key (opsiyonel; imza doğrulaması step'ini aktif eder) |

Secrets varsa workflow şu adımları çalıştırır:

1. `tools/secrets/sops-decrypt.sh` → `/tmp/backend.secrets.json` dosyasını üretir.
2. `tools/secrets/vault-export.sh` → `/tmp/vault.env` dosyasını üretir ve değerleri `GITHUB_ENV`’e yazar.
3. `SECRETS_BOOTSTRAPPED=true` işaretlenir; backend yeniden `config.Load()` çalıştırdığında environment değerleri hazırdır.
4. Cosign secrets sağlanmışsa `sigstore/cosign` kurulup üretilen ikililer imzalanır; `COSIGN_PUBLIC_KEY` de sağlanmışsa imzalar release’e yüklemeden önce doğrulanır.

`ci-backend` job’u `staging`, `release` job’u `production` environment’ına bağlıdır; environment seviyesindeki yaptırım/approval mekanizmaları GitHub üzerinden yönetilir.

## Önerilen Dizayn

- **Production**: Vault birincil kaynak, SOPS ise acil durum/offline kurtarma için.
- **Dev**: SOPS veya `.env.local` dosyası; Vault gerekli değil.
- CI pipeline’larında secrets değerlerini hiçbir log’a yazmayın. Testlerde `SOPS_SECRETS_ENABLED=false` bırakın.

## Yardımcı Komutlar

```bash
# SOPS ile şifreli dosya oluşturma (AES-GCM + age anahtarı)
sops --encrypt --age "age1..." secrets/backend.json > secrets/backend.enc.json

# Deşifre edip runtime'a yazma
envsubst < secrets/backend.enc.json | sops -d > /run/secrets/backend.json
```

Daha fazla bilgi için `tools/secrets/` altındaki örnek scriptlere bakın.
