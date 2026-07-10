# 内网 API 调试工具 - otester - PRD

## 1. 文档信息

| 项目 | 内容 |
|---|---|
| 产品名称 | otester |
| 产品类型 | 基于 Go + Wails 的桌面 API 调试工具 |
| 目标平台 | Windows 优先，兼容 macOS / Linux |
| 主要用途 | 内网 API 调试、请求构造、OAuth 2.0 Client Credentials 鉴权 |
| 文档版本 | v1.0 |
| 状态 | 初稿 |

---

## 2. 产品背景

企业内网环境中的 API 往往无法通过公网 SaaS 工具直接访问，同时还可能存在安全、合规、账号隔离及网络代理等限制。开发、测试和运维人员需要一款可离线运行、可从本地配置加载 API 列表、支持 Microsoft OAuth 2.0 Client Credentials 鉴权，并具备类似 Postman 请求调试能力的桌面应用。

本产品采用 Go 语言和 Wails 框架开发，程序启动后从当前运行目录读取 `config.json`，加载允许访问的 API Endpoint、HTTP 请求参数及 OAuth 配置。用户可在图形界面中选择 API、修改请求参数、获取 Access Token，并向内网 API 发起调试请求。

---

## 3. 产品目标

1. 提供轻量、可离线部署的桌面 API 调试工具。
2. 从程序当前运行目录的 `config.json` 读取可访问的 API Endpoint 列表。
3. 支持常见 HTTP Method、Headers、Query Parameters 和 Body 编辑。
4. 支持直接访问 API。
5. 支持 Microsoft OAuth 2.0 Client Credentials 鉴权。
6. 自动将获取到的 Access Token 注入 API 请求的 `Authorization` Header。
7. 提供请求、响应、耗时、状态码和错误信息展示。
8. 提供具有 Fluent Design 特征的亚克力视觉风格。
9. 支持 `light`、`dark`、`system` 三种主题模式，并在顶部菜单中切换。
10. 尽量避免敏感信息泄露，不在日志和界面中明文暴露 `client_secret` 和完整 Access Token。

---

## 4. 非目标

以下功能不属于首个版本范围：

- 云端账号体系及跨设备同步。
- 团队协作、在线 Workspace 或远程共享。
- 公网 API Marketplace。
- OAuth Authorization Code、Device Code 等其他 OAuth 流程。
- 自动化测试流水线和复杂断言脚本。
- WebSocket、gRPC、GraphQL 专用调试能力。
- 插件市场。
- 云端保存 Client Secret。
- 完整替代 Postman 的所有功能。

---

## 5. 目标用户

### 5.1 后端开发人员

需要快速调试内网服务、查看请求与响应、验证接口参数和鉴权结果。

### 5.2 测试人员

需要使用预置 API 配置重复发起请求，验证状态码、响应体及异常情况。

### 5.3 运维及支持人员

需要通过受控的 Endpoint 列表诊断内网服务状态，不希望手工编写 Curl 命令。

---

## 6. 核心使用场景

### 6.1 直接访问 API

1. 用户启动程序。
2. 程序读取当前运行目录中的 `config.json`。
3. 用户从左侧 Endpoint 列表选择一个 API。
4. 界面自动加载 URL、Method、Headers、Query Parameters 和 Body。
5. 用户可临时修改请求参数。
6. 用户点击“发送”。
7. 程序向目标 API 发起请求。
8. 界面展示状态码、耗时、响应 Headers 和响应 Body。

### 6.2 使用 Microsoft OAuth 访问 API

1. 用户选择启用 OAuth 的 API。
2. 程序从配置中读取 `org_id_uuid`、`client_id`、`client_secret`、`scope`。
3. 程序向 Microsoft Token Endpoint 发送 `application/x-www-form-urlencoded` 请求。
4. 成功后解析 `access_token`、`token_type`、`expires_in`。
5. 程序将 Token 缓存在内存中。
6. 程序自动添加 Header：

   ```http
   Authorization: Bearer <access_token>
   ```

7. 程序向目标 API 发起请求。
8. 若 Token 未过期，后续请求复用 Token。
9. 若 Token 已过期或即将过期，程序重新获取 Token。
10. 界面展示 API 请求结果，但不显示完整 Token。

---

## 7. 功能需求

## 7.1 应用启动与配置加载

### 7.1.1 配置文件位置

程序必须从程序运行时当前工作目录读取：

```text
./config.json
```

说明：

- 当前工作目录指程序启动时的 Working Directory。
- 不默认读取用户目录或系统配置目录。
- 后续版本可增加命令行参数覆盖配置路径，但不属于首版强制范围。

### 7.1.2 配置加载行为

启动时：

