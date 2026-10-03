# LightDocs 后端接口文档

> 文档版本：v1.0
>
> 对应前端：`frontend/src/types`、`frontend/src/services`、`frontend/src/stores/siteSettings`
>
> 当前实现：Go + Gin + PostgreSQL 16

## 1. 接口约定

### 1.1 基础信息

- 基础路径：`/api/v1`
- 管理端接口统一使用 `Authorization: Bearer <access_token>`；第一版 access token 为不透明随机 token，服务端只保存其哈希。
- 时间统一使用 ISO 8601 UTC，例如 `2026-09-30T10:32:00Z`
- ID 使用正整数或 UUID 均可；本文示例沿用前端的正整数资源 ID
- 请求和响应编码为 UTF-8 JSON
- 文件上传使用 `multipart/form-data`

### 1.2 统一响应结构

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "requestId": "req_01J8..."
}
```

分页响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [],
    "total": 42,
    "page": 1,
    "pageSize": 10
  },
  "requestId": "req_01J8..."
}
```

错误响应：

```json
{
  "code": 40101,
  "message": "账号或密码不正确",
  "data": null,
  "requestId": "req_01J8..."
}
```

分页参数约束：`page` 从 1 开始，`pageSize` 默认 10，最大 100。超出范围时服务端应截断或返回参数错误，建议统一截断到允许范围。

### 1.3 通用业务对象

#### 文章 `Article`

```json
{
  "id": 1,
  "title": "Vue 3 项目搭建指南",
  "slug": "vue-3-project-guide",
  "categoryId": 1,
  "category": "前端开发",
  "status": "published",
  "tags": ["Vue", "前端"],
  "content": "# Vue 3 项目搭建指南",
  "summary": "介绍如何从零搭建 Vue 3 项目。",
  "createdAt": "2026-09-28T10:00:00Z",
  "updatedAt": "2026-09-30T10:32:00Z",
  "publishedAt": "2026-09-28T10:00:00Z"
}
```

`status` 取值：`draft`（草稿）或 `published`（已发布）。`slug` 全局唯一，只允许小写字母、数字和连字符。

#### 分类 `Category`

```json
{
  "id": 1,
  "name": "前端开发",
  "parentId": null,
  "sort": 1
}
```

#### 图片 `ImageAsset`

```json
{
  "id": 1,
  "name": "docker-compose.png",
  "url": "https://cdn.example.com/images/2026/09/docker-compose.png",
  "size": 128,
  "sizeBytes": 131072,
  "mimeType": "image/png",
  "source": "article",
  "createdAt": "2026-09-24T14:32:00Z"
}
```

`size` 为兼容当前前端展示的 KiB 整数，数据库以 `size_bytes` 保存精确字节数。`source` 取值：`upload`、`article`、`system`。

## 2. 登录与会话

### 2.1 创建登录会话

`POST /auth/session`

请求：

```json
{
  "username": "admin",
  "password": "your-password",
  "rememberMe": false
}
```

