# SSH Hardening, Bastion ve JIT Erişim Runbook

Bu runbook, TriDot VPN altyapısındaki sunucuların SSH üzerinden güvenli şekilde yönetilmesi için önerilen kontrolleri listeler. Amaç, erişimi bastion host üzerinden sınırlandırmak, anahtar yönetimini sıkılaştırmak ve Just-In-Time (JIT) onay süreci ile operasyonel güvenliği artırmaktır.

## 1. Mimarinin Özeti

1. **Bastion Host**: Tüm SSH erişimi tek bir bastion üzerinden geçer. Bu makinede kayıt altına alma (session logging) ve PAM entegrasyonu yapılır.
2. **Private Subnet**: VPN node’ları ve kontrol düzlemi sunucuları yalnızca bastion’un bulunduğu CIDR’den SSH kabul eder.
3. **JIT Portal**: Vault veya Okta Workflows tabanlı bir otomasyon, erişim isteğini onayladıktan sonra bastion üzerindeki `authorized_keys` dosyasına kısa ömürlü anahtar yazar.

## 2. Bastion Kurulumu

- **OS Harden**: Son güncellemeleri yükleyin, gereksiz servisleri kapatın (`iptables`, `fail2ban` önerilir).
- **SSH Config** (`/etc/ssh/sshd_config`):
  ```
  Port 22
  Protocol 2
  PermitRootLogin no
  PasswordAuthentication no
  PubkeyAuthentication yes
  AllowAgentForwarding no
  ClientAliveInterval 300
  ClientAliveCountMax 2
  AllowTcpForwarding no
  ````
- **Audit Logs**: `Auditd` veya `tlog + SSM` ile session logging etkinleştirin.
- **MFA**: Google Authenticator veya hardware token kullanarak bastion girişine PAM tabanlı MFA ekleyin.

## 3. VPN Node’larında SSH Hardening

1. `sshd_config`:
   ```
   PermitRootLogin no
   PasswordAuthentication no
   AllowTcpForwarding no
   X11Forwarding no
   ClientAliveInterval 60
   ClientAliveCountMax 2
   AllowUsers bastion-sync
   ```
2. `/etc/hosts.allow`:
   ```
   sshd: 198.51.100.10  # bastion public IP
   sshd: 10.0.0.0/24    # bastion private subnet
   ```
3. `/etc/hosts.deny`:
   ```
   sshd: ALL
   ```
4. Fail2ban jail’leri, audit log rotasyonu ve unattended-upgrades aktif edilir.

## 4. JIT Erişim Akışı

1. Operatör Slack’te `/jit-access staging-node-1 2h` komutu çalıştırır.
2. Workflow (ör. Cloud Function) Vault’ta ops yetkililerinden onay ister.
3. Onay sonrası bastion’da aşağıdaki adımlar otomatik yapılır:
   ```bash
   cat /tmp/temp-key.pub >> /home/bastion-sync/.ssh/authorized_keys
   at now + 2 hours <<'TASK'
     sed -i '/temp-key-comment/d' /home/bastion-sync/.ssh/authorized_keys
   TASK
   ```
4. Süre dolduğunda anahtar authorized listeden çıkarılır, bastion logları arşivlenir.

## 5. Terraform/Ansible Notları

- Terraform security group örneği:
  ```hcl
  resource "aws_security_group_rule" "bastion_ssh" {
    type              = "ingress"
    from_port         = 22
    to_port           = 22
    protocol          = "tcp"
    cidr_blocks       = [var.corporate_cidr]
  }
  resource "aws_security_group_rule" "nodes_ssh" {
    type              = "ingress"
    from_port         = 22
    to_port           = 22
    protocol          = "tcp"
    security_group_id = aws_security_group.nodes.id
    source_security_group_id = aws_security_group.bastion.id
  }
  ```
- Ansible: `ansible/roles/ssh-hardening` rolü ile `sshd_config`, fail2ban ve auditd ayarlarını template olarak dağıtın.

## 6. İzleme ve Uyarılar

- Prometheus alert: `node_ssh_login_failures_total` metriğinde spike olduğunda uyarı üretin.
- Loki sorgusu: `{app="sshd", level="warning"}` filtreleri ile unauthorized girişleri görün.
- Bastion logları günlük olarak S3/Blob Storage’a arşivlenir, `retention` ≥ 180 gün.

## 7. Check-list

- [ ] Bastion host IAM rolleri ve MFA doğrulandı.
- [ ] Tüm prod node’lar bastion IP’si dışındaki SSH girişlerini reddediyor.
- [ ] JIT workflow test edildi (anahtar otomatik siliniyor).
- [ ] Prometheus & Loki panolarında `ssh-failures` grafiği ekli.
- [ ] Dokümantasyon güncellendi (`docs/runbooks/`).

## 8. Referanslar

- CIS Benchmark: Ubuntu 22.04 LTS – SSH bölümü.
- AWS SSM Session Manager (parolasız erişim) dokümanı.
- HashiCorp Vault – SSH Secrets Engine.