1. 检查 `config.json` 是否存在。
2. 检查文件是否可读。
3. 解析 JSON。
4. 校验必填字段。
5. 将 Endpoint 列表加载至界面。
6. 若配置加载失败，显示明确错误，不允许静默忽略。

错误示例：

- 找不到 `config.json`。
- JSON 格式无效。
- Endpoint 缺少 `name`、`url` 或 `method`。
- OAuth 配置缺少必要字段。
- HTTP Method 不受支持。
- URL 格式无效。

### 7.1.3 配置重新加载

顶部菜单提供“重新加载配置”操作。

重新加载时：

- 重新读取当前工作目录中的 `config.json`。
- 保留当前主题设置。
- 清除已加载的 Endpoint 缓存。
- OAuth Token 是否清除应由配置项控制，默认清除。
- 若新配置无效，保留旧配置并显示错误提示。

---

## 7.2 Endpoint 列表

左侧导航区域展示所有可访问 API。

每个 Endpoint 至少显示：

- 名称。
- HTTP Method。
- 可选分组。
- 可选说明。
- OAuth 状态标识。
- 启用或禁用状态。

支持：

- 按名称搜索。
- 按分组筛选。
- 按 HTTP Method 筛选。
- 折叠分组。
- 显示最近使用的 Endpoint。
- 配置中标记为禁用的 Endpoint 不允许发送请求。

---

## 7.3 请求编辑器

### 7.3.1 HTTP Method

支持：

- GET
- POST
- PUT
- PATCH
- DELETE
- HEAD
- OPTIONS

Method 默认来自配置文件，用户可在界面中临时切换。

### 7.3.2 URL

支持：

- 完整 URL。
- HTTP 和 HTTPS。
- Query Parameters 自动拼接。
- URL 中的变量替换。
- URL 合法性校验。

首版建议支持简单变量语法：

```text
{{variable_name}}
```

变量来源：

- 全局变量。
- Endpoint 变量。
- 用户当前会话临时变量。

### 7.3.3 Headers

以 Key-Value 表格形式编辑。

每行包含：

- 启用复选框。
- Header Key。
- Header Value。
- 删除按钮。

行为要求：

- 支持新增、修改、删除。
- Header Key 不区分大小写。
- 重复 Header 是否允许由配置决定，默认允许。
- OAuth 模式下自动插入 `Authorization`。
- 自动插入的 Authorization Header 默认锁定，用户可选择覆盖。
- `client_secret` 不得作为普通 Header 自动发送给业务 API。

### 7.3.4 Query Parameters

以 Key-Value 表格形式编辑。

每行包含：

- 启用复选框。
- Key。
- Value。
- 删除按钮。

行为要求：

- 自动 URL Encode。
- 仅启用的参数参与请求。
- 支持重复 Key。
- 实时更新最终请求 URL 预览。

### 7.3.5 Body

支持以下类型：

- None
- JSON
- Text
- `application/x-www-form-urlencoded`
- Raw

JSON 模式要求：

- 语法高亮。
- JSON 格式校验。
- 格式化。
- 压缩。
- 显示行号。

当 Method 通常不使用 Body 时，可显示提示，但不强制禁止。

### 7.3.6 请求超时

- 全局默认超时：30 秒。
- Endpoint 可单独覆盖。
- 用户可在当前请求中临时修改。
- 超时后显示可识别的错误状态。

### 7.3.7 TLS 设置

首版建议支持：

- 默认校验证书。
- 可通过配置允许跳过 TLS 证书校验。
- 跳过证书校验时，界面必须显示明显警告。
- 默认不允许关闭 TLS 校验。

---

## 7.4 Microsoft OAuth 2.0 Client Credentials

### 7.4.1 Token Endpoint

默认格式：

```text
https://login.microsoftonline.com/{org_id_uuid}/oauth2/v2.0/token
```

其中：

- `org_id_uuid` 从 `config.json` 读取。
- 配置可允许直接指定完整 `token_url`。
- 若同时存在 `token_url` 和 `org_id_uuid`，优先使用 `token_url`。

### 7.4.2 请求格式

HTTP Method：

```http
POST
```

Content-Type：

```http
application/x-www-form-urlencoded
```

表单字段：

```text
client_id=<client_id>
scope=<scope>
client_secret=<client_secret>
grant_type=client_credentials
```

默认 Scope：

```text
https://graph.microsoft.com/.default
```

配置文件必须允许覆盖 Scope，以适配自定义 API：

```text
api://<application-id>/.default
```

### 7.4.3 Token 响应

预期格式：

```json
{
  "token_type": "Bearer",
  "expires_in": 3599,
  "ext_expires_in": 3599,
  "access_token": "******"
}
```

程序至少解析：

- `token_type`
- `expires_in`
- `access_token`

可选解析：

- `ext_expires_in`

