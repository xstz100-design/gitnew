# 东方优选 TONGFANG YOUXUAN — 项目完整技术文档

---

## 一、项目概述

**项目名称**：东方优选（TONGFANG YOUXUAN）  
**业务定位**：柬埔寨批发管理系统 — 面向中国在柬商人的 B2B 批发下单平台  
**核心功能**：商品管理、商户下单、配送调度、月结账单、库存预警、Telegram Mini App 免登录  
**线上地址**：https://khmerai.cn  
**Telegram Bot**：[@TONGFANGyouxuan_bot](https://t.me/TONGFANGyouxuan_bot)

---

## 二、技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| **前端框架** | Vue 3 (Composition API) | ^3.4.15 |
| **构建工具** | Vite | ^5.0.12 |
| **路由** | Vue Router (History 模式) | ^4.2.5 |
| **状态管理** | Pinia + pinia-plugin-persistedstate | ^2.1.7 |
| **PC端UI** | Element Plus | ^2.5.6 |
| **移动端UI** | Vant 4 | ^4.8.11 |
| **国际化** | vue-i18n | ^11.2.8 |
| **HTTP** | Axios | ^1.6.7 |
| **样式** | SCSS (设计系统变量) | sass ^1.97.3 |
| **后端框架** | FastAPI | 0.109.2 |
| **ORM** | SQLModel (SQLAlchemy 封装) | 0.0.16 |
| **数据库** | SQLite (WAL 模式) | — |
| **认证** | JWT (python-jose) + bcrypt (passlib) | 3.3.0 / 1.7.4 |
| **Web服务器** | Nginx (反向代理+静态资源) | 1.24.0 |
| **应用服务器** | Uvicorn | 0.27.1 |
| **SSL** | Let's Encrypt (certbot 自动续期) | — |
| **服务管理** | systemd (wholesale.service) | — |
| **Telegram SDK** | telegram-web-app.js | — |
| **图片处理** | Pillow | 10.2.0 |

---

## 三、服务器部署架构

```
用户(Telegram/浏览器)
         │
         ▼
┌─────────────────────────────────────────────────┐
│  Nginx (443 SSL / 80→301)                       │
│  Ubuntu, Let's Encrypt                          │
│  khmerai.cn / 43.134.13.229                     │
├─────────────────────────────────────────────────┤
│  /           → dist/index.html (Vue SPA)        │
│  /api/*      → proxy http://127.0.0.1:8000      │
│  /uploads/*  → alias backend/uploads/            │
└─────────────────────────────────────────────────┘
         │ proxy_pass
         ▼
┌─────────────────────────────────────────────────┐
│  Uvicorn :8000                                  │
│  FastAPI 应用                                    │
│  SQLite cambodia_wholesale.db                   │
└─────────────────────────────────────────────────┘
```

**服务器**：Ubuntu on `43.134.13.229`，SSH 用户 `ubuntu`  
**部署路径**：`/opt/wholesale/`  
**虚拟环境**：`/opt/wholesale/venv/`  
**数据库文件**：`/opt/wholesale/backend/cambodia_wholesale.db`  
**前端文件**：`/opt/wholesale/dist/`（Vite 构建输出）  
**图片上传**：`/opt/wholesale/backend/uploads/`  
**SSL 证书**：`/etc/letsencrypt/live/khmerai.cn/`（到期 2026-07-14，自动续期）

---

## 四、目录结构 — 完整说明

```
e:\Program_ayang\vue\
│
├── backend/                     # FastAPI 后端
│   ├── main.py                  # 应用入口：FastAPI 实例、中间件、路由注册
│   ├── init_db.py               # 数据库初始化脚本（建表+创建管理员+演示数据）
│   ├── requirements.txt         # Python 依赖
│   ├── cambodia_wholesale.db    # SQLite 数据库文件（生产环境在服务器）
│   ├── uploads/                 # 图片上传目录
│   │   └── thumbnails/          # 缩略图
│   ├── app/
│   │   ├── __init__.py
│   │   ├── models/              # 数据模型（6个模型文件）
│   │   │   ├── __init__.py      # 统一导出所有模型和枚举
│   │   │   ├── base.py          # User 模型 + Product 模型
│   │   │   ├── order.py         # Order 模型 + OrderItem 模型
│   │   │   ├── transaction.py   # Transaction 模型（资金流水）
│   │   │   ├── category.py      # Category 模型（商品分类）
│   │   │   ├── announcement.py  # Announcement 模型（公告）
│   │   │   └── monthly_bill.py  # MonthlyBill 模型（月结账单）
│   │   ├── core/                # 核心模块
│   │   │   ├── __init__.py
│   │   │   ├── config.py        # 配置中心（pydantic-settings，支持 .env 覆盖）
│   │   │   ├── database.py      # 数据库引擎+会话（WAL模式+连接池优化）
│   │   │   ├── security.py      # JWT 签发/验证 + bcrypt 密码哈希
│   │   │   ├── dependencies.py  # FastAPI 依赖注入（认证、角色检查器）
│   │   │   ├── telegram.py      # Telegram initData HMAC-SHA256 签名验证
│   │   │   └── rate_limit.py    # 速率限制（全局IP/登录防暴力/上传三层）
│   │   ├── api/                 # API 端点
│   │   │   ├── __init__.py
│   │   │   ├── auth.py          # 认证（登录/注册/TG免登/审核/用户管理）
│   │   │   ├── schemas.py       # Pydantic 请求/响应 Schema
│   │   │   ├── products.py      # 商品 CRUD
│   │   │   ├── orders.py        # 订单创建/查询/更新/取消
│   │   │   ├── categories.py    # 分类管理
│   │   │   ├── announcements.py # 公告管理
│   │   │   ├── billing.py       # 月结账单生成/管理
│   │   │   └── upload.py        # 图片上传（流式写入+自动压缩）
│   │   └── services/            # 业务服务
│   │       ├── __init__.py
│   │       ├── telegram.py      # Telegram 通知（新订单/库存预警推送管理员）
│   │       └── image.py         # 图片压缩优化服务（Pillow）
│   └── migrate_*.py             # 各类数据库迁移脚本（历史）
│
├── frontend/                    # Vue 3 前端
│   ├── index.html               # 入口HTML（含 Telegram WebApp SDK <script>）
│   ├── package.json             # npm 依赖
│   ├── vite.config.js           # Vite 配置（@别名 + 开发代理 /api → :8000）
│   ├── dist/                    # 构建输出（不入版本控制）
│   ├── dist.tar.gz              # 部署打包
│   ├── public/images/           # 公共静态图片
│   └── src/
│       ├── main.js              # 应用引导（Pinia/Router/i18n/Vant/ElementPlus/TG自动登录）
│       ├── App.vue              # 根组件（<router-view>）
│       ├── api/
│       │   └── index.js         # 所有 API 调用函数（40+ 个）
│       ├── router/
│       │   └── index.js         # 路由定义 + 路由守卫（PC/移动自动切换）
│       ├── stores/
│       │   ├── user.js          # 用户状态（token/userInfo/登录/TG登录/登出，持久化localStorage）
│       │   └── cart.js          # 购物车状态（增删改查/库存校验/持久化localStorage）
│       ├── utils/
│       │   ├── request.js       # Axios 封装（Bearer token 自动注入/401自动登出/错误提示）
│       │   ├── telegram.js      # Telegram Mini App 检测/initData获取/初始化
│       │   ├── device.js        # 设备检测（移动端/触屏/视口宽度判断）
│       │   ├── format.js        # 格式化工具（USD/KHR货币转换/日期时间）
│       │   └── colors.js        # 颜色工具
│       ├── i18n/
│       │   ├── index.js         # vue-i18n 配置（中/英文切换，持久化localStorage）
│       │   ├── zh.js            # 中文翻译文件
│       │   └── en.js            # 英文翻译文件
│       ├── styles/
│       │   ├── variables.scss   # SCSS 设计系统（颜色/间距/圆角/阴影变量）
│       │   ├── global.scss      # 全局样式
│       │   └── element-override.scss  # Element Plus 主题覆盖
│       ├── components/
│       │   ├── SkeletonProduct.vue   # 商品加载骨架屏
│       │   └── SkeletonTable.vue     # 表格加载骨架屏
│       ├── layouts/
│       │   ├── AdminLayout.vue       # 管理端布局（侧边栏+头部）
│       │   ├── MerchantLayout.vue    # PC端商户布局（侧边栏）
│       │   └── MobileLayout.vue      # 移动端商户布局（底部Tab栏）
│       └── views/
│           ├── Login.vue             # 登录页（账号密码表单）
│           ├── Register.vue          # 注册页（手机号自助注册）
│           ├── admin/                # 管理端 8 个页面
│           │   ├── Dashboard.vue     # 仪表盘（订单统计/库存预警/快捷操作）
│           │   ├── Products.vue      # 商品管理（CRUD+多图上传+库存编辑）
│           │   ├── Orders.vue        # 订单管理（状态变更/配送调度）
│           │   ├── Merchants.vue     # 商户管理（查看/编辑/启禁用/月结设置）
│           │   ├── Approvals.vue     # 用户审核（通过/拒绝+原因）
│           │   ├── Categories.vue    # 分类管理（排序/启禁用）
│           │   ├── Announcements.vue # 公告管理（滚动通知/联系客服/关于）
│           │   └── Profile.vue       # 管理员个人中心
│           ├── merchant/             # PC端商户 4 个页面
│           │   ├── Products.vue      # 浏览商品（表格视图+分类筛选）
│           │   ├── Cart.vue          # 购物车（加减/删除/提交订单）
│           │   ├── Orders.vue        # 我的订单（状态筛选/取消订单）
│           │   └── Profile.vue       # 个人信息（修改资料/密码）
│           └── mobile/               # 移动端商户 5 个页面（Vant UI）
│               ├── Shop.vue          # 商品列表主页（分类Tab+搜索+卡片展示）
│               ├── Shop_new.vue      # 新版商品列表（备用）
│               ├── Cart.vue          # 购物车
│               ├── Orders.vue        # 订单列表（状态Tab筛选）
│               └── Profile.vue       # 个人中心（信息/密码/语言切换/登出）
│
├── deploy/                      # 部署配置
│   ├── nginx.conf               # Nginx HTTPS 配置（HTTP→HTTPS重定向+反向代理）
│   ├── wholesale.service         # systemd 服务定义
│   ├── setup.sh                 # 服务器初始化脚本
│   ├── setup_https.sh           # Let's Encrypt 证书申请脚本
│   ├── bt_setup.sh              # 宝塔面板部署脚本
│   ├── bt_nginx.conf            # 宝塔 Nginx 配置
│   ├── backup_db.sh             # 数据库备份脚本
│   └── BT_DEPLOY.md            # 宝塔部署文档
│
├── deploy_ssh.py                # Python SSH 自动化部署脚本（paramiko）
├── deploy_rebuild.sh            # 完整重建部署 Shell 脚本
├── deploy_*.py                  # 其他部署辅助脚本
├── check_*.py                   # 服务器检查脚本
├── test_api*.py                 # API 测试脚本
└── vue.code-workspace           # VS Code 工作区配置
```

---

## 五、数据库模型 — 全部 7 张表

### 5.1 `users` — 用户表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `id` | INTEGER | PK, 自增 | 用户ID |
| `username` | VARCHAR(50) | UNIQUE, INDEX | 商户=手机号, 管理员=6位数字(100001起), TG用户=`tg_{telegram_id}` |
| `hashed_password` | VARCHAR | NOT NULL | bcrypt 哈希密码 |
| `full_name` | VARCHAR(100) | NOT NULL | 姓名/店铺名 |
| `role` | ENUM | `admin` / `merchant` | 用户角色 |
| `phone` | VARCHAR(20) | INDEX | 手机号 |
| `address` | VARCHAR(200) | — | 配送地址 |
| `credit_limit` | FLOAT | 默认 0.0 | 月结累计金额（随订单状态联动增减） |
| `billing_day` | INTEGER | 1-31 | 每月结账日 |
| `allow_monthly_billing` | BOOL | 默认 false | 是否允许赊账/月结 |
| `location_url` | VARCHAR(500) | — | 谷歌地图链接 |
| `store_photo` | VARCHAR(500) | — | 门面照片 URL |
| `telegram_id` | INTEGER | UNIQUE, INDEX | Telegram 用户 ID（用于免登录绑定） |
| `telegram_bot_token` | VARCHAR(200) | — | 商户自有 bot token（推送通知用） |
| `telegram_chat_id` | VARCHAR(100) | — | Telegram chat ID（推送通知用） |
| `approval_status` | ENUM | `pending`/`approved`/`rejected` | 审核状态 |
| `rejected_reason` | VARCHAR(200) | — | 拒绝原因 |
| `approved_at` | DATETIME | — | 审核通过时间（柬埔寨时区） |
| `must_change_password` | BOOL | 默认 true | 首次登录强制改密码 |
| `is_active` | BOOL | 默认 true | 是否启用（false=禁用） |
| `created_at` | DATETIME | UTC+7 | 创建时间（柬埔寨时区） |

**角色说明**：
- `admin`：超级管理员用户名 `100001`，拥有密码重置、用户删除等最高权限
- `merchant`：商户，可通过手机号注册、管理员创建、或 Telegram 自动创建

**关系**：
- `orders` → 一对多关联 Order（通过 merchant_id）
- `transactions` → 一对多关联 Transaction（通过 user_id）

### 5.2 `products` — 商品表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER PK | 自增 |
| `name` | VARCHAR(100) INDEX | 商品名（中文） |
| `name_kh` | VARCHAR(100) | 高棉语名称 |
| `unit` | VARCHAR(20) | 单位：件/箱/包，默认"件" |
| `specs` | VARCHAR(100) | 规格说明（如"24瓶/箱"） |
| `barcode` | VARCHAR(100) | 商品条码 |
| `price_usd` | FLOAT >0 | 批发价（美金），必填 |
| `retail_price_usd` | FLOAT | 建议零售价（美金） |
| `stock` | INTEGER ≥0 | 当前库存数量 |
| `stock_warning` | INTEGER | 库存预警阈值，默认 10 |
| `description` | VARCHAR(500) | 商品描述 |
| `image_url` | VARCHAR(500) | 主图 URL |
| `img1` ~ `img5` | VARCHAR(500) | 5 张附加图片 URL |
| `category` | VARCHAR(50) | 分类名（字符串关联 categories.name） |
| `sort_order` | INTEGER | 排序序号（越小越靠前） |
| `is_featured` | BOOL | 是否推荐置顶 |
| `is_active` | BOOL | 是否上架（false=软删除/下架） |
| `created_at` | DATETIME | 创建时间 |
| `updated_at` | DATETIME | 更新时间 |

**计算属性**：`is_low_stock` → `stock <= stock_warning`  
**关系**：`order_items` → 一对多关联 OrderItem

### 5.3 `orders` — 订单表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER PK | 自增 |
| `order_no` | VARCHAR(50) UNIQUE INDEX | 订单号，格式：`ORD20260415000001` |
| `merchant_id` | FK→users.id | 下单商户 |
| `total_usd` | FLOAT ≥0 | 订单总金额（美金） |
| `payment_status` | ENUM | `unpaid`=未支付 / `cash`=现结 / `monthly`=月结 |
| `delivery_status` | ENUM | `pending`=待派送 / `delivering`=送货中 / `delivered`=已签收 / `cancelled`=已取消 |
| `delivery_address` | VARCHAR(200) | 配送地址（默认取商户地址） |
| `delivery_phone` | VARCHAR(20) | 联系电话（默认取商户电话） |
| `delivery_person_id` | FK→users.id | 配送员ID |
| `note` | VARCHAR(500) | 订单备注 |
| `created_at` | DATETIME | 创建时间 |
| `updated_at` | DATETIME | 更新时间 |
| `delivered_at` | DATETIME | 签收时间 |

**订单号生成规则**：`ORD` + `YYYYMMDD`（当天日期） + `000001`（6位序号，每天重新计数）  
**计算属性**：`total_khr` → `total_usd × 4000`（美元转瑞尔）  
**关系**：`items` → 一对多关联 OrderItem；`merchant` → 多对一关联 User

### 5.4 `order_items` — 订单明细表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER PK | 自增 |
| `order_id` | FK→orders.id | 所属订单 |
| `product_id` | FK→products.id | 商品 |
| `quantity` | INTEGER >0 | 购买数量 |
| `unit_price_usd` | FLOAT >0 | 下单时单价快照（不受后续改价影响） |
| `subtotal_usd` | FLOAT ≥0 | 小计 = unit_price_usd × quantity |

### 5.5 `transactions` — 资金流水表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER PK | 自增 |
| `user_id` | FK→users.id | 关联用户 |
| `amount_usd` | FLOAT | 金额（美金） |
| `transaction_type` | ENUM | `payment`=收款 / `refund`=退款 / `expense`=支出 / `adjustment`=调整 |
| `order_id` | FK→orders.id | 关联订单（可选） |
| `description` | VARCHAR(200) | 描述 |
| `note` | VARCHAR(500) | 备注 |
| `created_at` | DATETIME | 创建时间 |
| `created_by` | FK→users.id | 操作人 |

**计算属性**：`amount_khr` → `amount_usd × 4000`

### 5.6 `categories` — 商品分类表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER PK | 自增 |
| `name` | VARCHAR(50) UNIQUE | 分类名 |
| `sort_order` | INTEGER | 排序序号 |
| `is_active` | BOOL | 是否启用 |

**初始数据**：饮料(1)、食品(2)、调料(3)、日用品(4)、零食(5)

### 5.7 `announcements` — 公告表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER PK | 自增 |
| `type` | ENUM | `notice`=滚动公告 / `contact`=联系客服 / `about`=关于系统 |
| `content_zh` | VARCHAR(2000) | 中文内容 |
| `content_en` | VARCHAR(2000) | 英文内容 |
| `is_active` | BOOL | 是否激活 |
| `sort_order` | INTEGER | 排序 |
| `created_at` / `updated_at` | DATETIME | 时间戳 |

### 5.8 `monthly_bills` — 月结账单表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER PK | 自增 |
| `merchant_id` | FK→users.id | 商户 |
| `year` | INTEGER | 账期年份 |
| `month` | INTEGER | 账期月份（1-12） |
| `total_amount` | FLOAT | 月结总金额 |
| `paid_amount` | FLOAT | 已付金额 |
| `status` | ENUM | `unpaid`=未结清 / `paid`=已结清 / `partial`=部分结清 |

---

## 六、API 端点 — 完整列表

### 6.1 认证模块 `/api/auth/*`

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| `POST` | `/telegram-auth` | 公开 | **Telegram Mini App 免登录**：验证 initData 签名 → 自动创建或查找用户 → 返回 JWT |
| `POST` | `/public-register` | 公开 | 用户自助注册（手机号+密码+姓名），状态=待审核 |
| `POST` | `/login` | 公开 | 用户名密码登录（OAuth2 表单格式），含防暴力破解 |
| `POST` | `/register` | admin | 管理员创建用户（商户用手机号作账号/管理员自动编号） |
| `GET` | `/me` | 已登录 | 获取当前用户信息 |
| `PATCH` | `/me` | 已登录 | 更新个人信息（姓名/电话/地址/门面照片/TG通知设置） |
| `GET` | `/users` | admin | 用户列表（可按角色筛选） |
| `PATCH` | `/users/{id}` | admin | 更新用户（含启用/禁用/月结设置） |
| `DELETE` | `/users/{id}` | 超管(100001) | 删除用户（有关联订单时拒绝，建议停用） |
| `POST` | `/users/{id}/reset-password` | 超管(100001) | 重置密码为 `123456` |
| `POST` | `/change-password` | 已登录 | 修改自己密码（需验证旧密码） |
| `GET` | `/pending-users` | admin | 待审核用户列表 |
| `GET` | `/all-registrations` | admin | 所有注册用户列表（可按审核状态筛选） |
| `POST` | `/users/{id}/approve` | admin | 审核用户（通过/拒绝+拒绝原因） |
| `GET` | `/pending-count` | admin | 待审核用户数量 |

### 6.2 商品模块 `/api/products/*`

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| `GET` | `/` | 已登录 | 商品列表（支持分类/上架状态/低库存筛选，按 sort_order 排序） |
| `GET` | `/{id}` | 已登录 | 商品详情（含 stock_warning、thumbnail_url 等完整信息） |
| `POST` | `/` | admin | 创建商品 |
| `PATCH` | `/{id}` | admin | 更新商品（库存低于预警值时自动触发 Telegram 通知） |
| `DELETE` | `/{id}` | admin | 软删除（设 is_active=false，不真正删除） |

### 6.3 订单模块 `/api/orders/*`

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| `POST` | `/` | 已登录（已审核通过） | **创建订单**：验证库存→原子扣库存→创建订单+明细→月结联动→TG通知 |
| `GET` | `/` | 已登录 | 订单列表（商户只看自己的，管理员看全部，支持多维筛选） |
| `GET` | `/{id}` | 已登录 | 订单详情（含商品明细+商户信息+未回款天数+距结账日天数） |
| `PATCH` | `/{id}` | admin | 更新订单状态（支付/配送状态变更，联动月结金额） |
| `POST` | `/{id}/cancel` | 已登录 | 取消订单（回滚库存+扣减月结金额，已签收订单禁止取消） |

**订单创建核心逻辑**（`create_order`）：
1. 检查用户审核状态（必须 `approved`）
2. 验证月结权限（如选择月结支付方式）
3. 逐个验证商品（存在性 + 上架状态 + 库存充足）
4. 使用 **原子 SQL UPDATE** 扣库存（`WHERE stock >= quantity`，防并发超卖）
5. 生成订单号 `ORD{YYYYMMDD}{6位序号}`
6. 月结订单自动累加商户 `credit_limit`
7. 异步发送 Telegram 通知给管理员（新订单详情 + 库存预警）

**订单取消逻辑**（`cancel_order`）：
1. 权限校验（商户只能取消自己的）
2. 状态校验（已签收/已取消的禁止操作）
3. 遍历订单明细，逐个回滚商品库存
4. 月结订单扣减商户 `credit_limit`
5. 标记 `delivery_status = cancelled`

### 6.4 分类模块 `/api/categories/*`

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| `GET` | `/` | 已登录 | 活跃分类列表（按 sort_order 排序） |
| `GET` | `/all` | admin | 全部分类（含禁用的） |
| `POST` | `/` | admin | 创建分类（名称唯一性校验） |
| `PATCH` | `/{id}` | admin | 更新分类 |
| `DELETE` | `/{id}` | admin | 删除分类 |

### 6.5 公告模块 `/api/announcements/*`

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| `GET` | `/public` | **公开（无需登录）** | 获取激活的公告列表 |
| `GET` | `/` | admin | 全部公告（含禁用的） |
| `POST` | `/` | admin | 创建公告 |
| `PATCH` | `/{id}` | admin | 更新公告 |
| `DELETE` | `/{id}` | admin | 删除公告 |

**公告类型**：`notice`（首页滚动通知）、`contact`（联系客服信息）、`about`（关于系统信息）

### 6.6 月结模块 `/api/billing/*`

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| `GET` | `/` | 已登录 | 月结账单列表（商户看自己的，管理员看全部，支持年/月筛选） |
| `POST` | `/generate?year=&month=` | admin | 按年月生成月结账单 |
| `PATCH` | `/{id}` | admin | 更新账单（记录已付金额，自动计算状态） |

**账单生成逻辑**：
- 遍历所有有 `billing_day` 设置的商户
- 统计周期：上月 billing_day → 本月 billing_day（含当天整天）
- 汇总该时段内非取消订单总金额
- 自动创建 MonthlyBill 记录

### 6.7 上传模块 `POST /api/upload`

| 特性 | 说明 |
|------|------|
| 上传方式 | 流式分块读取（64KB/次），绝不将整个文件载入内存 |
| 大小限制 | 单文件最大 5MB（超限立即中止并删除残留文件） |
| 格式限制 | `.jpg`/`.jpeg`/`.png`/`.webp`/`.gif` 扩展名白名单 |
| 安全 | 磁盘空间预检（至少保留500MB可用）、UUID随机文件名（防注入） |
| 速率限制 | IP维度（1分钟≤20次）+ 用户维度（1分钟≤15次） |
| 优化 | 自动压缩（最大分辨率 1200×1200，JPEG 质量 85%） |
| 返回 | `{ "url": "/uploads/{uuid}.jpg", "filename": "{uuid}.jpg" }` |

---

## 七、安全体系 — 四层防护

### 7.1 认证层（JWT + bcrypt）

- **密码哈希**：bcrypt（passlib），自动加盐，抗彩虹表
- **JWT 签发**：python-jose HS256，payload `{"sub": user_id_string, "exp": timestamp}`
- **有效期**：7天（ACCESS_TOKEN_EXPIRE_MINUTES = 10080）
- **前端传输**：Axios 请求拦截器自动附加 `Authorization: Bearer {token}`
- **失效处理**：响应拦截器检测 401 → 自动调用 `logout()` → 清除 localStorage → 跳转登录页

### 7.2 速率限制层（三层滑动窗口，纯内存实现）

| 层级 | 规则 | 超限响应 | 实现类 |
|------|------|---------|--------|
| 全局 IP 限制 | 同 IP 1秒内 ≤ 30次请求 | HTTP 429 | `RateLimitMiddleware` |
| 登录防暴力 | 同 IP 5分钟 ≤ 10次；同账号 5分钟 ≤ 5次失败 | HTTP 429 + 锁定5分钟 | `LoginProtector` |
| 上传速率限制 | 同 IP 1分钟 ≤ 20次；同用户 1分钟 ≤ 15次 | HTTP 429 | `UploadRateLimiter` |

所有滑动窗口均使用线程锁（`threading.Lock`），带定时清理机制（每60秒清理过期记录），零外部依赖。

### 7.3 Telegram initData 签名验证

```
验证流程（遵循 Telegram 官方规范 https://core.telegram.org/bots/webapps#validating-data）：

1. parse_qsl(init_data) → 解析 URL 编码的键值对
2. 提取 hash 字段值 → received_hash
3. 剩余键值对按 key 字母序排列
4. data_check_string = "key1=val1\nkey2=val2\n..."（用换行符连接）
5. secret_key = HMAC-SHA256(key="WebAppData", msg=bot_token)
6. computed_hash = HMAC-SHA256(key=secret_key, msg=data_check_string)
7. hmac.compare_digest(computed_hash, received_hash) → 防时序攻击
8. 验证 auth_date 不超过 24 小时（防重放攻击）
9. JSON.parse(user) → 提取 telegram_id, first_name, last_name, username
```

**Bot Token**: `8768383980:AAHm4h_tAewLc7XJ6qhPbP93hF4kiQCFkHo`

### 7.4 权限控制层（RBAC）

- **依赖注入**：`get_current_user` → 解码 JWT 获取 user_id → 查数据库获取 User 对象
- **角色检查器**：`RoleChecker([UserRole.ADMIN])` → 通过 `Depends(require_admin)` 注入
- **超级管理员保护**：用户名 `100001` 不能被禁用/删除
- **数据隔离**：商户只能查看和操作自己的订单、账单

---

## 八、前端架构 — 双端自适应

### 8.1 入口引导流程 (`main.js`)

```
应用启动
  │
  ├─ 创建 Pinia（注册 persistedstate 插件 → localStorage 持久化）
  ├─ 注册 Vue Router / vue-i18n / Element Plus(zh-cn|en) / Vant 4
  │
  ├─ bootstrapTelegram()  ← 异步执行，完成后才 mount
  │   ├─ 非 TG 环境 → return false（跳过）
  │   ├─ 是 TG 环境:
  │   │   ├─ initTelegramWebApp() → WebApp.ready() + WebApp.expand()
  │   │   ├─ 已有 token → fetchUserInfo() 验证有效性
  │   │   │   ├─ 有效 → return true
  │   │   │   └─ 失效 → logout() → 继续下方重新登录
  │   │   └─ 无 token → telegramLogin(initData)
  │   │       → POST /api/auth/telegram-auth
  │   │       → 保存 token + userInfo 到 Pinia/localStorage
  │   │       → return true
  │
  └─ app.mount('#app')  ← 登录完成后才挂載 DOM
```

### 8.2 路由系统 (`router/index.js`)

**三套路由+自动切换**：

| 路径前缀 | UI 框架 | 用户角色 | 布局组件 | 页面数 |
|----------|---------|---------|----------|--------|
| `/m/*` | Vant 4 | merchant | MobileLayout（底部4个Tab：首页/购物车/订单/我的） | 4 |
| `/merchant/*` | Element Plus | merchant | MerchantLayout（左侧边栏） | 4 |
| `/admin/*` | Element Plus | admin | AdminLayout（左侧边栏+顶部头栏） | 8 |
| `/login`, `/register` | Element Plus | 公开 | 无布局 | 2 |

**路由守卫逻辑**（`beforeEach`）：
1. 未登录 + 需认证页面 → 跳转 `/login`
2. 已登录访问 `/login` → 按角色重定向（admin→`/admin/dashboard`，merchant→移动/PC首页）
3. 角色不匹配 → 重定向到正确角色首页
4. **PC↔移动自动切换**：
   - 商户在移动端访问 `/merchant/*` → 自动映射到 `/m/*`（如 `/merchant/cart` → `/m/cart`）
   - 商户在 PC 端访问 `/m/*` → 自动映射到 `/merchant/*`
5. 根路径 `/` → 重定向到 `/m/shop`

### 8.3 状态管理

**用户 Store**（`stores/user.js`）：
- **持久化字段**：`token`, `userInfo` → `localStorage` key: `user`
- **计算属性**：`isLoggedIn`（token+userInfo 都存在）、`userRole`、`isAdmin`、`isMerchant`
- **方法**：
  - `login(username, password)` → POST /api/auth/login → 保存 token+user
  - `telegramLogin(initData)` → POST /api/auth/telegram-auth → 保存 token+user
  - `logout()` → 清空 token+userInfo
  - `fetchUserInfo()` → GET /api/auth/me → 更新 userInfo（失败则 logout）

**购物车 Store**（`stores/cart.js`）：
- **持久化字段**：`items[]` → `localStorage` key: `cart`
- **每个 item**：`{ id, name, price_usd, stock, quantity, image_url, unit }`
- **计算属性**：`totalCount`（总件数）、`totalPrice`（总价 USD）
- **方法**：
  - `addItem(product, quantity)` → 已有则加数量（检查库存上限），否则新增
  - `removeItem(productId)` → 删除
  - `updateQuantity(productId, quantity)` → 更新数量
  - `clearCart()` → 清空

### 8.4 HTTP 请求层 (`utils/request.js`)

- 基于 **Axios**，baseURL 为空（相对路径，由 Nginx 代理或 Vite 开发代理转发）
- **超时**：30 秒
- **请求拦截器**：每次请求从 Pinia store 实时读取 token → 设置 `Authorization: Bearer {token}`
- **响应拦截器**：
  - 成功响应：直接返回 `response.data`（剥离 Axios 封装层）
  - HTTP 401（非登录接口）：自动调用 `useUserStore().logout()`
  - 所有错误：`ElMessage.error(detail || message || i18n翻译的默认错误文案)`

### 8.5 Telegram Mini App 工具 (`utils/telegram.js`)

| 函数 | 返回值 | 说明 |
|------|--------|------|
| `getTelegramWebApp()` | `WebApp实例 \| null` | 读取 `window.Telegram?.WebApp` |
| `isTelegramMiniApp()` | `boolean` | `WebApp` 存在且 `initData` 非空 |
| `getInitData()` | `string` | `WebApp.initData`（URL编码的签名字符串） |
| `getTelegramUser()` | `object \| null` | `WebApp.initDataUnsafe.user`（id/first_name/username/photo_url） |
| `initTelegramWebApp()` | `void` | 调用 `WebApp.ready()` + `WebApp.expand()`（全屏展开） |

### 8.6 设备检测 (`utils/device.js`)

| 函数 | 说明 |
|------|------|
| `isMobile()` | UA 正则 + 屏幕宽度 <768px 综合判断 |
| `isTouchDevice()` | 检测 `ontouchstart` 或 `maxTouchPoints` |
| `getViewportWidth()` | 视口宽度 |
| `isSmallScreen()` | <768px |
| `hapticFeedback(style)` | 触觉反馈（`navigator.vibrate`，支持 light/medium/heavy/success/error） |

### 8.7 格式化工具 (`utils/format.js`)

| 函数 | 说明 |
|------|------|
| `usdToKhr(usd, rate=4000)` | 美元→瑞尔 |
| `khrToUsd(khr, rate=4000)` | 瑞尔→美元 |
| `formatUSD(amount)` | `$12.50` 格式 |
| `formatKHR(amount)` | `50,000 ៛` 格式（千分位+瑞尔符号） |
| `formatDateTime(dateString)` | 根据当前语言格式化日期时间 |

### 8.8 国际化 (`i18n/`)

- 支持 **中文（zh）** 和 **英文（en）**
- 语言选择持久化到 `localStorage('app-lang')`，默认中文
- Element Plus 的 locale 跟随切换（`zhCn` / `en`）
- 业务文案全部通过 `$t('key')` 引用

### 8.9 设计系统 (`styles/variables.scss`)

**色彩策略（无色系+品牌色点缀）**：
- 背景：`#F8F9FA`（应用）→ `#FFFFFF`（内容卡片）→ `#F5F6F7`（悬浮态）
- 文字：`#1A1A1A`（标题）→ `#4A4A4A`（正文）→ `#8C8C8C`（辅助）→ `#BFBFBF`（禁用）
- 边框：`#E8E8E8`（常规）→ `#F0F0F0`（浅色）→ `#D9D9D9`（深色）
- 品牌色：深邃宝石蓝 `#1D4ED8`（仅用于核心操作按钮）
- 语义色：成功 `#16A34A` / 预警 `#EA580C` / 危险 `#DC2626` / 信息 `#0891B2`

**几何系统**：
- 圆角：0px（直角工业感）→ 2px → 4px → 6px
- 间距：8px → 12px → 16px → 24px → 32px → 48px
- 阴影：全部禁用，改用边框分隔

### 8.10 页面清单（25 个 .vue 文件）

| 模块 | 文件 | 功能描述 |
|------|------|---------|
| 公共 | `Login.vue` | 登录页（账号+密码表单，支持 Enter 提交） |
| 公共 | `Register.vue` | 注册页（手机号+密码+姓名+地址，提交后等待审核） |
| 移动端 | `mobile/Shop.vue` | 商品主页（顶部分类Tab+搜索框+商品卡片网格+加入购物车） |
| 移动端 | `mobile/Shop_new.vue` | 新版商品列表（备用方案） |
| 移动端 | `mobile/Cart.vue` | 购物车（商品列表+数量加减+删除+总价+提交订单） |
| 移动端 | `mobile/Orders.vue` | 订单列表（顶部状态Tab筛选+订单卡片+取消订单） |
| 移动端 | `mobile/Profile.vue` | 个人中心（用户信息卡片+修改资料+修改密码+语言切换+退出） |
| PC商户 | `merchant/Products.vue` | 商品浏览（表格视图+分类筛选+搜索+加入购物车） |
| PC商户 | `merchant/Cart.vue` | 购物车（表格+数量编辑+提交） |
| PC商户 | `merchant/Orders.vue` | 我的订单（表格+状态筛选+详情弹窗+取消） |
| PC商户 | `merchant/Profile.vue` | 个人信息（表单编辑+密码修改） |
| 管理端 | `admin/Dashboard.vue` | 仪表盘（统计卡片：今日订单/总收入/库存预警/商户数+快捷操作） |
| 管理端 | `admin/Products.vue` | 商品管理（表格CRUD+图片上传+库存编辑+排序+上下架） |
| 管理端 | `admin/Orders.vue` | 订单管理（表格+多维筛选+状态变更下拉+配送调度） |
| 管理端 | `admin/Merchants.vue` | 商户管理（表格+查看详情+编辑资料+启禁用+月结设置） |
| 管理端 | `admin/Approvals.vue` | 用户审核（待审核列表+通过/拒绝按钮+拒绝原因输入） |
| 管理端 | `admin/Categories.vue` | 分类管理（表格+新增/编辑对话框+排序+启禁用） |
| 管理端 | `admin/Announcements.vue` | 公告管理（三种类型Tab+富文本编辑+中英双语） |
| 管理端 | `admin/Profile.vue` | 管理员个人中心 |
| 布局 | `layouts/AdminLayout.vue` | 管理端布局：左侧导航菜单+顶部头栏（用户名/退出） |
| 布局 | `layouts/MerchantLayout.vue` | PC商户布局：左侧导航+内容区 |
| 布局 | `layouts/MobileLayout.vue` | 移动端布局：内容区+底部4个Tab（带购物车角标） |
| 组件 | `components/SkeletonProduct.vue` | 商品卡片加载占位骨架屏 |
| 组件 | `components/SkeletonTable.vue` | 表格数据加载占位骨架屏 |

---

## 九、Telegram Mini App 集成 — 完整流程

### 9.1 前端接入

**`index.html`**：引入 Telegram SDK
```html
<script src="https://telegram.org/js/telegram-web-app.js"></script>
```

**`utils/telegram.js`**：5 个工具函数封装 `window.Telegram.WebApp`

**`main.js`**：应用启动时自动检测并执行 TG 登录（在 `app.mount()` 之前完成）

### 9.2 后端验证

**`core/telegram.py`**：HMAC-SHA256 签名验证（详见安全体系章节）

**`api/auth.py` → `/telegram-auth`**：
1. 验证 initData 签名 → 提取 `telegram_id`
2. 查找 `User.telegram_id == tg_id` 的已绑定用户
3. 未找到 → 自动创建新用户：
   - `username = "tg_{telegram_id}"`
   - `full_name = "{first_name} {last_name}"` 或 `"TG_{id}"`
   - `role = merchant`
   - `approval_status = approved`（TG签名验证即信任）
   - `hashed_password = bcrypt("tg_auto_{id}")`（随机密码，不用于登录）
4. 检查 `is_active` 和 `approval_status`
5. 签发 JWT → 返回 `{access_token, user}`

### 9.3 完整登录时序

```
[Telegram 客户端] 用户点击 Bot 菜单中的 Mini App
        │
        ▼
[Telegram] 向 https://khmerai.cn 加载 WebApp，注入 initData
        │
        ▼
[浏览器] 加载 index.html → telegram-web-app.js → main.js
        │
        ▼
[main.js] isTelegramMiniApp() === true
        │
        ▼
[main.js] initTelegramWebApp() → WebApp.ready() + WebApp.expand()
        │
        ├─ localStorage 有 token → fetchUserInfo() → GET /api/auth/me
        │   ├─ 200 → 用户信息有效，跳过登录
        │   └─ 401 → token 过期，logout() 清除 → 继续重新登录
        │
        ▼
[main.js] telegramLogin(WebApp.initData)
        │
        ▼
[stores/user.js] apiTelegramAuth(initData)
        │
        ▼
[api/index.js] POST /api/auth/telegram-auth { init_data: "..." }
        │
        ▼
[后端 auth.py] validate_init_data(init_data)
        │
        ├─ 签名不匹配 → 401 "签名验证失败"
        ├─ auth_date >24h → 401 "initData 已过期"
        │
        ▼
[后端] 查找/创建 User → 签发 JWT
        │
        ▼
[前端] 保存 token + userInfo → Pinia store + localStorage
        │
        ▼
[main.js] app.mount('#app')
        │
        ▼
[Router] 路由守卫 → isLoggedIn=true → 放行 → /m/shop（移动端商品页）
```

---

## 十、Nginx 配置详解 (`deploy/nginx.conf`)

```nginx
# 第一个 server 块：HTTP → HTTPS 强制跳转
server {
    listen 80;
    server_name khmerai.cn;

    # Let's Encrypt ACME 验证（certbot 证书续期需要）
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    # 其余所有请求 301 永久重定向到 HTTPS
    location / {
        return 301 https://$host$request_uri;
    }
}

# 第二个 server 块：HTTPS 主站
server {
    listen 443 ssl http2;
    server_name khmerai.cn;

    # Let's Encrypt SSL 证书
    ssl_certificate /etc/letsencrypt/live/khmerai.cn/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/khmerai.cn/privkey.pem;

    # SSL 安全参数
    ssl_protocols TLSv1.2 TLSv1.3;          # 仅允许安全协议版本
    ssl_ciphers HIGH:!aNULL:!MD5;            # 强加密套件
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # HSTS 头（强制浏览器后续请求使用 HTTPS，1年有效）
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # 前端静态文件（Vue SPA）
    root /opt/wholesale/dist;                 # 注意：非 /opt/wholesale/frontend/dist
    index index.html;

    # Vue Router history 模式：所有路径回退到 index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 反向代理到 FastAPI 后端
    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300;
    }

    # 上传的图片（直接由 Nginx 提供静态服务，30天缓存）
    location /uploads/ {
        alias /opt/wholesale/backend/uploads/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # 前端 JS/CSS 静态资源（1年缓存，Vite 会自动加 hash 文件名）
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # 上传文件大小限制
    client_max_body_size 20M;
}
```

---

## 十一、SQLite 数据库优化 (`core/database.py`)

| PRAGMA | 值 | 说明 |
|--------|-----|------|
| `journal_mode` | WAL | Write-Ahead Logging，允许并发读写 |
| `busy_timeout` | 5000ms | 锁等待超时，避免立即 SQLITE_BUSY |
| `synchronous` | NORMAL | 平衡写入性能与数据安全 |
| `cache_size` | -8000 (8MB) | 增大页面缓存，减少磁盘IO |

**连接池配置**：
- `pool_size=5`：最多 5 个常驻连接
- `max_overflow=0`：不允许溢出连接（防止内存不足）
- `pool_recycle=3600`：1小时回收空闲连接
- `pool_pre_ping=True`：使用前 ping 一次，防止用到断开的连接
- `check_same_thread=False`：允许 FastAPI 多线程复用连接

---

## 十二、部署流程 (`deploy_ssh.py`)

自动化部署脚本使用 **paramiko** 库通过 SSH/SFTP 执行：

```
步骤 1:  SSH 连接服务器 43.134.13.229 (ubuntu / qer1235A@)
步骤 2:  sudo systemctl stop wholesale
步骤 3:  删除旧数据库 rm -f *.db*
步骤 4:  SFTP 上传 30 个后端 Python 文件到 /opt/wholesale/backend/
         (models/ core/ api/ services/ + main.py + init_db.py)
步骤 5:  执行 /opt/wholesale/venv/bin/python init_db.py
         → 建表 + 创建管理员(100001) + 演示商户 + 5分类 + 4商品
步骤 6:  SFTP 上传 dist.tar.gz → 解压到 /opt/wholesale/dist/
步骤 7:  sudo chown -R ubuntu:ubuntu /opt/wholesale/
步骤 8:  sudo systemctl start wholesale
步骤 9:  验证 systemctl is-active = "active"
步骤 10: curl localhost:8000 → {"status": "running"}
步骤 11: POST /api/auth/login (100001/123456) → 获取 access_token ✅
```

---

## 十三、初始数据 (`init_db.py`)

| 类别 | 数据 |
|------|------|
| **超级管理员** | 用户名 `100001`，密码 `123456`，角色 admin，状态 approved，需改密码 |
| **演示商户** | 用户名 `merchant1`，密码 `123456`，角色 merchant，状态 approved |
| **5 个分类** | 饮料(sort=1)、食品(2)、调料(3)、日用品(4)、零食(5) |
| **4 个演示商品** | 可口可乐 330ml($12/箱,库存100)、康师傅方便面($8.5/箱,库存50)、白糖($25/袋,库存30)、食用油($18/桶,库存5) |

---

## 十四、已修复的关键 Bug（历史记录）

| # | Bug 描述 | 根本原因 | 修复方案 |
|---|---------|---------|---------|
| 1 | **TG 用户永远无法登录**（"账号正在审核中"） | `/telegram-auth` 创建用户时 `approval_status=PENDING`，但同一请求内立即检查审核状态返回 400 | 改为 `approval_status=APPROVED`（TG 签名验证→自动信任） |
| 2 | **管理员无法登录** | `init_db.py` 创建管理员用户名为 `admin`，但超管保护代码写死检查 `100001` | 管理员用户名改为 `100001`，使用正确的枚举类型 |
| 3 | **initData 签名验证失败** | 使用 `parse_qs`（返回 `{key: [value]}` 列表）而非 `parse_qsl`（返回 `[(key, value)]` 扁平对） | 改用 `parse_qsl` 保证正确构建 data-check-string |
| 4 | **TG 环境页面空白/API全部401** | 路由守卫在 TG 环境跳过认证检查（`next()`放行），导致未登录用户进入受保护页面 | 移除 TG 特殊放行，统一要求登录；`main.js` 在 mount 前完成登录 |
| 5 | **Token 过期后 TG 不重新登录** | `bootstrapTelegram()` 检测到 `isLoggedIn=true` 就直接跳过 | 改为先 `fetchUserInfo()` 验证 token 有效性，失效则 logout 后重新 telegramLogin |
| 6 | **时区不一致** | `approved_at` 使用 `datetime.now()` 无时区信息 | 统一使用 `datetime.now(timezone(timedelta(hours=7)))` 柬埔寨时区 |

---

## 十五、配置项一览 (`core/config.py`)

| 配置项 | 默认值 | 支持 .env 覆盖 | 说明 |
|--------|--------|---------------|------|
| `APP_NAME` | `柬埔寨批发管理系统` | ✅ | 应用名称 |
| `APP_VERSION` | `1.0.0` | ✅ | 版本号 |
| `DEBUG` | `False` | ✅ | 调试模式 |
| `DATABASE_URL` | `sqlite:///./cambodia_wholesale.db` | ✅ | 数据库连接字符串 |
| `SECRET_KEY` | `your-secret-key-...` | ✅ | JWT 签名密钥（**生产环境必须修改**） |
| `ALGORITHM` | `HS256` | ✅ | JWT 算法 |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | `10080`（7天） | ✅ | Token 有效期 |
| `ALLOWED_ORIGINS` | `["*"]` | ✅ | CORS 允许的来源 |
| `TELEGRAM_BOT_TOKEN` | `None` | ✅ | Telegram 通知用 Bot Token |
| `TELEGRAM_CHAT_ID` | `None` | ✅ | Telegram 通知目标 Chat ID |
| `TG_BOT_TOKEN` | `8768383980:AAHm4h_...` | ✅ | **Mini App 签名验证用** Bot Token |
| `USD_TO_KHR_RATE` | `4000.0` | ✅ | 美元→瑞尔汇率 |

配置优先级：`.env` 文件 > 环境变量 > 代码默认值

---

## 十六、前端 API 函数清单 (`api/index.js`)

共 **40+** 个 API 调用函数：

**认证类（15个）**：
`login` · `telegramAuth` · `getCurrentUser` · `register` · `publicRegister` · `changePassword` · `resetUserPassword` · `deleteUser` · `updateProfile` · `getPendingUsers` · `getAllRegistrations` · `approveUser` · `getPendingCount` · `getUserList` · `updateUser`

**商品类（5个）**：
`getProducts` · `getProduct` · `createProduct` · `updateProduct` · `deleteProduct`

**订单类（5个）**：
`getOrders` · `getOrder` · `createOrder` · `updateOrder` · `cancelOrder`

**分类类（5个）**：
`getCategories` · `getAllCategories` · `createCategory` · `updateCategory` · `deleteCategory`

**上传（1个）**：
`uploadImage`

**公告类（5个）**：
`getPublicAnnouncements` · `getAnnouncements` · `createAnnouncement` · `updateAnnouncement` · `deleteAnnouncement`

**月结类（3个）**：
`getMonthlyBills` · `generateMonthlyBills` · `updateMonthlyBill`

---

## 十七、当前线上状态

| 项目 | 状态 | 说明 |
|------|------|------|
| 后端服务 | ✅ 运行中 | systemd active，API 健康检查 200 |
| 管理员登录 | ✅ 正常 | `100001` / `123456` 测试通过 |
| 数据库 | ✅ 全新 | 重建后含管理员+商户+5分类+4商品 |
| SSL 证书 | ✅ 有效 | Let's Encrypt，到期 2026-07-14（自动续期） |
| Telegram 登录 | ✅ 代码已修复 | 需在 TG 客户端中实际验证 |
| 前端页面 | ⚠️ 待确认 | Nginx root 路径需确认是否指向 `/opt/wholesale/dist` |
