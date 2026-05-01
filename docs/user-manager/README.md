# 外部用户管理接口

该目录记录外部项目调用用户管理接口的方式。当前接口用于创建普通用户，并同时创建一个可直接使用的默认令牌。

## 配置

在服务端环境变量中配置固定授权码：

```env
USER_MANAGER_AUTH_KEY=change-me-to-a-long-random-secret
```

未配置 `USER_MANAGER_AUTH_KEY` 时，用户管理接口会拒绝访问。生产环境应使用足够长的随机字符串，并只交给可信调用方。

## 鉴权

接口支持以下任意一种 header：

```http
Authorization: Bearer <USER_MANAGER_AUTH_KEY>
```

或：

```http
X-User-Manager-Key: <USER_MANAGER_AUTH_KEY>
```

## 创建用户

```http
POST /api/user-manager/users
Content-Type: application/json
Authorization: Bearer <USER_MANAGER_AUTH_KEY>
```

请求体：

```json
{
  "username": "testuser",
  "password": "password123",
  "display_name": "testuser"
}
```

字段说明：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `username` | 是 | 用户名，最大 20 个字符 |
| `password` | 是 | 密码，8 到 20 个字符 |
| `display_name` | 否 | 显示名称，不填时使用用户名 |

成功响应：

```json
{
  "success": true,
  "message": "",
  "data": {
    "user_id": 123,
    "username": "testuser",
    "token_name": "default",
    "token_key": "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  }
}
```

## 默认规则

- 创建的用户固定为普通用户。
- 用户分组固定为 `default`。
- 默认令牌名称固定为 `default`。
- 默认令牌分组固定为 `default`。
- 默认令牌不过期。
- 默认令牌不限制模型。
- 默认令牌为无限额度。
- 响应中的 `token_key` 是完整 `sk-` 令牌，可直接用于模型转发接口。

## 查询用户统计

```http
GET /api/user-manager/users/{user_id}/stats
Authorization: Bearer <USER_MANAGER_AUTH_KEY>
```

可选 query 参数：

| 参数 | 说明 |
| --- | --- |
| `start_timestamp` | 统计开始 Unix 时间戳，秒 |
| `end_timestamp` | 统计结束 Unix 时间戳，秒 |

成功响应：

```json
{
  "success": true,
  "message": "",
  "data": {
    "user_id": 123,
    "username": "testuser",
    "account_data": {
      "current_balance": 0,
      "historical_consumption": 0,
      "quota": 99999705,
      "used_quota": 295,
      "total_quota": 100000000,
      "current_balance_amount": 199.99941,
      "used_quota_amount": 0.00059,
      "total_quota_amount": 200,
      "quota_per_unit": 500000
    },
    "usage_stats": {
      "request_count": 18,
      "stat_count": 18
    },
    "resource_consumption": {
      "stat_quota": 295,
      "stat_tokens": 12345
    },
    "performance_metrics": {
      "avg_rpm": 0,
      "avg_tpm": 0
    }
  }
}
```

字段对应关系：

| 截图模块 | 字段 |
| --- | --- |
| 账户数据 / 当前余额 | `account_data.current_balance_amount` |
| 账户数据 / 历史消耗 | `account_data.used_quota_amount` |
| 使用统计 / 请求次数 | `usage_stats.request_count` |
| 使用统计 / 统计次数 | `usage_stats.stat_count` |
| 资源消耗 / 统计额度 | `resource_consumption.stat_quota` |
| 资源消耗 / 统计Tokens | `resource_consumption.stat_tokens` |
| 性能指标 / 平均RPM | `performance_metrics.avg_rpm` |
| 性能指标 / 平均TPM | `performance_metrics.avg_tpm` |

额度说明：