### 7.4.4 Token 缓存

要求：

- Token 仅缓存在进程内存中。
- 默认不写入磁盘。
- 按 OAuth 配置唯一标识缓存。
- 缓存 Key 建议由以下字段组合：
  - `token_url`
  - `client_id`
  - `scope`
- 在 Token 到期前提前刷新。
- 默认刷新缓冲时间：60 秒。
- 应使用单飞机制，避免多个并发请求重复刷新 Token。
- 应提供“清除 Token 缓存”操作。

### 7.4.5 Authorization Header 注入

成功获取 Token 后，向业务 API 添加：

```http
Authorization: Bearer <access_token>
```

规则：

- 若 `token_type` 存在，优先使用返回值。
- 若返回值为空，默认使用 `Bearer`。
- 若配置中已存在 Authorization Header：
  - 默认由 OAuth 自动值覆盖。
  - 可通过配置指定是否允许用户自定义覆盖。
- 日志中必须脱敏。

### 7.4.6 OAuth 错误处理

必须覆盖：

- 无法连接 Token Endpoint。
- DNS 错误。
- TLS 错误。
- 超时。
- 400 / 401 / 403。
- 返回非 JSON。
- 缺少 `access_token`。
- `expires_in` 非法。
- Client Secret 错误。
- Tenant / Org ID 错误。
- Scope 错误。

界面应展示：

- 错误类型。
- HTTP 状态码。
- Microsoft 返回的错误码。
- Microsoft 返回的错误描述。
- 可安全展示的响应片段。

不得展示：

- 完整 Client Secret。
- 完整 Access Token。
- 包含敏感信息的完整请求。

---

## 7.5 请求执行流程

### 7.5.1 直接请求

```text
选择 Endpoint
→ 合并配置和用户临时修改
→ 校验 URL 与参数
→ 构造 HTTP Request
→ 发送请求
→ 解析响应
→ 展示结果
```

### 7.5.2 OAuth 请求

```text
选择 Endpoint
→ 判断是否启用 OAuth
→ 检查 Token 缓存
→ Token 不可用时请求 Microsoft Token Endpoint
→ 获取并缓存 Access Token
→ 注入 Authorization Header
→ 构造业务 API Request
→ 发送请求
→ 展示结果
```

### 7.5.3 请求取消

- 用户发送请求后，发送按钮切换为“取消”。
- 用户点击取消后，Go Context 取消底层 HTTP Request。
- UI 显示“请求已取消”。
- 取消业务 API 请求不强制清除 Token。

---

## 7.6 响应查看器

### 7.6.1 响应概览

显示：

- HTTP 状态码。
- 状态文本。
- 请求耗时。
- 响应大小。
- Content-Type。
- 请求时间。
- 是否经过 OAuth。
- Token 是否来自缓存。

### 7.6.2 Response Body

支持：

- JSON 格式化。
- JSON 折叠。
- 文本查看。
- Raw 查看。
- 搜索。
- 复制。
- 保存到文件。
- 大响应体保护。

默认最大内存响应体建议为 10 MB，可配置。

超过限制时：

- 停止完整加载。
- 显示截断提示。
- 可选择保存原始响应到本地文件。

### 7.6.3 Response Headers

以表格展示：

- Header Name。
- Header Value。
- 支持复制单项。
- 支持复制全部。

### 7.6.4 请求详情

提供只读的最终请求视图：

- 最终 URL。
- Method。
- Headers。
- Query Parameters。
- Body。
- Curl 预览。

敏感字段必须脱敏。

---

## 7.7 请求历史

首版建议保留当前会话历史。

每条记录包含：

- Endpoint 名称。
- Method。
- URL。
- 时间。
- 状态码。
- 耗时。
- 成功或失败状态。
- 是否使用 OAuth。

安全要求：

- 默认不持久化完整 Request Body 和 Response Body。
- 默认不保存 Access Token。
- 默认不保存 Client Secret。
- 可配置是否将非敏感请求历史写入本地。
- 提供“清除历史”。

---

## 7.8 顶部菜单

顶部菜单至少包含：

### 文件

- 重新加载配置。
- 打开配置所在目录。
- 退出。

### 视图

- Light。
- Dark。
- System。
- 重置布局。
- 显示或隐藏侧栏。

### 工具

- 清除 Token 缓存。
- 清除请求历史。
- 打开日志目录。
- 配置校验。

### 帮助

- 关于。
- 版本信息。
- 配置文件格式说明。

主题切换入口必须位于顶部菜单，并提供当前选中状态。

---

## 8. UI 与视觉设计

## 8.1 总体风格

采用 Fluent Design 方向的“亚克力”视觉语言：

