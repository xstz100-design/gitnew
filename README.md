# 柬埔寨批发管理系统

面向柬埔寨批发业务的 B2B 下单与管理系统，支持 PC 管理后台、PC 商户端、移动商户端和 Telegram Mini App 场景。

系统目标不是做一个通用商城，而是围绕批发业务里的几个核心问题展开：价格保护、商户审核、库存安全、信用额度、订单追踪和移动端快速下单。

## 项目简介

本项目采用前后端分离架构：

- 前端使用 Vue 3 + Vite，PC 端基于 Element Plus，移动端基于 Vant。
- 后端使用 FastAPI + SQLModel，默认运行在 SQLite WAL 模式，也支持通过 `DATABASE_URL` 切换到 PostgreSQL。
- 登录体系同时支持管理员账号密码登录和 Telegram Mini App 免登录。
- 商户下单链路内置审核状态校验、权限收口、限流、防重复提交和库存扣减保护。

线上运行环境：

- 线上域名：`https://khmerai.cn`
- 服务器：Ubuntu
- 反向代理：Nginx
- 应用服务：Uvicorn + systemd

## 核心能力

### 双端业务架构

- PC 管理端：面向管理员，处理商品、订单、商户、公告、分类和审核流程。
- PC 商户端：适合办公室或仓库场景，进行商品浏览、购物车编辑和订单查询。
- 移动商户端：适合 Telegram 内和手机浏览器场景，强调低操作成本和高频下单效率。

### 权限与账号体系

- `super admin`：数据库字段驱动，不再依赖写死用户名判断。
- `admin`：可管理商品、订单、分类、公告、商户和审核流程。
- `merchant`：只能访问自己的资料、购物车和自己的订单。

### 商户下单与信用控制

- 支持现结与月结两种支付方式。
- 商户未审核、资料不完整或被拒绝时不能下单，但可以先浏览商品并保留购物车。
- 支持商户信用控制、账期展示、未回款天数展示和管理员侧月结管理。

### Telegram 集成

- Telegram Mini App 自动登录。
- 管理员通知：新订单、库存预警。
- 商户通知：下单成功、订单完成。
- 管理员 Telegram 绑定和通知开关已支持。

### 移动端体验优化

- 自动识别移动设备和 Telegram 场景，按设备跳转到对应路由。
- iPhone 安全区适配、Telegram 视口高度适配。
- 购物车、订单、个人中心针对移动端做了交互简化。
- 下单按钮具备 loading 与禁用态，避免弱网下重复点击造成重复下单。

## 技术栈

### 前端

- Vue 3
- Vite 5
- Vue Router 4
- Pinia + pinia-plugin-persistedstate
- Element Plus
- Vant 4
- vue-i18n
- Axios
- Sass

### 后端

- FastAPI 0.109.2
- SQLModel 0.0.16
- Uvicorn
- python-jose
- passlib[bcrypt]
- Pillow
- loguru

### 数据与部署

- 默认数据库：SQLite（WAL 模式）
- 可切换数据库：PostgreSQL
- Web 服务：Nginx
- 进程守护：systemd
- SSL：Let's Encrypt

## 当前实际架构

```text
用户浏览器 / Telegram Mini App
            |
            v
       Nginx 80/443
            |
   +--------+---------+
   |                  |
   v                  v
Vue dist         FastAPI /api
静态资源         Uvicorn 127.0.0.1:8000
                      |
                      v
                SQLite / PostgreSQL
```

当前生产目录约定：

- 前端目录：`/opt/wholesale/frontend/dist`
- 后端目录：`/opt/wholesale/backend`
- 虚拟环境：`/opt/wholesale/venv`
- systemd 服务名：`wholesale`
- 数据库文件：`/opt/wholesale/backend/cambodia_wholesale.db`

## 目录结构

