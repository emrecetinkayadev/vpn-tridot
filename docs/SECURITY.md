# SECURITY.md

Güvenlik politikası, raporlama süreci, hardening kuralları ve olay yönetimi. Bu depo **VPN MVP** projesine aittir.

## 1) İletişim ve Zafiyet Bildirimi

* E‑posta: **[security@yourdomain.example](mailto:security@yourdomain.example)**
* PGP: `pubkey.asc` (repo kökünde)
* security.txt: `/.well-known/security.txt` (prod ortamda servis edilir)
* Acil durum (aktif sömürü): Konu başına **\[URGENT]** ekleyin.

**Sorumlu açıklama**: 90 güne kadar koordineli yayın. PoC yeterli. Exploit paylaşımı gizli tutulur.

### Nasıl rapor edilir

1. Kapsamı ve etkiyi yazın.
2. Tekrarlanabilir adımlar, loglar, ekran görüntüleri.
3. Etkilenen sürümler, commit veya imaj etiketi.
4. Varsa öneri/düzeltme.

Şablon: `docs/security/vulnerability_report_template.md`.

## 2) Kapsam

* **Dahil**: Backend API, Node Agent, Frontend panel, Terraform/Ansible tanımları, konteyner imajları.
* **Hariç**: Üçüncü taraf resmi WireGuard istemcileri, kullanıcı cihazları, tarayıcı eklentileri (v1.1+).

## 3) Desteklenen Sürümler

| Sürüm            | Destek  | Güvenlik yamaları  |
| ---------------- | ------- | ------------------ |
| v1.x (son minör) | Evet    | Evet               |
| v1-(n-1)         | Kısıtlı | Kritik düzeltmeler |
| daha eski        | Hayır   | Yok                |

## 4) Tehdit Modeli (Özet)

* Ağ pasif/aktif dinleme, orta adam, trafik sınıflandırma (DPI), IP kara listeleme.
* Kontrol düzlemine yetkisiz erişim denemeleri, API suistimali, oran aşımı.
* Node ele geçirilmesi, gizli anahtar sızıntısı.
* Tedarik zinciri: bağımlılık, container, CI gizli bilgileri.
* Uç cihaz zafiyetleri **kapsam dışı**.

## 5) Güvenlik İlkeleri

* **Veri minimizasyonu**: Trafik/içerik logu yok. Sadece toplam bayt, son bağlantı zamanı.
* **mTLS**: Control Plane ↔ Agent.
* **JWT**: Dış REST çağrıları. Kısa TTL, refresh ile yenileme.
* **Least privilege**: Node’larda yalnızca gerekli yetkiler.
* **Secrets**: SOPS/age veya Vault, üretimde KMS.
* **İmzalar**: Container imaj imzalama (cosign) v1.1.
* **SBOM**: syft, tarama grype; sonuçlar CI artefakt.
* **Ratelimit & Bot koruma**: hCaptcha, IP/UID bazlı.

## 6) Kriptografi ve Anahtar Yönetimi

* **WireGuard**: Curve25519, ChaCha20‑Poly1305 (kernel).
* **mTLS**: TLS 1.3, ECDSA/P‑256 sertifikalar.
* **Anahtar depolama**: Sunucu WG private key node üzerinde 0600 izin.
* **Rotasyon**: Node WG anahtarı 90 günde bir kademeli. mTLS sertifikaları 180 gün.
* **Config Paylaşımı**: Tek kullanımlık, 24 saat TTL’li imzalı URL.

## 7) Altyapı Sertleştirme

* Cloud firewall: yalnızca UDP WG portu ve gerekli yönetim portları.
* Kernel ayarları: `net.ipv4.ip_forward=1`, `rp_filter=2`, `net.ipv6.conf.all.disable_ipv6=0`.
* Kill‑switch kuralları: tünel düşerse trafik kapalı.
* SSH: key‑only, `PermitRootLogin no`, `PasswordAuthentication no`.
* Disk şifreleme: prod veri diskleri şifreli.
* Zaman senkronizasyonu: `chrony`.

## 8) Uygulama Güvenliği

* Giriş doğrulama: e‑posta doğrulama, TOTP (v1.1).
* Şifre saklama: Argon2id.
* CSRF/CORS: Sıkı origin, SameSite=Lax+Secure.
* Ratelimit: auth, checkout, peer CRUD.
* Girdi doğrulama: DTO/Schema validation.
* Kod kalitesi: golangci‑lint, eslint.
* Gizli taraması: gitleaks.
* Bağımlılık güncellemeleri: haftalık tarama, `renovate` önerilir.