- 半透明背景。
- 轻度背景模糊。
- 适度饱和度。
- 柔和高光。
- 浅色透明卡片。
- 等距圆角。
- 轻量阴影。
- Fluent 蓝及相近蓝色作为主要强调色。
- 干净的线性图标。
- 适中的科技感。
- 避免过度霓虹、强烈玻璃反射或影响可读性的透明度。

### 8.1.1 亚克力材质建议

卡片：

```css
background: rgba(255, 255, 255, 0.58);
backdrop-filter: blur(18px) saturate(135%);
border: 1px solid rgba(255, 255, 255, 0.42);
box-shadow:
  0 8px 30px rgba(0, 0, 0, 0.10),
  inset 0 1px 0 rgba(255, 255, 255, 0.45);
border-radius: 14px;
```

深色模式可使用：

```css
background: rgba(28, 32, 40, 0.68);
backdrop-filter: blur(18px) saturate(125%);
border: 1px solid rgba(255, 255, 255, 0.10);
box-shadow:
  0 8px 30px rgba(0, 0, 0, 0.32),
  inset 0 1px 0 rgba(255, 255, 255, 0.08);
```

说明：实际实现需兼容不同 Wails WebView 平台对 `backdrop-filter` 的支持情况，并提供无模糊降级样式。

## 8.2 布局

建议采用三栏结构：

```text
┌─────────────────────────────────────────────────────┐
│ 顶部菜单 / 主题切换 / 配置状态 / 发送按钮            │
├─────────────┬────────────────────────┬──────────────┤
│ Endpoint    │ Request Editor         │ Response     │
│ 列表        │ URL / Params / Body    │ 状态与内容   │
│ (可隐藏)    │ Headers / Auth         │ Headers      │
├─────────────┴────────────────────────┴──────────────┤
│ 状态栏：配置路径、网络状态、版本、请求耗时            │
└─────────────────────────────────────────────────────┘
```

在较窄窗口中：

- Response 区域可切换到下方。
- Endpoint 侧栏可折叠。
- 主要操作按钮保持可见。

## 8.3 组件规范

### 卡片

- 圆角：12–16 px。
- 边框：1 px 半透明高光。
- 阴影：低对比度柔和阴影。
- 内边距：12–20 px。

### 按钮

- 主按钮使用 Fluent 蓝。
- 圆角：8–10 px。
- Hover 有轻微亮度和阴影变化。
- Active 有轻微下压。
- Disabled 降低饱和度和不透明度。
- “发送”按钮在请求中变为“取消”。

### 输入框

- 半透明底色。
- 焦点状态使用蓝色描边和轻微外发光。
- 保持高对比度。
- Secret 字段默认掩码。

### 图标

使用线性图标，建议统一采用：

- Fluent UI System Icons。
- Lucide。
- 其他风格一致的 SVG 图标集。

不得混用明显不同风格的图标。

### 状态颜色

- 成功：绿色。
- 警告：琥珀色。
- 错误：红色。
- 信息：Fluent 蓝。
- 禁用：中性灰。

颜色不能作为唯一状态信息，必须配合文本或图标。

---

## 9. 主题系统

## 9.1 模式

支持：

- `light`
- `dark`
- `system`

### Light

- 浅色半透明卡片。
- 深色文本。
- 柔和蓝色强调。
- 背景可使用浅蓝灰渐变。

### Dark

- 深灰蓝半透明卡片。
- 浅色文本。
- 蓝色强调。
- 控制反射高光，避免低对比度。

### System

- 跟随操作系统主题。
- 监听系统主题变化。
- 无需重启应用即可切换。
- 手动选择 Light 或 Dark 后，不再跟随系统。

## 9.2 主题持久化

主题偏好保存在本地非敏感设置中。

建议：

- 使用 Wails Runtime 或前端 Local Storage。
- 存储值仅为 `light`、`dark`、`system`。
- 不写入业务 `config.json`，避免污染部署配置。

---

## 10. 配置文件设计

## 10.1 示例配置