```text
backend/
├── app/
│   ├── api/                 # 路由层：认证、商品、订单、分类、公告、上传、月结
│   ├── core/                # 配置、数据库、依赖注入、安全、限流、日志
│   ├── models/              # User、Product、Order、OrderItem 等模型
│   └── services/            # Telegram 通知、图片处理等业务服务
├── init_db.py               # 初始化数据库与演示数据
├── main.py                  # FastAPI 应用入口
└── requirements.txt

frontend/
├── public/
├── src/
│   ├── api/                 # Axios API 封装
│   ├── components/          # 骨架屏等基础组件
│   ├── i18n/                # 中英文文案
│   ├── layouts/             # Admin / Merchant / Mobile 三套布局
│   ├── router/              # 路由和守卫
│   ├── stores/              # user / cart 状态管理
│   ├── styles/              # 全局样式和变量
│   ├── utils/               # request、telegram、device、format 等工具
│   └── views/               # admin / merchant / mobile 页面
├── index.html
├── package.json
└── vite.config.js

deploy/
├── nginx.conf               # 标准 Nginx 配置
├── bt_nginx.conf            # 宝塔 Nginx 配置
├── setup.sh                 # 传统部署初始化脚本
├── bt_setup.sh              # 宝塔部署脚本
├── backup_db.sh             # 数据库备份脚本
├── wholesale.service        # systemd 服务定义
└── BT_DEPLOY.md             # 宝塔部署文档
```

## 功能清单

### 管理端

- 仪表板：销售统计、订单概览、库存预警
- 商品管理：新增、编辑、上下架、软删除、多图上传字段支持
- 订单管理：查看订单、更新支付状态、更新配送状态
- 商户管理：创建用户、重置密码、启停用户、账期和信用控制
- 超级管理员管理：提升为超级管理员、取消超级管理员
- 分类管理：分类排序和维护
- 公告管理：公告内容与展示控制
- 审核流：商户资料审核、拒绝原因记录

### 商户端

- 商品浏览、搜索、分类筛选
- 购物车持久化
- 提交订单、查看历史订单、取消未完成订单
- 完善资料、修改密码、查看审核状态
- Telegram 绑定与通知开关

### 移动端

- 商品卡片浏览
- 购物车全选、分项调整、快捷提交
- 订单列表筛选、下拉刷新、详情弹层
- 个人中心资料维护
- 针对 Telegram Mini App 的视口和安全区适配

## 权限模型

### 超级管理员

- 拥有管理员全部能力
- 可管理其他管理员的超级管理员状态
- 受最后一个超级管理员保护逻辑约束，避免误操作导致系统失去最高权限账号

### 普通管理员

- 可管理商户、订单、商品、分类、公告
- 不应访问超级管理员专属的危险操作

### 商户

- 只能访问自己的资料和自己的订单
- 订单详情查询和取消操作都带商户维度限制

## 安全与反作弊

当前代码已具备以下防护能力：

### 接口权限收口

- 商户订单列表只返回自己的订单
- 商户订单详情只允许查询自己的订单
- 商户取消订单只允许操作自己的订单
- 管理员接口通过依赖注入进行角色校验

### 限流

- 全局 IP 速率限制中间件
- 登录防暴力破解限制
- 上传接口限流
- 下单接口专门限流，降低脚本刷单与价格探测风险

### 重复下单防护

- 前端提交订单时生成一次性 `client_request_id`
- 后端按 `merchant_id + client_request_id` 做幂等校验
- 数据库层有唯一索引兜底

### 数据安全

- `.env` 已被 `.gitignore` 忽略，不会默认提交到版本库
- 商品与订单增加了软删除语义，避免误删后无法对账
- 商品库存使用原子扣减，降低并发超卖风险
- Axios 超时和 401 统一收口处理

### 日志与可观测性

- 使用 loguru 接管错误日志
- `ERROR` 级别以上日志单独写入 `logs/error.log`
- 服务异常由 systemd 自动拉起

## 数据模型说明

### 用户模型

用户模型包含以下关键字段：

