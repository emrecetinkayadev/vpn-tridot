# SBOM Üretimi ve Güvenlik Taramaları

Bu klasör, TriDot VPN projeleri için Software Bill of Materials (SBOM) ve container taramalarını yürütmek amacıyla kullanılan komutları belgelemek içindir.

## Syft ile SBOM

Tüm depo için SBOM üretmek:
```bash
syft packages . -o json > tools/sbom/sbom.json
```

Staging build pipeline’ında aşağıdaki adımı planlayın:
```bash
syft packages . -o cyclonedx-json > tools/sbom/sbom-cyclonedx.json
```

## Grype ile İmaj Taraması

```bash
grype dir:. --fail-on high
```

CI’da `security-scan` workflow’u halihazırda bu komutları çalıştırır. Lokal geliştirme için `tools/` klasöründe saklandığı gibi MANIFEST dosyalarını `.gitignore` altında tutun.