```json
{
  "app": {
    "name": "otester tester",
    "title": "Sample API tester",
    "default_timeout_seconds": 30,
    "max_response_body_bytes": 10485760,
    "allow_insecure_tls": true,
    "persist_request_history": false
  },
  "variables": [
    {
      "id": "dev",
      "base_url": "https://sample-dev.url/",
      "environment": "uat"
    },
    {
      "id": "uat",
      "base_url": "https://sample-uat.url",
      "environment": "uat"
    },
    {
      "id": "dev",
      "base_url": "https://sample-internal.url",
      "environment": "internal"
    },
  ],
  "oauth_profiles": [
    {
      "id": "microsoft-graph",
      "name": "Microsoft Graph Client Credentials",
      "type": "microsoft_client_credentials",
      "org_id_uuid": "00000000-0000-0000-0000-000000000000",
      "client_id": "11111111-1111-1111-1111-111111111111",
      "client_secret": "replace-with-secret",
      "scope": "https://graph.microsoft.com/.default",
      "token_url": "https://login.microsoftonline.com/org-id-uuid/oauth2/v2.0/token",
      "refresh_before_expiry_seconds": 60
    }
  ],
  "endpoint_groups": [
    {
      "id": "sample-service",
      "name": "Sample Service"
    }
  ],
  "endpoints": [
    {
      "id": "sample-api",
      "name": "Sample API",
      "description": "Example OAuth protected endpoint",
      "group_id": "sample-service",
      "enabled": true,
      "method": "GET",
      "url": "{{base_url}}/sample/api/endpoint",
      "timeout_seconds": 30,
      "auth": {
        "type": "oauth2",
        "profile_id": "microsoft-graph",
        "allow_authorization_header_override": false
      },
      "headers": [
        {
          "key": "Accept",
          "value": "application/json",
          "enabled": true
        }
      ],
      "query_parameters": [
        {
          "key": "environment",
          "value": "{{environment}}",
          "enabled": false
        }
      ],
      "body": {
        "type": "none",
        "content": ""
      }
    },
    {
      "id": "health-check",
      "name": "Health Check",
      "description": "Direct API request without OAuth",
      "group_id": "sample-service",
      "enabled": true,
      "method": "GET",
      "url": "{{base_url}}/health",
      "auth": {
        "type": "none"
      },
      "headers": [],
      "query_parameters": [],
      "body": {
        "type": "none",
        "content": ""
      }
    }
  ]
}
```

## 10.2 配置字段说明

### app

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| name | string | 否 | 应用显示名称 |
| title | string | 是 | 显示的标题名称 |
| default_timeout_seconds | integer | 否 | 默认请求超时 |
| max_response_body_bytes | integer | 否 | 最大内存响应体 |
| allow_insecure_tls | boolean | 否 | 是否允许跳过 TLS 校验 |
| persist_request_history | boolean | 否 | 是否持久化请求历史 |

### oauth_profiles

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | string | 是 | OAuth 配置唯一 ID |
| name | string | 是 | 显示名称 |
| type | string | 是 | 固定为 `microsoft_client_credentials` |
| org_id_uuid | string | 条件必填 | Microsoft Tenant / Organization ID |
| client_id | string | 是 | 应用 Client ID |
| client_secret | string | 是 | 应用 Client Secret |
| scope | string | 是 | OAuth Scope |
| token_url | string | 否 | 自定义完整 Token URL |
| refresh_before_expiry_seconds | integer | 否 | 提前刷新秒数 |

### endpoints

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | string | 是 | Endpoint 唯一 ID |
| name | string | 是 | Endpoint 名称 |
| description | string | 否 | 说明 |
| group_id | string | 否 | 分组 ID |
| enabled | boolean | 否 | 是否启用，默认 true |
| method | string | 是 | HTTP Method |
| url | string | 是 | 请求 URL |
| timeout_seconds | integer | 否 | 单独请求超时 |
| auth | object | 是 | 鉴权配置 |
| headers | array | 否 | 请求 Headers |
| query_parameters | array | 否 | Query 参数 |
| body | object | 否 | 请求 Body |

### auth

无鉴权：

```json
{
  "type": "none"
}
```

OAuth：

```json
{
  "type": "oauth2",
  "profile_id": "microsoft-graph",
  "allow_authorization_header_override": false
}
```

---

## 11. 配置校验规则

程序启动及手动重新加载时，必须执行配置校验。

### 11.1 基础校验

- Root 必须是 JSON Object。
- `endpoints` 必须是数组。
- Endpoint `id` 不得重复。
- OAuth Profile `id` 不得重复。
- `group_id` 必须引用存在的分组。
- `profile_id` 必须引用存在的 OAuth Profile。
- Method 必须在支持列表中。
- URL 不得为空。
- Timeout 必须大于 0。
- `max_response_body_bytes` 必须大于 0。

### 11.2 OAuth 校验

当 `type` 为 `microsoft_client_credentials`：

- `client_id` 必填。
- `client_secret` 必填。
- `scope` 必填。
- `token_url` 为空时，`org_id_uuid` 必填。
- `refresh_before_expiry_seconds` 不得小于 0。
- Token URL 必须为 HTTPS，除非显式允许开发环境例外。

### 11.3 安全校验

提供警告但不一定阻止启动：

- `client_secret` 看起来为占位值。
- 配置文件权限过于宽松。
- Endpoint 使用 HTTP。
- 开启跳过 TLS 校验。
- 请求 Header 中配置了疑似密码、Token 或 Secret。
- Client Secret 被重复用于多个 Profile。

---

