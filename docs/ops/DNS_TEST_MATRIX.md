# DNS Sızıntı Test Matrisi

## Hedefler
- Laboratuvar ortamı: %100 başarı (tüm platformlarda DNS sızıntısı yok)
- Saha testi: %98 başarı (maksimum %2 sızıntı toleransı)

## Platformlar
| Platform | Test Aracı | Not |
|----------|------------|-----|
| macOS    | `scripts/leaktest.sh` + browserleaks.com | IPv6 aktif |
| Windows  | `dnsleaktest.com` + `nslookup` | WireGuard resmi client |
| iOS      | Browser-based test + Cloudflare Warp check | LTE/Wi-Fi |
| Android  | `ipleak.net` + `dig` (Termux) | |

## Raporlama
- Spreadsheet: her test için `pass/fail`, IP provider, notlar.
- Haftalık QA: % başarı oranı hesaplanır.

## Otomasyon
- Plan: BrowserStack veya Sauce Labs mobil testleri ile otomatik tarama.
