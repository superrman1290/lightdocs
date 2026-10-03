# LightDocs 数据库设计

> 数据库：PostgreSQL 16+
>
> 目标：支撑当前管理后台和公开文档页，保留后续多用户、对象存储、审计和全文搜索的扩展空间。

## 1. 设计原则

- 业务主键使用 `bigint generated always as identity`，会话、回收站记录使用 UUID。
- 登录采用不透明随机 access token，数据库只保存 token 哈希；JWT 暂不作为第一版会话实现。
- 所有业务表使用 `timestamptz`，由数据库保存 UTC 时间。
- 密码只保存 Argon2id/bcrypt 哈希；`/settings/admin` 接口永远不返回哈希。
- 文章和图片采用软删除：业务表写入 `deleted_at`，并在同一事务中写入 `recycle_bin` 快照；永久删除时再清理业务记录和对象存储文件。
- 文章正文保存 Markdown 原文；渲染 HTML 在应用层完成，避免数据库保存不一致的派生内容。
- 文章标签由服务端限制单个标签最多 8 个字符，正文默认最多 2 MiB；图片上传必须通过真实文件类型校验。
- 列表查询优先使用组合索引；管理端模糊搜索使用 PostgreSQL `pg_trgm`，公开全文搜索使用 `tsvector` 和 GIN。

## 2. 实体关系

```text
users 1 ───── n auth_sessions
  │
  ├──── n reauth_tokens（短期、一次性重新验证凭证）
  │
  ├──── n articles n ───── 1 categories
  │           │
  │           └──── n article_images n ───── 1 images
  │
  └──── n recycle_bin（按 item_type + item_id 保存资源快照和关联关系）

site_settings 1 ───── 1 security_settings
```

## 3. 表设计

### 3.1 `users` 管理员用户

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | bigint | PK | 用户 ID |
| `username` | varchar(30) | UNIQUE NOT NULL | 登录账号 |
| `password_hash` | varchar(255) | NOT NULL | Argon2id/bcrypt 哈希 |
| `role` | varchar(20) | NOT NULL | 当前固定为 `admin` |
| `status` | varchar(20) | NOT NULL | `active`、`disabled` |
| `failed_login_count` | integer | NOT NULL DEFAULT 0 | 连续失败次数 |
| `locked_until` | timestamptz | NULL | 锁定截止时间 |
| `last_login_at` | timestamptz | NULL | 最近登录时间 |
| `created_at` / `updated_at` | timestamptz | NOT NULL | 审计时间 |

### 3.2 `auth_sessions` 登录会话

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | uuid | PK | 会话 ID |
| `user_id` | bigint | FK users | 所属用户 |
| `token_hash` | char(64) | UNIQUE NOT NULL | 只保存 token 哈希 |
| `expires_at` | timestamptz | NOT NULL | 到期时间 |
| `last_seen_at` | timestamptz | NOT NULL | 最近使用时间 |
| `ip` | inet | NULL | 登录 IP |
| `user_agent` | text | NULL | 客户端信息 |
| `revoked_at` | timestamptz | NULL | 主动撤销时间 |
| `created_at` | timestamptz | NOT NULL | 创建时间 |

有效会话条件：`revoked_at IS NULL AND expires_at > now()`。

### 3.2a `reauth_tokens` 短期重新验证凭证

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | uuid | PK | 凭证记录 ID |
| `user_id` | bigint | FK users | 所属管理员 |
| `token_hash` | char(64) | UNIQUE NOT NULL | 只保存凭证哈希 |
| `expires_at` | timestamptz | NOT NULL | 默认 5 分钟后过期 |
| `created_at` | timestamptz | NOT NULL | 创建时间 |

重新验证成功后凭证立即删除，只允许使用一次；过期凭证会在生成新凭证时清理。

### 3.3 `site_settings` 站点设置

单例表，固定 `id = 1`。

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | smallint | PK CHECK (id = 1) | 单例标识 |
| `site_name` | varchar(30) | NOT NULL | 网站名称 |
| `site_title` | varchar(80) | NOT NULL | 浏览器标题 |
| `logo_url` | text | NOT NULL | Logo / Favicon 地址 |
| `updated_by` | bigint | FK users | 最近修改人 |
| `created_at` / `updated_at` | timestamptz | NOT NULL | 审计时间 |