## 12. 安全与隐私

## 12.1 敏感数据

敏感数据包括：

- `client_secret`
- `access_token`
- `Authorization`
- Cookie
- 可能包含个人信息的请求体或响应体

处理要求：

- Client Secret 在 UI 中默认掩码。
- 日志中 Client Secret 全量替换为 `******`。
- Access Token 日志中仅显示前 6 位和后 4 位，或完全隐藏。
- 复制 Curl 时默认将敏感值替换为占位符。
- Token 仅存内存。
- 应用退出时清空 Token。
- 崩溃日志不得包含完整请求和响应。

## 12.2 配置文件安全

由于 Client Secret 存放在 `config.json`，产品必须在文档中明确：

- 配置文件应限制操作系统文件权限。
- 配置文件不得提交到 Git。
- 建议使用部署工具或安全渠道下发。
- 生产环境建议后续支持环境变量或操作系统凭据存储。
- 首版可支持环境变量引用作为增强项：

```json
{
  "client_secret": "${ENV:MICROSOFT_CLIENT_SECRET}"
}
```

## 12.3 网络安全

- 默认校验 TLS 证书。
- 默认仅允许 HTTPS Token Endpoint。
- HTTP 请求禁止自动将 Authorization Header 重定向到不同 Host。
- 限制最大重定向次数。
- 跨域重定向时移除敏感 Header。
- 设置合理的连接、TLS 握手和整体请求超时。
- 这个程序不使用 Cookie

---

## 13. 技术架构建议

## 13.1 技术栈

### 后端

- Go。
- Wails。
- 标准库 `net/http`。
- `context.Context` 请求取消。
- JSON 使用 `encoding/json`。
- 日志可使用 `slog` 或结构化日志库。

### 前端

可选：

- React + TypeScript。
- Vue 3 + TypeScript。
- Svelte + TypeScript。

建议 React 或 Vue 3，以便实现组件化 Request Editor 和 Response Viewer。

### UI

- CSS Variables 实现主题。
- `backdrop-filter` 实现亚克力效果。
- Fluent UI System Icons 或 Lucide。
- Monaco Editor 或 CodeMirror 用于 JSON 编辑与高亮。

## 13.2 Go 模块划分 （建议）

建议目录：

```text
/
├─ main.go
├─ config.json
├─ internal/
│  ├─ app/
│  │  └─ app.go
│  ├─ config/
│  │  ├─ loader.go
│  │  ├─ model.go
│  │  └─ validator.go
│  ├─ httpclient/
│  │  ├─ client.go
│  │  ├─ request_builder.go
│  │  └─ response.go
│  ├─ oauth/
│  │  ├─ microsoft.go
│  │  ├─ cache.go
│  │  └─ model.go
│  ├─ security/
│  │  └─ redact.go
│  └─ history/
│     └─ store.go
├─ frontend/
│  ├─ src/
│  │  ├─ components/
│  │  ├─ views/
│  │  ├─ stores/
│  │  ├─ themes/
│  │  └─ types/
│  └─ package.json
└─ README.md
```

## 13.3 后端接口建议

Wails 暴露给前端的方法：

```go
LoadConfig() (*ConfigView, error)
ReloadConfig() (*ConfigView, error)
ValidateConfig() (*ValidationResult, error)
SendRequest(input RequestInput) (*ResponseOutput, error)
CancelRequest(requestID string) error
ClearTokenCache() error
GetTokenStatus(profileID string) (*TokenStatus, error)
GetAppInfo() (*AppInfo, error)
OpenConfigDirectory() error
OpenLogDirectory() error
```

说明：

- 前端不得直接读取 `client_secret`。
- `LoadConfig` 返回给前端时，OAuth Secret 应脱敏或不返回。
- OAuth 和 HTTP 请求必须由 Go 后端执行。
- 不应由 WebView 前端直接请求 Microsoft Token Endpoint 或内网 API。

---

## 14. 数据模型建议

### RequestInput

```go
type RequestInput struct {
    RequestID       string            `json:"requestId"`
    EndpointID      string            `json:"endpointId"`
    Method          string            `json:"method"`
    URL             string            `json:"url"`
    Headers         []KeyValueItem    `json:"headers"`
    QueryParameters []KeyValueItem    `json:"queryParameters"`
    BodyType        string            `json:"bodyType"`
    Body            string            `json:"body"`
    TimeoutSeconds  int               `json:"timeoutSeconds"`
    UseOAuth        bool              `json:"useOAuth"`
    OAuthProfileID  string            `json:"oauthProfileId"`
}
```

### ResponseOutput