- `role`：`admin` / `merchant`
- `is_super_admin`：是否超级管理员
- `approval_status`：`pending` / `approved` / `rejected`
- `allow_credit`：是否允许月结
- `billing_cycle_days`：账期天数
- `notify_enabled`：是否接收 Telegram 通知
- `must_change_password`：首次登录是否必须改密

### 商品模型

商品重点字段：

- 名称、分类、单位、规格、条码
- 批发价、建议零售价
- 当前库存、库存预警值
- 推荐位标记
- `is_active`：是否上架
- `is_deleted`：是否软删除

### 订单模型

订单重点字段：

- `order_no`：订单号
- `merchant_id`：商户 ID
- `payment_status`：未支付 / 现结 / 月结
- `delivery_status`：待派送 / 送货中 / 已签收 / 已取消
- `client_request_id`：防重复提交请求号
- `is_deleted`：软删除标记

## 本地开发

### 1. 启动后端

```bash
cd backend

python -m venv venv
venv\Scripts\activate

pip install -r requirements.txt
python init_db.py
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

后端默认地址：`http://localhost:8000`

OpenAPI 文档：

- Swagger UI：`http://localhost:8000/docs`
- ReDoc：`http://localhost:8000/redoc`

### 2. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端默认地址：`http://localhost:5173`

### 3. 初始化说明

`python init_db.py` 会执行以下动作：

- 创建数据库表
- 创建首个超级管理员
- 创建演示商户
- 创建演示分类
- 创建演示商品

## 初始账号与密码策略

系统已不再依赖固定默认密码。

- 超级管理员默认用户名为 `100001`
- 初始化脚本会为超级管理员生成临时密码，并在控制台输出
- 演示商户 `merchant1` 也会生成临时密码
- 管理员后台新建用户或重置密码时，会返回一次性临时密码
- 所有临时密码都应在首次登录后立即修改

## 环境变量

后端配置位于 `backend/app/core/config.py`，支持 `.env` 覆盖。常用配置如下：

| 变量 | 说明 | 默认值 |
| ---- | ---- | ------ |
| `APP_NAME` | 应用名称 | 柬埔寨批发管理系统 |
| `APP_VERSION` | 应用版本 | 1.0.0 |
| `DEBUG` | 调试模式 | false |
| `DATABASE_URL` | 数据库连接串 | sqlite:///./cambodia_wholesale.db |
| `SECRET_KEY` | JWT 密钥 | 内置默认值，生产建议显式覆盖 |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | Token 有效期 | 10080 |
| `ALLOWED_ORIGINS` | CORS 来源列表 | `[*]` |
| `TG_BOT_TOKEN` | Telegram Mini App 校验 Token | 当前配置可读到默认值 |
| `TELEGRAM_BOT_TOKEN` | 通知 Bot Token | 可选 |
| `TELEGRAM_CHAT_ID` | 默认通知 Chat ID | 可选 |
| `USD_TO_KHR_RATE` | 汇率 | 4000 |

生产建议：

- 显式配置 `SECRET_KEY`
- 显式配置 `DATABASE_URL`
- 按实际域名收紧 `ALLOWED_ORIGINS`
- 不要把生产 `.env` 提交到版本库

## API 分组

主要 API 路由如下：

- `/api/auth`：登录、注册、Telegram 认证、用户资料、管理员用户管理
- `/api/products`：商品列表、详情、创建、编辑、软删除
- `/api/orders`：订单创建、列表、详情、取消、管理员状态更新
- `/api/categories`：分类管理
- `/api/announcements`：公告管理
- `/api/upload`：文件上传
- `/api/billing`：月结账单相关接口

## 前端交互约定

### 路由切换

- 移动设备访问 `/merchant/*` 会自动跳到 `/m/*`
- PC 设备访问 `/m/*` 会自动跳回 `/merchant/*`
- Telegram 环境下会优先尝试自动登录

### 请求层规则

