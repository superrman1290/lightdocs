# LightDocs 后端设计文档

| 文档 | 内容 |
| --- | --- |
| [backend-api.md](./backend-api.md) | REST API、请求参数、响应结构、错误码、安全约定 |
| [database-design.md](./database-design.md) | PostgreSQL 实体、字段、索引、事务和数据流设计 |
| [../db/schema.sql](../db/schema.sql) | PostgreSQL 16 建表及初始化脚本 |

接口文档以当前 Vue 前端的类型和 Service 调用为准，数据库脚本以文档中的表设计为准。实际开发时建议使用迁移工具拆分执行，并通过环境变量注入管理员初始密码哈希。
