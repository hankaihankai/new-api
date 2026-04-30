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