- `quota`、`used_quota`、`total_quota`、`current_balance`、`historical_consumption` 都是数据库中的原始额度值。
- 金额换算公式为 `金额 = 原始额度 / quota_per_unit`，默认 `quota_per_unit = 500000`。
- `current_balance_amount`、`used_quota_amount`、`total_quota_amount` 是后端按上述公式换算后的数值，不包含 `$` 或其他货币符号。
- `stat_count`、`stat_quota`、`stat_tokens` 来自 `quota_data` 表，其中 `stat_tokens` 使用 `quota_data.token_used` 汇总。
- `avg_rpm` 和 `avg_tpm` 仅在同时传入有效 `start_timestamp`、`end_timestamp` 时按该时间段计算。

## 查询用户额度调用记录

```http
GET /api/user-manager/users/{user_id}/quota/records
Authorization: Bearer <USER_MANAGER_AUTH_KEY>
```

分页参数：

| 参数 | 说明 |
| --- | --- |
| `p` | 页码，默认 1 |
| `page_size` | 每页数量，默认使用系统分页配置，最大 100 |

过滤参数：

| 参数 | 说明 |
| --- | --- |
| `start_timestamp` | 开始 Unix 时间戳，秒 |
| `end_timestamp` | 结束 Unix 时间戳，秒 |
| `model_name` | 模型名称，精确匹配 |

成功响应：

```json
{
  "success": true,
  "message": "",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 0,
    "items": []
  }
}
```

`items` 内字段来自 `quota_data` 表，结构如下：

| 字段 | 说明 |
| --- | --- |
| `id` | 记录 ID |
| `user_id` | 用户 ID |
| `username` | 用户名 |
| `model_name` | 模型名称 |
| `created_at` | 记录时间，Unix 时间戳，秒，精确到小时 |
| `token_used` | 消耗 Token 数 |
| `count` | 请求次数 |
| `quota` | 消耗额度 |

## 设置用户额度

```http
POST /api/user-manager/users/{user_id}/quota
Content-Type: application/json
Authorization: Bearer <USER_MANAGER_AUTH_KEY>
```

请求体：

```json
{
  "mode": "add",
  "value": 500000
}
```

字段说明：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `mode` | 是 | 操作模式：`add`（增加）、`subtract`（减少）、`override`（覆盖） |
| `value` | 是 | 额度值；`add`/`subtract` 时必须大于 0 |

成功响应：

```json
{
  "success": true,
  "message": ""
}
```

行为说明：

- `add`：在用户当前额度基础上增加 `value`。
- `subtract`：在用户当前额度基础上减少 `value`；若用户额度不足，会返回错误。
- `override`：直接将用户额度覆盖为 `value`，不依赖当前额度。
- 每次操作都会写入管理日志，标记来源为 `user_manager`。

## 查询用户可用模型

```http
GET /api/user-manager/users/{user_id}/models
Authorization: Bearer <USER_MANAGER_AUTH_KEY>
```

成功响应：

```json
{
  "success": true,
  "message": "",
  "data": [
    "gpt-4o",
    "claude-3-5-sonnet"
  ]
}
```

行为说明：

- 该接口复用 `/api/user/models` 的模型可用性逻辑。
- 后端先读取用户分组，再计算该用户可使用的分组列表。
- 返回所有可用分组中已启用的模型名称，并自动去重。

## 错误场景

| 场景 | 结果 |
| --- | --- |
| 未配置 `USER_MANAGER_AUTH_KEY` | 返回 `success=false`，提示接口未配置授权码 |
| 未携带授权码或授权码错误 | 返回 `success=false`，提示授权码无效 |
| 用户名或密码为空 | 返回参数错误 |
| 用户名已存在或已删除 | 返回用户已存在 |
| 数据库写入失败 | 用户和默认令牌都会回滚，不会只创建其中一个 |

## curl 示例

```bash
curl -X POST "http://127.0.0.1:3000/api/user-manager/users" \
  -H "Authorization: Bearer ${USER_MANAGER_AUTH_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "display_name": "testuser"
  }'
```