### 3.4 `security_settings` 安全设置

同样为单例表，固定 `id = 1`。

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | smallint | PK CHECK (id = 1) | 单例标识 |
| `max_login_failures` | smallint | 1–20 | 最大失败次数 |
| `lock_minutes` | integer | 1–1440 | 锁定分钟数 |
| `session_days` | integer | 1–365 | 会话有效天数 |
| `remember_login` | boolean | NOT NULL | 是否允许保持登录 |
| `logout_other_devices_on_password_change` | boolean | NOT NULL | 修改密码时注销其他设备 |
| `require_password_reauth_on_save` | boolean | NOT NULL | 保存设置是否二次验证 |
| `updated_by` | bigint | FK users | 最近修改人 |
| `created_at` / `updated_at` | timestamptz | NOT NULL | 审计时间 |

### 3.5 `categories` 分类

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | bigint | PK | 分类 ID |
| `name` | varchar(80) | NOT NULL | 分类名 |
| `parent_id` | bigint | FK categories(id) | 父分类，可空 |
| `sort_order` | integer | NOT NULL DEFAULT 0 | 同级排序 |
| `created_by` / `updated_by` | bigint | FK users | 操作人 |
| `created_at` / `updated_at` | timestamptz | NOT NULL | 审计时间 |

删除策略为 `ON DELETE RESTRICT`，数据库触发器和应用层事务校验共同保证不能形成父子循环。

### 3.6 `articles` 文章

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | bigint | PK | 文章 ID |
| `title` | varchar(200) | NOT NULL | 标题 |
| `slug` | varchar(160) | UNIQUE NOT NULL | 公开 URL 标识 |
| `category_id` | bigint | FK categories | 所属分类 |
| `status` | varchar(20) | NOT NULL | `draft` / `published` |
| `tags` | text[] | NOT NULL DEFAULT '{}' | 标签数组 |
| `content` | text | NOT NULL DEFAULT '' | Markdown 正文 |
| `summary` | varchar(500) | NOT NULL DEFAULT '' | 摘要 |
| `published_at` | timestamptz | NULL | 首次/最近发布时间 |
| `created_by` / `updated_by` | bigint | FK users | 操作人 |
| `deleted_at` | timestamptz | NULL | 软删除时间 |
| `deleted_by` | bigint | FK users | 删除人 |
| `search_vector` | tsvector | GENERATED | 全文搜索向量 |
| `created_at` / `updated_at` | timestamptz | NOT NULL | 审计时间 |

`category` 名称不单独落库，接口查询通过 JOIN 得到，避免分类改名后出现脏数据。软删除文章不参与公开列表、后台列表和 active slug 唯一性校验。

### 3.7 `images` 图片资源

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | bigint | PK | 图片 ID |
| `name` | varchar(255) | NOT NULL | 原始文件名 |
| `storage_key` | text | UNIQUE NOT NULL | 对象存储 key 或本地相对路径 |
| `url` | text | NOT NULL | 对外访问地址 |
| `mime_type` | varchar(100) | NOT NULL | MIME 类型 |
| `size_bytes` | bigint | NOT NULL | 精确字节数 |
| `source` | varchar(20) | NOT NULL | `upload` / `article` / `system` |
| `width` / `height` | integer | NULL | 图片尺寸 |
| `created_by` | bigint | FK users | 上传人 |
| `deleted_at` | timestamptz | NULL | 软删除时间 |
| `deleted_by` | bigint | FK users | 删除人 |
| `created_at` | timestamptz | NOT NULL | 上传时间 |

接口层将 `size_bytes` 转换为前端兼容的 `size = ceil(size_bytes / 1024)`。

### 3.8 `article_images` 文章图片引用

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `article_id` | bigint | PK/FK | 文章 ID |
| `image_id` | bigint | PK/FK | 图片 ID |
| `created_at` | timestamptz | NOT NULL | 建立引用时间 |

图片可能在 Markdown 正文中直接引用 URL。服务端在保存文章时解析站内 `/uploads/...` URL 并维护该表；外部图片不建立站内资源关系。