```go
type ResponseOutput struct {
    RequestID          string              `json:"requestId"`
    StatusCode         int                 `json:"statusCode"`
    Status             string              `json:"status"`
    Headers            map[string][]string `json:"headers"`
    Body                string              `json:"body"`
    BodyTruncated       bool                `json:"bodyTruncated"`
    DurationMilliseconds int64              `json:"durationMilliseconds"`
    SizeBytes           int64               `json:"sizeBytes"`
    ContentType         string              `json:"contentType"`
    UsedOAuth           bool                `json:"usedOAuth"`
    TokenFromCache      bool                `json:"tokenFromCache"`
    ErrorCode           string              `json:"errorCode"`
    ErrorMessage        string              `json:"errorMessage"`
}
```

---

## 15. OAuth 实现伪代码

```go
func (s *OAuthService) GetAccessToken(
    ctx context.Context,
    profile OAuthProfile,
) (string, bool, error) {
    cacheKey := BuildCacheKey(profile.TokenURL, profile.ClientID, profile.Scope)

    if token, ok := s.cache.GetValid(cacheKey, profile.RefreshBeforeExpirySeconds); ok {
        return token.AccessToken, true, nil
    }

    token, err := s.singleflight.Do(cacheKey, func() (*CachedToken, error) {
        if token, ok := s.cache.GetValid(cacheKey, profile.RefreshBeforeExpirySeconds); ok {
            return token, nil
        }

        tokenURL := profile.TokenURL
        if tokenURL == "" {
            tokenURL = fmt.Sprintf(
                "https://login.microsoftonline.com/%s/oauth2/v2.0/token",
                url.PathEscape(profile.OrgIDUUID),
            )
        }

        form := url.Values{}
        form.Set("client_id", profile.ClientID)
        form.Set("scope", profile.Scope)
        form.Set("client_secret", profile.ClientSecret)
        form.Set("grant_type", "client_credentials")

        req, err := http.NewRequestWithContext(
            ctx,
            http.MethodPost,
            tokenURL,
            strings.NewReader(form.Encode()),
        )
        if err != nil {
            return nil, err
        }

        req.Header.Set(
            "Content-Type",
            "application/x-www-form-urlencoded",
        )

        resp, err := s.client.Do(req)
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()

        // 限制响应大小、检查状态码、解析 JSON、验证 access_token。
        // 将 token 和 expires_at 写入内存缓存。
        return parsedToken, nil
    })

    if err != nil {
        return "", false, err
    }

    return token.AccessToken, false, nil
}
```

---

## 16. 日志与可观测性

日志级别：

- Debug。
- Info。
- Warn。
- Error。

日志内容建议：

- 应用启动和版本。
- 配置加载成功或失败。
- Endpoint ID。
- 请求开始和结束。
- 状态码。
- 耗时。
- OAuth Token 是否命中缓存。
- Token 刷新成功或失败。
- 网络错误类型。

禁止记录：

- Client Secret。
- 完整 Access Token。
- 完整 Authorization Header。
- 未脱敏 Cookie。
- 默认情况下的完整 Request / Response Body。

---

## 17. 错误模型

建议统一错误结构：

```json
{
  "code": "OAUTH_TOKEN_REQUEST_FAILED",
  "message": "无法获取 Access Token",
  "details": "Microsoft token endpoint returned HTTP 401",
  "retryable": false
}
```

建议错误码：

- `CONFIG_FILE_NOT_FOUND`
- `CONFIG_READ_FAILED`
- `CONFIG_PARSE_FAILED`
- `CONFIG_VALIDATION_FAILED`
- `ENDPOINT_DISABLED`
- `INVALID_URL`
- `INVALID_METHOD`
- `INVALID_BODY`
- `REQUEST_TIMEOUT`
- `REQUEST_CANCELLED`
- `DNS_LOOKUP_FAILED`
- `TLS_HANDSHAKE_FAILED`
- `CONNECTION_FAILED`
- `RESPONSE_TOO_LARGE`
- `OAUTH_PROFILE_NOT_FOUND`
- `OAUTH_TOKEN_REQUEST_FAILED`
- `OAUTH_TOKEN_RESPONSE_INVALID`
- `OAUTH_ACCESS_TOKEN_MISSING`
- `API_REQUEST_FAILED`

---

## 18. 性能要求

- 启动至主界面可交互：目标小于 2 秒。
- 配置文件小于 1 MB 时加载时间：目标小于 200 ms。
- UI 操作无明显卡顿。
- HTTP 请求在 Go 后端异步执行，不阻塞前端。
- 响应体大于限制时避免占用过多内存。
- Token 缓存并发安全。
- 同一 OAuth Profile 并发刷新仅执行一次。
- 请求历史默认最多保留 100 条会话记录。

---

## 19. 兼容性要求

### Windows

- Windows 10 及以上。
- 使用 WebView2。
- 亚克力和系统主题优先适配 Windows。