## 9) Günlükleme ve Telemetri

* Uygulama hataları ve altyapı olayları loglanır.
* **Trafik içeriği ve hedefleri loglanmaz.**
* PII maskesi: e‑posta ve IP kısmi maskeleme loglarda.
* Metrikler: CPU/RAM, throughput, aktif peer, handshake hata oranı, p50/p95 latency.

## 10) KVKK/GDPR

* Aydınlatma ve rıza: `docs/privacy.md`.
* Haklar: erişim, düzeltme, silme, veri taşınabilirliği.
* Saklama: asgari, süre/amaç sınırlı.
* Veri işleyen/sorumlu roller tanımlı.
* Talep kanalı: **[privacy@yourdomain.example](mailto:privacy@yourdomain.example)**.

## 11) Olay Müdahalesi (IR)

* **Tespit**: Alertmanager, Sentry/Loki uyarıları.
* **Sınıflandırma**: Seviyeler A/B/C.

| Seviye | Tanım                        | Örnek                                    |
| ------ | ---------------------------- | ---------------------------------------- |
| A      | Aktif sömürü, yüksek etki    | Yetkisiz erişim, gizli anahtar sızıntısı |
| B      | Potansiyel sömürü, orta etki | RCE PoC, auth bypass ihtimali            |
| C      | Düşük etki                   | Bilgi sızıntısı, yanlış konfig           |

* **SLA**

  * A: 24 saat içinde geçici önlem, 72 saat içinde kalıcı yama.
  * B: 3 iş günü PoC doğrulama, 7 gün yama.
  * C: 14 gün yama.

* **Akış**: Tespit → IR lideri → kapsam/etki → containment → eradication → recovery → postmortem.

* **İletişim**: Kullanıcı bildirimi (gerekirse), otoritelerle koordinasyon.

* **Postmortem**: `docs/runbooks/postmortems/INC-YYYYMMDD.md` şablonu.

## 12) Dağıtım Güvenliği

* CI: imaj oluşturma, SBOM, imza; policy kontrol.
* CD: blue/green veya canary.
* Rollback: son başarılı imaj, veritabanı için backward‑compatible migrasyonlar.

## 13) Erişim Kontrolleri

* Admin panel IP allowlist.
* Prod erişimi: bastion, MFA, just‑in‑time.
* Ayrı hesaplar: dev/staging/prod ayrık.
* Least privilege IAM roller.

## 14) Yedekleme ve Kurtarma

* Postgres: günlü yedekler, günlük geri alma (PITR).
* Redis: önemli değilse ephemeral; gerekiyorsa snapshot.
* Yedek şifreleme, erişim kontrolü ve kurtarma testleri 90 günde bir.

## 15) Hukuki Talepler ve Veri İstekleri

* Talep doğrulama ve kapsam netleştirme.
* Yalnızca zorunlu meta veriler var; trafik içerikleri yok.
* Hukuk ekibi onayı olmadan veri paylaşımı yapılmaz.

## 16) Raporlama ve Şeffaflık

* Güvenlik sürüm notları: `RELEASE_NOTES.md`.
* Açıklanan zafiyetler ve CVE/CVSS bilgisi varsa eklenir.
* Durum sayfası: planlı bakım ve olaylar.

## 17) Yol Haritası (Güvenlik)

* v1.0: mTLS, ratelimit, SOPS, temel SBOM, gizli taraması.
* v1.1: imzalı imajlar (cosign), TOTP, WAF/DoS profilleri.
* v1.2: Obfuscation gateway, eBPF tabanlı trafik korumaları.

## 18) Hızlı Kontrol Listeleri

* **Yeni node**: patch level güncel → WireGuard → agent → mTLS → Prometheus target → leak test.
* **Yeni sürüm**: testler → SBOM → imza → staging canary → prod rollout → izleme.
* **Gizli rotasyonu**: KMS/sops anahtarları → mTLS cert → WG anahtarı → erişim tokenları.

---

### Örnek `security.txt`

```
Contact: mailto:security@yourdomain.example
Encryption: https://yourdomain.example/pgp/pubkey.asc
Preferred-Languages: tr, en
Policy: https://yourdomain.example/SECURITY
Acknowledgments: https://yourdomain.example/security/hall-of-fame
```

> Sorular için: **[security@yourdomain.example](mailto:security@yourdomain.example)**. Bu dosya değiştikçe `CHANGELOG.md` içinde not düşülür.