响应 `200`：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "accessToken": "ld_at_01J8...",
    "tokenType": "Bearer",
    "expiresAt": "2026-10-08T10:32:00Z",
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin"
    }
  }
}
```

处理规则：

1. 密码使用 Argon2id 或 bcrypt 校验，禁止明文比对或明文存储。
2. 登录失败累计 `failed_login_count`，达到安全设置中的 `maxLoginFailures` 后锁定 `lockMinutes` 分钟。
3. `rememberMe=false` 使用短会话；`true` 使用 `sessionDays` 作为有效期。
4. 登录成功后重置失败次数并记录 IP、User-Agent、最后登录时间。

账号处于锁定期间或本次失败已达到锁定阈值时返回 HTTP `423`、业务码 `40102`；临时锁定到期后，失败计数窗口自动重置。

### 2.2 当前用户

`GET /auth/me`

返回当前用户，不返回密码和密码哈希。

### 2.3 删除当前登录会话

`DELETE /auth/session`

撤销当前会话，返回 `204 No Content` 或统一成功响应均可。

### 2.4 删除全部登录会话

`DELETE /auth/sessions`

管理员在“设置 → 安全 → 清除会话”中调用。默认撤销当前管理员除当前会话外的全部会话；若产品希望全部退出，可在响应中返回 `forceReLogin: true`。

### 2.5 重新验证密码

`POST /auth/re-auth`

请求当前密码，成功返回短时有效的 `reauthToken`。当安全设置启用
`requirePasswordReauthOnSave` 时，敏感设置请求必须携带：

```http
X-Reauth-Token: <reauth_token>
```

凭证有效期为 5 分钟，服务端只保存其哈希；每个凭证成功通过一次敏感设置请求后立即失效，不能重复使用。验证失败返回 `40103`。

## 3. 仪表盘

### 3.1 获取概览

`GET /dashboard/overview`

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "statistics": {
      "articles": 24,
      "categories": 8,
      "images": 126
    },
    "latestArticles": [
      {
        "id": 1,
        "title": "Vue 3 项目搭建指南",
        "category": "前端开发",
        "status": "published",
        "tags": ["Vue", "前端"],
        "updatedAt": "2026-09-30T10:32:00Z"
      }
    ]
  }
}
```

`latestArticles` 默认返回最近更新的 5 条文章，后台可通过 `limit` 查询参数覆盖，最大 20。

## 4. 文章

### 4.1 文章列表

`GET /articles`

查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `status` | string | 否 | `published`、`draft`；不传表示全部 |
| `keyword` | string | 否 | 匹配标题、摘要、正文、标签 |
| `categoryId` | integer | 否 | 分类 ID |
| `page` | integer | 否 | 默认 1 |
| `pageSize` | integer | 否 | 默认 10，最大 100 |

响应 `data` 为分页结构，`list` 中返回 `Article`。

### 4.2 获取文章详情

`GET /articles/:id`

返回完整 `Article`，包含 Markdown `content`。

### 4.3 新建文章

`POST /articles`

请求：

```json
{
  "title": "Vue 3 项目搭建指南",
  "slug": "vue-3-project-guide",
  "categoryId": 1,
  "status": "draft",
  "tags": ["Vue", "前端"],
  "content": "# Vue 3 项目搭建指南",
  "summary": "介绍如何从零搭建 Vue 3 项目。"
}
```

校验：标题、slug、分类必填；slug 唯一；标签去重并限制单个标签最多 8 个字符；正文默认最多 2 MiB，可通过 `MAX_ARTICLE_CONTENT_BYTES` 配置。成功返回创建后的完整 `Article`。

### 4.4 更新文章

`PATCH /articles/:id`

请求字段与新建相同，全部可选，只更新提交字段。修改分类时同步返回新的 `category` 名称。若文章从草稿变为已发布，服务端写入 `publishedAt`；从已发布改为草稿时清空 `publishedAt`。

### 4.5 删除文章（移入回收站）

`DELETE /articles/:id`

不要直接物理删除。服务端在同一事务中把文章完整快照（包含 `article_image_ids`）写入回收站，再设置 `articles.deleted_at` 和 `deleted_by`。返回 `204 No Content`。

### 4.6 导出文章

`GET /articles/:id/export?format=markdown`

`format` 取值：`markdown`、`html`、`pdf`。响应使用文件流，并设置：

```http
Content-Disposition: attachment; filename="vue-3-project-guide.md"
```

## 5. 公开文档与搜索

### 5.1 公开文章详情

`GET /public/docs/:slug`

只允许访问 `published` 文章，返回：

```json
{
  "id": 1,
  "title": "Vue 3 项目搭建指南",
  "slug": "vue-3-project-guide",
  "category": "前端开发",
  "tags": ["Vue", "前端"],
  "summary": "介绍如何从零搭建 Vue 3 项目。",
  "content": "# Vue 3 项目搭建指南",
  "updatedAt": "2026-09-30T10:32:00Z"
}
```

### 5.2 全文搜索

`GET /public/search`

