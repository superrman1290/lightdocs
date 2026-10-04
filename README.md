# LightDocs

LightDocs 是一个基于 Vue 3、TypeScript、Go、Gin 和 PostgreSQL 的轻量文档管理系统。前端业务数据统一通过后端 RESTful API 读写。

## 项目结构

```text
lightdocs/
├── frontend/
│   └── src/       # Vue 3 前端源码
├── backend/       # Go + Gin 后端
├── public/
├── db/
├── docs/
├── package.json
└── vite.config.ts
```

## 开发

```bash
npm install
npm run dev
```

生产构建：

```bash
npm run build
```

## 部署与启动

本地 Windows 开发步骤保留在本节。Linux 生产服务器请使用完整的
[Linux 部署指南](./docs/linux-deployment.md)，其中包含 Nginx、systemd、HTTPS、备份和恢复配置。

### 1. 环境要求

- Docker Desktop（用于运行 PostgreSQL 16）
- Go 1.23 或更高版本
- Node.js 20 或更高版本，以及 npm

### 2. 配置环境变量

复制根目录的环境变量示例，并按部署环境修改 PostgreSQL 配置：

```powershell
Copy-Item .env.example .env
```

根目录 `.env` 用于 Docker Compose 创建 PostgreSQL 容器。后端运行时还需要设置 `DATABASE_URL`、`HTTP_ADDR`、`FRONTEND_ORIGINS` 等变量。后端程序不会自动读取 `.env` 文件，因此需要在启动终端中设置这些变量，或由进程管理器注入。可以复制后端示例作为配置参考：

```powershell
Copy-Item backend/.env.example backend/.env
```

不要把 `.env` 或 `backend/.env` 提交到 Git；生产环境请使用部署平台的密钥管理功能注入密码和数据库连接串。

### 3. 启动 PostgreSQL

在项目根目录执行：

```powershell
docker compose up -d postgres
docker compose ps
```

确认 `lightdocs-postgres` 显示 `healthy` 后再执行数据库初始化。

### 4. 初始化数据库

在 PowerShell 中设置后端数据库连接：

```powershell
$env:DATABASE_URL="postgres://lightdocs:change-this-password@localhost:5432/lightdocs?sslmode=disable"
```

首次部署空数据库时执行迁移：

```powershell
cd backend
go run ./cmd/migrate
```

后续数据库结构变更应使用独立迁移脚本；不要在已有生产数据库上反复执行完整建表脚本。

首次部署时创建管理员。密码只通过环境变量传入，程序保存的是 bcrypt 哈希：

```powershell
$env:SEED_ADMIN_USERNAME="admin"
$env:SEED_ADMIN_PASSWORD="请替换为强密码"
go run ./cmd/seed
```

如果数据库中已经有文章和图片，执行一次历史文章图片关系回填：

```powershell
go run ./cmd/backfillimages
```

回填命令可以重复执行，不会重复创建 `article_images` 关系。

### 5. 启动 Go 后端

在 `backend` 目录设置运行参数：

```powershell
$env:HTTP_ADDR=":8080"
$env:FRONTEND_ORIGINS="http://localhost:5173,http://127.0.0.1:5173"
go run ./cmd/server
```

后端健康检查地址：

```text
http://127.0.0.1:8080/api/v1/health
```

### 6. 启动 Vue 前端

另开一个终端，回到项目根目录执行：

```powershell
cd D:\Tools\LightDocs\lightdocs
npm install
npm run dev -- --host 127.0.0.1 --port 5173
```

浏览器访问：

```text
http://127.0.0.1:5173/
```

管理后台登录页为 `/login`，游客文档页为 `/docs`。

### 7. 生产构建

构建前端静态文件：

```powershell
npm run build
```

构建产物位于 `dist/`。Go 后端可编译为独立可执行文件：

```powershell
cd backend
go build -o lightdocs-server.exe ./cmd/server
```

生产环境应通过反向代理提供 HTTPS，将前端静态文件指向 `dist/`，并将 `/api/` 和 `/uploads/` 转发到 Go 后端；同时把 `FRONTEND_ORIGINS` 修改为实际前端域名。

### 8. 停止服务

停止 PostgreSQL 容器但保留数据库卷：

```powershell
docker compose stop postgres
```

删除容器但保留数据卷：

```powershell
docker compose down
```

不要在生产环境使用 `docker compose down -v`，该命令会删除 PostgreSQL 数据卷。

## 后端设计文档

- [Linux 生产部署指南](./docs/linux-deployment.md)
- [接口文档](./docs/backend-api.md)
- [数据库设计](./docs/database-design.md)
- [数据库建表脚本](./db/schema.sql)

接口文档覆盖登录鉴权、仪表盘、文章、公开文档、搜索、分类、图片、回收站和站点设置。数据库脚本面向 PostgreSQL 16，生产环境请通过迁移工具执行，并从环境变量注入管理员初始密码哈希。
