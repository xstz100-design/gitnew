#!/bin/bash
set -e
echo "=== 1. 清理旧 dist ==="
sudo rm -rf /opt/wholesale/frontend/dist/*
sudo chown -R ubuntu:ubuntu /opt/wholesale/frontend/

echo "=== 2. 解压新 dist ==="
cd /opt/wholesale/frontend/
tar xzf /tmp/dist.tar.gz
echo "dist 文件数: $(find dist -type f | wc -l)"

echo "=== 3. 重启后端 ==="
sudo systemctl restart wholesale
sleep 3
curl -s http://127.0.0.1:8000/ && echo ""

echo "=== 4. 验证 HTTPS ==="
curl -sk https://khmerai.cn/ -o /dev/null -w "HTTPS status: %{http_code}\n"
curl -sk https://khmerai.cn/api/auth/telegram-auth -X POST -H 'Content-Type: application/json' -d '{"init_data":"test"}' -o /dev/null -w "TG auth endpoint: %{http_code}\n"

echo "=== DONE ==="
