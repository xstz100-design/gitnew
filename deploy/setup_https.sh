#!/bin/bash
set -e

# ===================================================
# 免费 HTTPS 部署脚本 (Let's Encrypt + certbot)
# 域名: khmerai.cn
# 适用: Ubuntu / Debian 纯命令行服务器
# ===================================================

DOMAIN="khmerai.cn"
EMAIL="admin@${DOMAIN}"
WEBROOT="/var/www/certbot"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ---- 检查 root ----
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}请使用 root 运行此脚本${NC}"
    exit 1
fi

echo "=============================================="
echo "  Let's Encrypt 免费 HTTPS 配置"
echo "  域名: ${DOMAIN}"
echo "=============================================="

# ===== 第一步: 确保域名 DNS 已解析到本机 =====
echo ""
echo -e "${GREEN}[1/5] 检查 DNS 解析...${NC}"
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || curl -s icanhazip.com 2>/dev/null)
DNS_IP=$(dig +short ${DOMAIN} 2>/dev/null | tail -1)

if [ -z "$DNS_IP" ]; then
    echo -e "${RED}无法解析 ${DOMAIN} 的 DNS!${NC}"
    echo "请先在域名服务商添加 A 记录:"
    echo "  ${DOMAIN} → ${SERVER_IP}"
    echo ""
    echo "DNS 生效后再运行此脚本"
    exit 1
fi

if [ "$DNS_IP" != "$SERVER_IP" ]; then
    echo -e "${YELLOW}警告: DNS 解析 IP (${DNS_IP}) 与本机 IP (${SERVER_IP}) 不一致${NC}"
    echo "如果你确认这是正确的(如使用 CDN)，请按 Enter 继续，否则 Ctrl+C 退出"
    read -r
fi
echo "DNS 解析正确: ${DOMAIN} → ${DNS_IP} ✓"

# ===== 第二步: 安装 certbot =====
echo ""
echo -e "${GREEN}[2/5] 安装 certbot...${NC}"

if command -v certbot &>/dev/null; then
    echo "certbot 已安装 ✓"
else
    apt-get update -y
    apt-get install -y certbot
    echo "certbot 安装完成 ✓"
fi

# ===== 第三步: 先用 HTTP-only nginx 配置，确保 ACME 验证路径可用 =====
echo ""
echo -e "${GREEN}[3/5] 配置 Nginx (HTTP 临时模式)...${NC}"

mkdir -p ${WEBROOT}

# 写一个临时的 HTTP-only 配置用于 ACME 验证
cat > /etc/nginx/sites-available/wholesale <<'TMPCONF'
server {
    listen 80;
    server_name khmerai.cn;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        root /opt/wholesale/frontend/dist;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
TMPCONF

ln -sf /etc/nginx/sites-available/wholesale /etc/nginx/sites-enabled/wholesale
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx

# ===== 第四步: 申请证书 =====
echo ""
echo -e "${GREEN}[4/5] 申请 Let's Encrypt 证书...${NC}"

certbot certonly \
    --webroot \
    --webroot-path ${WEBROOT} \
    -d ${DOMAIN} \
    --email ${EMAIL} \
    --agree-tos \
    --non-interactive

if [ ! -f "/etc/letsencrypt/live/${DOMAIN}/fullchain.pem" ]; then
    echo -e "${RED}证书申请失败! 请检查:${NC}"
    echo "  1. 域名 ${DOMAIN} 是否已解析到本服务器"
    echo "  2. 80 端口是否开放"
    echo "  3. 查看日志: /var/log/letsencrypt/letsencrypt.log"
    exit 1
fi
echo "证书申请成功 ✓"

# ===== 第五步: 切换到完整 HTTPS Nginx 配置 =====
echo ""
echo -e "${GREEN}[5/5] 配置 Nginx HTTPS...${NC}"

cp /opt/wholesale/deploy/nginx.conf /etc/nginx/sites-available/wholesale
nginx -t && systemctl reload nginx

echo "HTTPS Nginx 配置完成 ✓"

# ===== 设置自动续期 =====
echo ""
echo -e "${GREEN}设置自动续期 (crontab)...${NC}"
# certbot renew 会自动检测，且自带随机延迟
(crontab -l 2>/dev/null | grep -v certbot; echo "0 3 * * * certbot renew --quiet --deploy-hook 'systemctl reload nginx'") | crontab -
echo "自动续期已配置 (每天凌晨 3 点检查) ✓"

echo ""
echo "=============================================="
echo -e "${GREEN}  HTTPS 配置完成!${NC}"
echo "=============================================="
echo ""
echo "  访问: https://${DOMAIN}"
echo "  API:  https://${DOMAIN}/api"
echo ""
echo "  证书路径:"
echo "    fullchain: /etc/letsencrypt/live/${DOMAIN}/fullchain.pem"
echo "    privkey:   /etc/letsencrypt/live/${DOMAIN}/privkey.pem"
echo ""
echo "  证书有效期 90 天，已配置自动续期"
echo "  手动续期: certbot renew"
echo "  查看证书: certbot certificates"
echo ""