查询参数：`keyword`（必填）、`categoryId`、`page`、`pageSize`。只搜索已发布文章的标题、摘要、正文和标签，使用数据库全文索引并返回分页结果；可额外返回 `highlight` 片段供搜索结果页展示。

管理端若复用搜索组件，可调用 `GET /articles` 并携带管理员令牌，以包含草稿。

## 6. 分类

### 6.1 分类树

`GET /categories`

返回扁平数组 `Category[]`，按 `parentId`、`sort`、`id` 排序。前端可自行构造树；如需要直接渲染树，可通过 `tree=true` 返回嵌套结构。

### 6.2 新建分类

`POST /categories`

```json
{
  "name": "前端开发",
  "parentId": null,
  "sort": 1
}
```

同一父分类下名称不允许重复，`parentId` 不能指向自身或不存在的分类；父级变更由数据库触发器和应用层事务校验共同保证不能形成循环。

### 6.3 更新分类

`PATCH /categories/:id`

请求字段同新建且均可选。修改父级时校验不能形成循环。

### 6.4 删除分类

`DELETE /categories/:id`

若存在子分类或文章引用，返回 `40901`，提示先移动或删除关联数据；默认不级联删除文章。

## 7. 图片资源

### 7.1 图片列表

`GET /images`

查询参数：`keyword`、`source`、`page`、`pageSize`。`source` 取值：`upload`、`article`、`system`。

### 7.2 上传图片

`POST /images`

请求 `multipart/form-data`：

- `files`: 一个或多个图片文件

限制：单文件默认最大 10 MB（由 `MAX_UPLOAD_BYTES` 配置）；允许 `image/jpeg`、`image/png`、`image/webp`、`image/gif`、`image/svg+xml`；服务端会读取文件内容校验真实 MIME，不信任扩展名或客户端传入的 `Content-Type`。

响应：`ImageAsset[]`。图片上传到对象存储或本地文件服务后，数据库只保存 `storage_key` 和可访问 `url`。

### 7.3 删除图片

`DELETE /images?ids=1,2,3`

只有未被 `article_images` 引用的图片，才会写入 `deleted_at` 并进入回收站。只要存在引用，图片页面删除请求直接返回 `40902`，不会创建回收站记录。回收站中的图片永久删除同样受 `article_images` 外键约束保护；引用关系会在对应文章永久删除后由外键级联清理。

## 8. 回收站

### 8.1 回收站列表

`GET /recycle-bin`

