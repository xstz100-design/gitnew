# 批发管理系统 Demo 部署教程

本项目是一套基于 **Vue 3 + Go** 的 B2B 批发管理系统，支持 Telegram Mini App 免登录、多语言（中/英/高棉）、移动端下单、管理后台、配送费估算等功能。

本教程指导你在一台全新 Ubuntu 服务器上部署一套 **独立的 Demo 环境**，与生产环境完全隔离。

---

## 一、系统架构

```
┌──────────────────────────────────────────┐
│            Nginx (80/443)                │
│   ┌────────────┐   ┌──────────────────┐  │
│   │ 前端静态文件│   │  /api/* → :8000  │  │
│   │  Vue 3 SPA │   │  Go 后端服务     │  │
│   └────────────┘   └──────────────────┘  │
└──────────────────────────────────────────┘
         ↓
   SQLite 数据库（单文件，零配置）
```

**技术栈**

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3 + Vite + Element Plus + Vant 4 + vue-i18n |
| 后端 | Go 1.24 + Gin + GORM + SQLite |
| 部署 | Nginx + systemd |
| 登录 | Telegram Bot（可选）+ 账号密码 |

---

## 二、服务器要求

- **OS**：Ubuntu 20.04 / 22.04 / 24.04（64位）
- **配置**：1 核 1GB 内存以上即可（Demo 用途）
- **开放端口**：80、443（如需 HTTPS）、22（SSH）
- **域名**：可选，有域名才能申请 SSL 证书

---

## 三、第一步：准备服务器

```bash
# 以 root 或 sudo 用户登录服务器
ssh ubuntu@<你的服务器IP>

# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装基础工具
sudo apt install -y nginx git curl wget unzip
```

---

## 四、第二步：准备前端代码（本地操作）

### 4.1 修改品牌信息

在 `frontend/src/` 中搜索并替换以下内容：

| 文件 | 需修改内容 | 替换为 |
|------|-----------|--------|
| `index.html` | `<title>` 标签内容 | 你的 Demo 名称 |
| `public/images/` | logo 图片 | 替换为你的 logo 或删除 |
| `i18n/zh.js` / `en.js` / `kh.js` | 品牌名称字符串 | 你的 Demo 名称 |

### 4.2 修改 API 地址

编辑 `frontend/src/utils/request.js`，确认 `baseURL` 使用相对路径（已是默认值，无需改动）：

```js
// 相对路径，由 Nginx 反代到后端，无需硬编码服务器 IP
baseURL: '/api'
```

### 4.3 构建前端

```bash
cd frontend
npm install
npm run build
# 产物在 frontend/dist/
```

---

## 五、第三步：准备后端（本地 Windows 交叉编译）

```powershell
cd backend-go

# 交叉编译为 Linux amd64 可执行文件
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o ..\wholesale-demo .

# 编译完成后在项目根目录生成 wholesale-demo 文件
```

---

## 六、第四步：准备配置文件

在 `backend-go/` 目录创建一个 Demo 专用的 `.env` 文件（**不要复制生产环境的 .env**）：

```bash
# demo.env — 新服务器使用此文件，命名为 .env 放到后端目录
DATABASE_URL=./data.db

# JWT 密钥：自己随机生成一串字符串
SECRET_KEY=your-random-secret-key-change-this

# 允许的前端来源（填你的 Demo 域名或 IP）
ALLOWED_ORIGINS=http://<你的IP>,https://<你的域名>

# Telegram Bot（Demo 可以不配，则只能账号密码登录）
TG_BOT_TOKEN=
TG_BOT_USERNAME=

# 站点 URL（用于 Telegram 回调）
SITE_URL=http://<你的IP>

# 货币汇率（演示用）
USD_TO_KHR_RATE=4100

# Token 有效期（分钟）
ACCESS_TOKEN_EXPIRE_MINUTES=10080
```

> **安全提示**：`SECRET_KEY` 必须修改，不要使用任何与生产环境相同的密钥。

---

## 七、第五步：上传文件到服务器

```bash
# 在本地执行：上传后端二进制
scp wholesale-demo ubuntu@<服务器IP>:/tmp/wholesale

# 上传前端构建产物（整个 dist 目录）
scp -r frontend/dist ubuntu@<服务器IP>:/tmp/frontend-dist

# 上传 .env 配置
scp demo.env ubuntu@<服务器IP>:/tmp/demo.env
```

或者使用 Python paramiko（见本项目 `deploy_backend.py` 和 `deploy_frontend.py`）。

---

## 八、第六步：服务器端部署

### 8.1 创建目录结构

```bash
sudo mkdir -p /opt/wholesale-demo/backend
sudo mkdir -p /opt/wholesale-demo/frontend/dist
sudo mkdir -p /opt/wholesale-demo/backend/uploads
sudo chown -R ubuntu:ubuntu /opt/wholesale-demo
```

### 8.2 部署后端

```bash
# 复制二进制和配置
cp /tmp/wholesale /opt/wholesale-demo/backend/wholesale
cp /tmp/demo.env /opt/wholesale-demo/backend/.env
chmod +x /opt/wholesale-demo/backend/wholesale

# 初始化数据库（首次部署执行一次）
cd /opt/wholesale-demo/backend
./wholesale --init-db   # 或者直接启动，程序会自动建表
```

### 8.3 部署前端

```bash
cp -r /tmp/frontend-dist/* /opt/wholesale-demo/frontend/dist/
```

### 8.4 配置 systemd 服务

