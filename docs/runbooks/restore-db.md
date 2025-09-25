# Runbook: Postgres Veritabanı Kurtarma

## Durum
TriDot VPN backend Postgres veritabanı bozuldu veya veri kaybı yaşandı. Amaç, son tutarlı yedeği staging → prod prosedürüne uygun şekilde geri yüklemek.

## Ön Koşullar
- Son tam yedek (pg_dump veya WAL archive) lokasyonu biliniyor.
- AWS RDS/Open-source Postgres fark etmeksizin admin erişimi mevcut.
- Uygulama katmanı maintenance moduna alındı (read-only).

## Adımlar
1. **Incident Bildirimi**: #incident-kanalında olayı duyurun, status page güncelleyin.
2. **Yedek Tespiti**: `backup/YYYYMMDD-hhmm.sql.gz` dosyasını seçin.
3. **Read-Only Mod**: kube deployment’ı scale 0 yapmadan önce backend’i maintenance banner’a alın.
4. **Veritabanını Drop/Create**:
   ```bash
   psql -h postgres -U postgres -c "DROP DATABASE vpn_prod;"
   psql -h postgres -U postgres -c "CREATE DATABASE vpn_prod;"
   ```
5. **Restore**:
   ```bash
   gunzip -c backup.sql.gz | psql -h postgres -U vpn_user vpn_prod
   ```
6. **Migration Kontrolü**:
   ```bash
   goose -dir backend/internal/storage/postgres/migrations postgres "$POSTGRES_DSN" status
   goose -dir backend/internal/storage/postgres/migrations postgres "$POSTGRES_DSN" up
   ```
7. **Verification**:
   - `SELECT COUNT(*) FROM users;`
   - Backend `/health/ready` endpoint’i 200 dönüyor mu?
8. **Uygulama Yeniden Başlatma**.
9. **Incident Kapanışı**: Postmortem ve kullanıcı bilgilendirmesi.

## Rollback
- Restore başarısızsa veritabanı snapshot’ına (RDS snapshot vb.) dönün.

## Notlar
- Şifreler Vault/SOPS üzerinden sağlanmalı.
- Test restore işlemini staging’de düzenli olarak tatbik edin.
