# Sızıntı ve Yük Testleri Runbook

Bu runbook, TriDot VPN node’larının DNS/IPv6 sızıntı kontrolleri ve throughput ölçümleri için adımları içerir.

## 1. DNS/IPv6 Sızıntı Testi

1. VPN’e bağlanın (WireGuard config ile).
2. `scripts/leaktest.sh` komutunu çalıştırın:
   ```bash
   ./scripts/leaktest.sh
   ```
3. Çıktı dosyaları `/tmp/leaktest.*` altında. `dig` sonucu ISP’nin DNS’i yerine beklenen resolver’dan dönmeli.
4. IPv6 cevabı varsa TriDot node’un IPv6 adresini doğrulayın; yoksa firewall kuralını kontrol edin.

## 2. Throughput (iperf3)

1. Node üzerinde `iperf3 -s` server’ı başlatın.
2. Müşteri tarafında:
   ```bash
   ./scripts/loadtest.sh vpn-node.internal 60
   ```
3. Hedef: `>= 2 Gbps aggregate` değerine ulaşmak.

## 3. Raporlama

- Grafana panosuna tester sonuçlarını not edin.
- Başarısızlık durumunda firewall/MTU loglarını inceleyin.