- Axios 默认超时为 15 秒
- 401 会清空登录态并跳回登录页
- 错误消息统一由请求层提示

### 购物车与下单

- 购物车数据通过 Pinia 持久化到本地
- 下单时会立刻进入提交状态并禁用按钮
- 防止弱网下重复点击造成重复订单

## 部署说明

### 方式一：传统部署

适合标准 Ubuntu 环境。

```bash
cd /opt/wholesale
bash deploy/setup.sh
```

核心文件：

- `deploy/setup.sh`
- `deploy/nginx.conf`
- `deploy/wholesale.service`

### 方式二：宝塔部署

适合需要图形化面板的环境。

```bash
cd /opt/wholesale
bash deploy/bt_setup.sh
```

核心文件：

- `deploy/bt_setup.sh`
- `deploy/bt_nginx.conf`
- `deploy/BT_DEPLOY.md`

### 生产更新建议流程

推荐顺序：

1. 本地构建前端 `npm run build`
2. 打包后端和前端发布归档
3. 上传到服务器 `/tmp`
4. 解压到 `/opt/wholesale/backend` 和 `/opt/wholesale/frontend`
5. 重启 `wholesale`
6. 校验 `/`、`/api` 和关键业务页面

## 备份与恢复

项目自带数据库备份脚本：

- 脚本位置：`deploy/backup_db.sh`
- 默认备份目录：`/opt/wholesale/backups`
- 默认保留天数：7 天
- 支持本地压缩备份
- 预留了邮件和 SCP 异地同步开关

建议上线前补充：

- 异地对象存储备份
- 定期恢复演练
- 上传目录外置存储或挂载独立磁盘

## 上线检查清单

上线前建议至少确认以下事项：

- 管理员和商户账号都可以正常登录
- 商户资料未审核时不能下单
- 商户只能看到自己的订单
- 重复点击下单不会生成重复订单
- 商品软删除后历史订单仍可正常查看
- Telegram 新订单通知和库存预警通知正常
- 备份脚本已配置定时任务
- 日志目录可写，`logs/error.log` 能正常生成

## 常见问题

### 移动端打开白屏或布局异常

检查项：

1. 是否走到了 `/m/shop`
2. 是否在 Telegram 或 iPhone 环境下正确拿到视口高度
3. 是否引入了全局样式和 Vant

### 商户无法下单

常见原因：

1. 资料不完整
2. 审核状态不是 `approved`
3. 没有填写地址或电话
4. 选择了月结但商户没有月结权限

### 订单重复

正常情况下不会重复生成，因为前后端都有防重逻辑。若仍出现重复，需要优先检查：

1. 客户端是否绕过了正常下单流程
2. 反向代理是否重复转发请求
3. 数据库唯一索引是否已建立

### 图片 404

检查项：

1. 后端 `uploads` 目录是否存在
2. Nginx 是否正确映射 `/uploads`
3. 前端是否使用了正确的图片路径

### 401 后频繁跳登录

检查项：

1. `SECRET_KEY` 是否变动过
2. 前后端时间是否严重漂移
3. 浏览器本地缓存的 token 是否已失效

## 后续建议

当前系统已经可用于实际业务，但如果继续向“稳定上线”推进，优先级建议如下：

### 高优先级

1. 为上传目录接入对象存储或独立挂载
2. 补充 Alembic 正式迁移链路
3. 增加关键接口自动化测试
4. 明确生产环境 `.env` 管理方式

### 中优先级

1. Redis 缓存热点商品与统计
2. WebSocket 实时订单通知
3. 导出 Excel / PDF
4. 完善审计日志

### 长期规划

1. 配送员端
2. 财务报表
3. 多仓库管理
4. 小程序或更深度的 Telegram 运营能力

## 许可证

MIT License

---

构建基础：FastAPI + Vue 3 + Element Plus + Vant  
设计目标：稳定、易维护、适合批发业务高频操作
