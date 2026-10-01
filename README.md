# LightDocs

LightDocs 是一个基于 Vue 3、TypeScript 和 Vite 的轻量文档管理后台。当前前端使用 Mock Service 运行，后端接口和数据库设计已整理完毕，便于替换为 Go + Gin + PostgreSQL 实现。

## 项目结构

```text
lightdocs/
├── frontend/
│   └── src/       # Vue 3 前端源码
├── backend/       # Go 后端（待实现）
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

## 后端设计文档

- [接口文档](./docs/backend-api.md)
- [数据库设计](./docs/database-design.md)
- [数据库建表脚本](./db/schema.sql)

接口文档覆盖登录鉴权、仪表盘、文章、公开文档、搜索、分类、图片、回收站和站点设置。数据库脚本面向 PostgreSQL 16，生产环境请通过迁移工具执行，并从环境变量注入管理员初始密码哈希。