文章新建和更新必须在同一事务中重建该文章的关系集合；图片页面和删除接口以该表作为引用判断依据。文章软删除时保留关系，文章恢复时从快照补回关系，文章永久删除时由外键级联清理。

### 3.9 `recycle_bin` 回收站

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | uuid | PK | 回收站记录 ID |
| `item_type` | varchar(20) | NOT NULL | `article` / `image` |
| `item_id` | bigint | NOT NULL | 原资源 ID |
| `snapshot` | jsonb | NOT NULL | 删除时完整快照；文章必须包含 `article_image_ids` |
| `deleted_by` | bigint | FK users | 删除人 |
| `deleted_at` | timestamptz | NOT NULL | 删除时间 |

不对 `item_type + item_id` 建唯一约束，允许同一资源 ID 在不同时间产生多个删除快照，但恢复时必须使用指定记录；旧快照不得覆盖更新后的资源。

### 3.10 `audit_logs` 审计日志

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | bigint | PK | 日志 ID |
| `user_id` | bigint | FK users | 操作人 |
| `action` | varchar(80) | NOT NULL | 如 `article.delete` |
| `resource_type` / `resource_id` | varchar / bigint | NULL | 目标资源 |
| `payload` | jsonb | NULL | 脱敏后的变更摘要 |
| `ip` | inet | NULL | 请求 IP |
| `created_at` | timestamptz | NOT NULL | 操作时间 |

密码、token、完整 Markdown 正文不得写入审计 `payload`。

## 4. 索引设计

1. `articles`: active 数据上的 `(status, updated_at DESC)`、`(category_id, status, updated_at DESC)`、`slug` partial unique 索引。
2. `articles`: `GIN(tags)`；启用 `pg_trgm` 后对 `title`、`summary`、`slug` 建 trigram 索引；对 `search_vector` 建 GIN 全文索引。
3. `categories`: `parent_id, sort_order, id`；同一父级下名称大小写不敏感唯一。
4. `images`: `(source, created_at DESC)`、`created_at DESC`、`name gin_trgm_ops`。
5. `recycle_bin`: `(item_type, deleted_at DESC)`、`deleted_at DESC`，`snapshot` 不直接建立全文索引。
6. `auth_sessions`: `user_id, revoked_at, expires_at`，并对 `token_hash` 建唯一索引。
7. `article_images`: `(image_id, article_id)` 反向索引，用于图片引用查询和删除保护。

## 5. 关键事务

### 5.1 删除文章

```text
BEGIN
  SELECT article FOR UPDATE
  INSERT recycle_bin(item_type='article', item_id, snapshot)
  UPDATE articles SET deleted_at = now(), deleted_by = ? WHERE id = ?
  INSERT audit_logs(action='article.delete')
COMMIT
```

### 5.2 恢复文章

```text
BEGIN
  SELECT recycle_bin FOR UPDATE
  校验 slug、category_id 是否仍可用
  UPDATE articles SET deleted_at = NULL, deleted_by = NULL WHERE id = ?
  恢复 snapshot.article_image_ids 对应的 article_images 关系
  DELETE recycle_bin WHERE id = ?
  INSERT audit_logs(action='article.restore')
COMMIT
```

图片删除使用相同的软删除事务。只要 `article_images` 存在引用，图片就不能删除；文章永久删除后关系由外键级联清理。

### 5.3 修改管理员密码

```text
BEGIN
  校验当前密码
  更新 users.password_hash、updated_at
  按 security_settings.logout_other_devices_on_password_change 撤销会话
  INSERT audit_logs(action='user.password_change')
COMMIT
```

## 6. 初始化数据

- 初始化一个 `admin` 用户，密码只能由部署脚本通过环境变量传入并生成哈希，禁止把默认明文密码提交到仓库。
- 初始化 `site_settings`：`轻文档`、`轻文档 - 专注技术教程的个人文档网站`、`/favicon.svg`。
- 初始化 `security_settings`：失败 5 次锁定 15 分钟，会话 7 天。
- 分类和文章可以通过导入脚本写入，不建议把演示数据混入生产迁移。