查询参数：`type`（`article`、`image`）、`keyword`、`page`、`pageSize`。返回分页结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [
      {
        "id": "article-1-1727685120000",
        "type": "article",
        "data": {},
        "deletedAt": "2026-09-30T11:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  },
  "requestId": "req_01J8..."
}
```

`data` 的结构与原始 `Article` 或 `ImageAsset` 一致，便于前端直接渲染。

### 8.2 更新回收站项目

`PATCH /recycle-bin/:id`

请求：

```json
{
  "action": "restore"
}
```

恢复到原资源表并删除回收站记录。若原 slug、分类、文件名已冲突，返回 `40903` 并附带 `conflicts` 字段；恢复必须是事务操作。

### 8.3 永久删除

`DELETE /recycle-bin/:id`：永久删除单条。若目标是仍被 `article_images` 引用的图片，返回 `40902`，必须先永久删除引用文章或移除文章中的图片引用。

`DELETE /recycle-bin?ids=article-1-1727685120000,image-2-1727685120001`

`DELETE /recycle-bin`：清空回收站。若其中包含仍被 `article_images` 引用的图片，操作返回 `40902`，无法删除的图片记录会保留；该操作不可恢复，建议服务端要求二次确认标记 `X-Confirm-Destructive: true`，并记录审计日志。

## 9. 站点与安全设置

### 9.1 获取站点设置

`GET /settings/site`

返回：

```json
{
  "siteName": "轻文档",
  "siteTitle": "轻文档 - 专注技术教程的个人文档网站",
  "logoUrl": "/uploads/logo.svg"
}
```

公开页面可使用 `GET /public/settings/site`，仅返回这三个字段，不需要管理员令牌。

### 9.2 更新站点设置

`PATCH /settings/site`

请求字段：`siteName`、`siteTitle`、`logoUrl`，均可选但不能保存空名称或空标题。Logo 建议先通过图片上传接口得到 URL，再提交设置。

### 9.3 获取和更新安全设置

`GET /settings/security`

`PATCH /settings/security`

字段：

```json
{
  "maxLoginFailures": 5,
  "lockMinutes": 15,
  "sessionDays": 7,
  "rememberLogin": false,
  "logoutOtherDevicesOnPasswordChange": true,
  "requirePasswordReauthOnSave": false
}
```

范围：`maxLoginFailures` 1–20，`lockMinutes` 1–1440，`sessionDays` 1–365。若修改操作本身要求重新验证密码，必须先调用 `POST /auth/re-auth`。

### 9.4 管理员信息与密码

`GET /settings/admin`：只返回 `id`、`username`，绝不返回密码、密码后四位或密码哈希。

`PATCH /settings/admin`：

```json
{
  "username": "admin",
  "currentPassword": "旧密码",
  "newPassword": "至少 8 位的新密码",
  "confirmPassword": "至少 8 位的新密码"
}
```

修改密码成功后，若 `logoutOtherDevicesOnPasswordChange=true`，撤销除当前会话外的所有会话。

## 10. 错误码

| HTTP | 业务码 | 说明 |
| --- | --- | --- |
| 400 | 40000 | 请求参数错误 |
| 400 | 40001 | 图片文件类型不支持或文件内容无效 |
| 401 | 40100 | 未登录或令牌无效 |
| 401 | 40101 | 账号或密码错误 |
| 401 | 40103 | 需要重新验证密码或验证凭证已失效 |
| 423 | 40102 | 账号已锁定 |
| 403 | 40300 | 无权限 |
| 404 | 40400 | 资源不存在 |
| 409 | 40901 | 分类存在子分类或文章引用 |
| 409 | 40902 | 图片仍被引用 |
| 409 | 40903 | 回收站恢复发生唯一性冲突 |
| 413 | 41300 | 上传文件过大 |
| 422 | 42200 | 业务校验失败 |
| 422 | 42201 | 标签长度或格式校验失败 |
| 422 | 42202 | 文章正文超过长度限制 |
| 429 | 42900 | 请求过于频繁 |
| 500 | 50000 | 服务端内部错误 |

## 11. 安全与运维要求

1. 登录接口按 IP + 账号限流，建议每分钟最多 10 次失败请求。
2. Access Token 建议使用短期 JWT（15–30 分钟）或随机不透明令牌；服务端必须保存会话撤销状态。
3. CORS 只允许配置的前端域名，生产环境必须使用 HTTPS。
4. 所有写操作记录 `created_by`、`updated_by` 或审计日志；永久删除、清空回收站、修改密码必须重点记录。
5. 上传文件使用不可预测的存储名，禁止把用户文件名直接拼接到路径；SVG 文件应做脚本清理或禁止外部脚本。
6. Markdown 渲染到 HTML 时做 XSS 过滤，禁止直接信任文章正文中的原始 HTML。

## 12. 当前前端接入说明

1. 当前 `frontend/src/services/*` 已通过 `frontend/src/services/apiClient.ts` 调用真实 REST API；响应会统一解包 `data`，非 2xx 或非零 `code` 会转换为前端异常。
2. `Article.category` 是接口层 JOIN `categories.name` 后的展示字段，不需要前端重复请求后再拼接。
3. 数据库的 `images.size_bytes` 已转换为当前前端 `ImageAsset.size` 使用的 KiB 整数，同时返回 `sizeBytes`。
4. 日期由页面按本地显示需要格式化，数据库和接口仍统一使用 UTC ISO 8601。
5. 登录页使用 `POST /auth/session` 获取不透明 `accessToken`，当前保存在 `sessionStorage`，管理端路由守卫检查令牌并由接口返回的 `401` 触发清理和重新登录；生产环境可进一步改用 HttpOnly、Secure、SameSite Cookie。