创建 `/etc/systemd/system/wholesale-demo.service`：

```ini
[Unit]
Description=Wholesale Demo Backend (Go)
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/wholesale-demo/backend
ExecStart=/opt/wholesale-demo/backend/wholesale
Restart=always
RestartSec=5
Environment=PATH=/usr/bin:/bin
Environment=GIN_MODE=release
Environment=PORT=8001

[Install]
WantedBy=multi-user.target
```

> 注意：Demo 使用 **8001 端口**，避免与生产环境的 8000 冲突（如果在同一台机器上）。

```bash
# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable wholesale-demo
sudo systemctl start wholesale-demo

# 检查状态
sudo systemctl status wholesale-demo
```

### 8.5 配置 Nginx

创建 `/etc/nginx/sites-available/wholesale-demo`：

```nginx
server {
    listen 80;
    server_name <你的Demo域名或IP>;

    # 前端静态文件
    root /opt/wholesale-demo/frontend/dist;
    index index.html;

    # Vue Router history 模式
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 反向代理到后端
    location /api/ {
        proxy_pass http://127.0.0.1:8001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300;
    }

    # 用户上传的图片
    location /uploads/ {
        alias /opt/wholesale-demo/backend/uploads/;
        expires 30d;
    }

    # 前端静态资源长缓存
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    client_max_body_size 20M;
}
```

```bash
# 启用站点
sudo ln -s /etc/nginx/sites-available/wholesale-demo /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## 九、第七步：配置 HTTPS（可选但推荐）

```bash
# 安装 certbot
sudo apt install -y certbot python3-certbot-nginx

# 申请证书（需要域名已解析到此服务器）
sudo certbot --nginx -d <你的Demo域名>

# certbot 会自动修改 nginx 配置并加上 SSL
# 证书 90 天自动续期：
sudo systemctl enable certbot.timer
```

---

## 十、第八步：创建管理员账号

后端首次启动后，通过 API 创建超级管理员：

```bash
# 在服务器上执行
curl -X POST http://localhost:8001/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "Demo@123456",
    "role": "super_admin"
  }'
```

或者直接操作数据库：

```bash
cd /opt/wholesale-demo/backend

# 安装 sqlite3 工具
sudo apt install -y sqlite3

# 查看用户表
sqlite3 data.db "SELECT id, username, role FROM users;"

# 将某个用户提升为 super_admin
sqlite3 data.db "UPDATE users SET role='super_admin', is_approved=1 WHERE username='admin';"
```

---

## 十一、功能说明

### 账号体系

| 角色 | 说明 | 访问路径 |
|------|------|---------|
| `super_admin` | 超级管理员，全功能 | `/admin` |
| `admin` | 普通管理员 | `/admin` |
| `merchant` | 商户（PC端） | `/merchant` |
| 移动端用户 | Telegram 或手机号注册 | `/m` |

### 主要功能模块

- **商品管理**：分类、SKU、图片、库存预警
- **订单管理**：下单、审核、配送状态跟踪
- **商户管理**：审核注册、授信额度
- **配送费**：基于 Google Maps 距离估算
- **月结账单**：自动生成 PDF 账单
- **公告系统**：管理员发布通知
- **系统设置**：联系方式、配送参数、Telegram 群组

### 登录方式

1. **账号密码**：直接访问 `/login`，输入用户名密码
2. **Telegram Mini App**：通过 Bot 启动（需配置 `TG_BOT_TOKEN`）
3. **Telegram 扫码**：网页端扫码（需配置 Bot）

---

## 十二、常用运维命令

```bash
# 查看后端日志
sudo journalctl -u wholesale-demo -f

# 重启后端
sudo systemctl restart wholesale-demo

# 重载 Nginx（更新前端后执行）
sudo systemctl reload nginx

# 查看端口占用
ss -tlnp | grep 8001

# 备份数据库
cp /opt/wholesale-demo/backend/data.db ~/data.db.bak
```

---

## 十三、更新 Demo

### 更新前端

```bash
# 本地重新构建
cd frontend && npm run build

# 上传并替换
scp -r dist/* ubuntu@<服务器IP>:/opt/wholesale-demo/frontend/dist/
ssh ubuntu@<服务器IP> "sudo systemctl reload nginx"
```

### 更新后端

```bash
# 本地重新编译
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o wholesale-demo backend-go/

# 上传、替换、重启
scp wholesale-demo ubuntu@<服务器IP>:/tmp/wholesale
ssh ubuntu@<服务器IP> "sudo cp /tmp/wholesale /opt/wholesale-demo/backend/wholesale && sudo chmod +x /opt/wholesale-demo/backend/wholesale && sudo systemctl restart wholesale-demo"
```

---

## 十四、注意事项

1. **不要使用生产数据库**：Demo 环境使用独立的 `data.db`，禁止挂载生产数据库
2. **密钥隔离**：`SECRET_KEY`、`TG_BOT_TOKEN` 必须与生产环境不同
3. **端口隔离**：Demo 后端用 `8001`，Nginx 虚拟主机用不同域名/IP
4. **定期清理**：Demo 数据库建议定期重置，避免演示数据污染
5. **访问控制**：可在 Nginx 中加 `allow/deny` 或 Basic Auth 限制访问范围

重置 Demo 数据：

```bash
# 停服务 → 删库 → 重启（自动重建空库）
sudo systemctl stop wholesale-demo
rm /opt/wholesale-demo/backend/data.db
sudo systemctl start wholesale-demo
```