### macOS

- 支持主流受支持版本。
- 提供相近的半透明材质效果。
- 不依赖 Windows 专属 API 才能正常使用。

### Linux

- 支持主流桌面发行版。
- 模糊效果不受支持时使用透明背景和渐变降级。
- 核心请求功能必须保持一致。

---

## 20. 可访问性

- 支持键盘导航。
- 主要操作提供快捷键。
- Light / Dark 模式保持足够对比度。
- 状态不可仅依靠颜色区分。
- 输入框、按钮和标签具备可读名称。
- 支持缩放。
- 动画应简短，并尊重系统减少动态效果设置。

建议快捷键：

| 快捷键 | 功能 |
|---|---|
| Ctrl / Cmd + Enter | 发送请求 |
| Esc | 取消请求 |
| Ctrl / Cmd + R | 重新加载配置 |
| Ctrl / Cmd + K | 搜索 Endpoint |
| Ctrl / Cmd + L | 聚焦 URL |
| Ctrl / Cmd + Shift + T | 切换主题菜单 |

---

## 21. 验收标准

### 配置加载

- 程序可从当前运行目录读取 `config.json`。
- 配置有效时正确展示 Endpoint。
- 配置无效时显示明确错误。
- 可手动重新加载配置。
- 新配置无效时不破坏当前可用状态。

### 直接请求

- 可选择 Endpoint 并发送请求。
- 可编辑 URL、Method、Headers、Query、Body。
- 可取消正在进行的请求。
- 可展示状态码、耗时、Headers 和 Body。
- 请求超时能正确提示。

### OAuth

- 可使用配置中的 `org_id_uuid`、`client_id`、`client_secret`、`scope` 获取 Token。
- Token 请求使用 `application/x-www-form-urlencoded`。
- 可解析 `access_token` 和 `expires_in`。
- 可自动添加 `Authorization: Bearer <token>`。
- Token 未过期时可复用。
- Token 到期前自动刷新。
- 获取 Token 失败时不发送业务 API 请求。
- 日志和 UI 不暴露完整 Secret 和 Token。

### UI

- 具备半透明、模糊、柔和高光的亚克力视觉效果。
- 卡片、按钮、输入框风格统一。
- 使用 Fluent 蓝作为主要强调色。
- 顶部菜单提供 `light`、`dark`、`system` 切换。
- `system` 可跟随操作系统主题。
- 模糊不受支持时有合理降级。

### 安全

- Token 不写入磁盘。
- Secret 不在普通前端数据中明文暴露。
- 导出 Curl 时默认脱敏。
- 跨 Host 重定向不携带 Authorization。
- 默认校验 TLS 证书。

---

## 22. 里程碑建议

### M1：基础框架

- Wails 项目初始化。
- Go 配置加载。
- 基础三栏布局。
- Endpoint 列表。
- 主题系统。

### M2：HTTP 调试

- 请求编辑器。
- Go HTTP Client。
- 响应查看器。
- 请求取消。
- 错误模型。

### M3：OAuth

- Microsoft Client Credentials。
- Token 缓存。
- Header 注入。
- Token 状态展示。
- 脱敏日志。

### M4：体验与稳定性

- 亚克力视觉完善。
- 历史记录。
- Curl 预览。
- 配置校验工具。
- 大响应保护。
- 跨平台测试。

### M5：发布

- Windows 安装包。
- macOS / Linux 构建验证。
- README。
- 配置示例。
- 安全说明。
- 自动化构建。

---

## 23. 后续版本候选

- 环境变量和系统凭据管理器读取 Secret。
- 多环境切换。
- Endpoint Collection 导入导出。
- 请求变量提取。
- JSONPath 断言。
- 请求前后脚本。
- OAuth Device Code。
- mTLS。
- 客户端证书。
- NTLM / Kerberos。
- 代理配置。
- gRPC。
- GraphQL。
- WebSocket。
- 命令行模式。
- 自动化测试集合。
- 报告导出。

---

## 24. 确认事项

1. `config.json` 中明文保存 `client_secret` 存在安全风险，这个版本确定采用铭文保存。
2. Microsoft OAuth Scope 不一定总是 Microsoft Graph，应允许每个 OAuth Profile 自定义。
3. 部分 Linux WebView 对 `backdrop-filter` 支持有限，必须设计降级方案。
4. 当前工作目录可能因快捷方式、安装器或启动脚本而变化，需要在 UI 中明确显示实际读取路径。
5. 是否允许用户在运行时修改并保存 `config.json`，首版确定只读，不允许修改。
6. 这个版本不需要需要持久化请求历史。
7. 允许跳过 TLS 校验。
8. 允许业务 API 请求跟随重定向，需要设置安全策略，并在界面中告诉用户。
