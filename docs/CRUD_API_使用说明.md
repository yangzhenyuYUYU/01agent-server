# CRUD API 使用说明

## 概述

本项目已实现通用的 CRUD（增删改查）接口系统，基于 `internal/tools/crud.go` 提供。所有管理员接口都需要 JWT 认证和管理员权限。

## 基础路径

所有管理员接口的基础路径：`/admin`

## 已配置的 CRUD 接口

### 1. 用户管理 (User)
- **基础路径**: `/admin/user`
- **主键**: `user_id`
- **可搜索字段**: `username`, `phone`, `email`, `nickname`
- **默认排序**: `created_at DESC`

### 2. 用户会话管理 (UserSession)
- **基础路径**: `/admin/user_session`
- **主键**: `id`
- **可搜索字段**: `user_id`
- **默认排序**: `created_at DESC`

### 3. 用户参数管理 (UserParameters)
- **基础路径**: `/admin/user_parameter`
- **主键**: `param_id`
- **可搜索字段**: `user_id`
- **默认排序**: `created_time DESC`

### 4. 邀请码管理 (InvitationCode)
- **基础路径**: `/admin/invitation_code`
- **主键**: `id`
- **可搜索字段**: `code`
- **默认排序**: `created_at DESC`

### 5. 邀请关系管理 (InvitationRelation)
- **基础路径**: `/admin/invitation_relation`
- **主键**: `id`
- **可搜索字段**: `code`
- **默认排序**: `created_at DESC`

### 6. 佣金记录管理 (CommissionRecord)
- **基础路径**: `/admin/commission_record`
- **主键**: `id`
- **可搜索字段**: `user_id`
- **默认排序**: `created_at DESC`

## 标准 CRUD 接口

每个资源都提供以下 5 个标准接口：

### 1. 列表查询 (List)
**接口**: `GET /admin/{resource}/list`

**请求参数** (Query Parameters):
- `page` (int, 可选): 页码，默认 1，最小值 1
- `page_size` (int, 可选): 每页数量，默认 10，最小值 1，最大值 9999
- `search` (string, 可选): 搜索关键词，会在配置的 `SearchFields` 中搜索
- `order_by` (string, 可选): 排序字段，默认使用配置的 `DefaultOrderBy`
- `order_direction` (string, 可选): 排序方向，`asc` 或 `desc`，默认 `desc`
- **动态筛选**: 除了上述系统参数外，任何其他查询参数都会尝试作为模型字段进行筛选

**响应格式**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

**示例**:
```bash
# 获取用户列表（第1页，每页10条）
GET /admin/user/list?page=1&page_size=10

# 搜索用户名包含 "test" 的用户
GET /admin/user/list?search=test

# 按邮箱筛选并排序
GET /admin/user/list?email=test@example.com&order_by=created_at&order_direction=asc

# 多条件筛选（动态筛选）
GET /admin/user/list?role=1&status=active&page=1&page_size=20
```

### 2. 详情查询 (Detail)
**接口**: `GET /admin/{resource}/:id`

**路径参数**:
- `id`: 资源的主键值（支持整数或字符串）

**响应格式**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    // 资源对象
  }
}
```

**示例**:
```bash
# 获取用户详情
GET /admin/user/123

# 获取邀请码详情
GET /admin/invitation_code/abc123
```

### 3. 创建 (Create)
**接口**: `POST /admin/{resource}`

**请求体**: JSON 格式，包含要创建的资源字段

**响应格式**:
```json
{
  "code": 0,
  "msg": "创建成功",
  "data": {
    // 创建后的资源对象
  }
}
```

**示例**:
```bash
POST /admin/user
Content-Type: application/json

{
  "username": "newuser",
  "email": "newuser@example.com",
  "phone": "13800138000",
  "nickname": "新用户"
}
```

### 4. 更新 (Update)
**接口**: `PUT /admin/{resource}/:id`

**路径参数**:
- `id`: 资源的主键值

**请求体**: JSON 格式，包含要更新的字段（部分更新）

**响应格式**:
```json
{
  "code": 0,
  "msg": "更新成功",
  "data": {
    // 更新后的资源对象
  }
}
```

**示例**:
```bash
PUT /admin/user/123
Content-Type: application/json

{
  "nickname": "更新后的昵称",
  "status": "active"
}
```

### 5. 删除 (Delete)
**接口**: `DELETE /admin/{resource}/:id`

**路径参数**:
- `id`: 资源的主键值

**响应格式**:
```json
{
  "code": 0,
  "msg": "删除成功"
}
```

**示例**:
```bash
DELETE /admin/user/123
```

## 其他接口

### 健康检查
**接口**: `GET /admin/health`

**说明**: 检查管理员 API 是否正常运行（不需要管理员权限）

**响应格式**:
```json
{
  "code": 0,
  "msg": "Admin API is running",
  "data": {
    "status": "ok"
  }
}
```

## 认证要求

所有管理员接口（除 `/admin/health` 外）都需要：

1. **JWT Token**: 在请求头中携带
   ```
   Authorization: Bearer <your_jwt_token>
   ```

2. **管理员权限**: 用户角色必须是管理员（`UserRoleAdmin = 0`）

## 错误响应格式

```json
{
  "code": 400,  // 错误码
  "msg": "错误信息"
}
```

常见错误码：
- `400`: 参数错误
- `401`: 未授权（未登录或 token 无效）
- `403`: 禁止访问（非管理员）
- `404`: 资源不存在
- `500`: 服务器内部错误

## 动态筛选功能

列表查询接口支持动态筛选，你可以通过查询参数直接筛选模型的任何字段：

**支持的字段类型**:
- `bool`: 自动转换为布尔值
- `int/int8/int16/int32/int64`: 自动转换为整数
- `float32/float64`: 自动转换为浮点数
- `string`: 使用 `LIKE` 模糊查询

**示例**:
```bash
# 筛选角色为 1 的用户
GET /admin/user/list?role=1

# 筛选状态为 active 的用户（字符串模糊匹配）
GET /admin/user/list?status=active

# 组合多个筛选条件
GET /admin/user/list?role=1&status=active&page=1&page_size=20
```

## 如何添加新的 CRUD 接口

在 `internal/router/admin.go` 的 `SetupAdminRoutes` 函数中添加：

```go
// 新资源管理 CRUD
newResourceCRUD := tools.NewCRUDHandler(tools.CRUDConfig{
    Model:          &models.NewResource{},  // 你的模型
    SearchFields:   []string{"field1", "field2"},  // 可搜索字段
    DefaultOrderBy: "created_at",  // 默认排序字段
    RequireAdmin:   true,  // 是否需要管理员权限
    PrimaryKey:     "id",  // 主键字段名
}, repository.DB)

newResourceGroup := admin.Group("/new_resource")
newResourceGroup.Use(middleware.AdminAuth())
{
    newResourceGroup.GET("/list", newResourceCRUD.List)
    newResourceGroup.GET("/:id", newResourceCRUD.Detail)
    newResourceGroup.POST("", newResourceCRUD.Create)
    newResourceGroup.PUT("/:id", newResourceCRUD.Update)
    newResourceGroup.DELETE("/:id", newResourceCRUD.Delete)
}
```

## 完整接口列表

| 资源 | 列表 | 详情 | 创建 | 更新 | 删除 |
|------|------|------|------|------|------|
| 用户 | `GET /admin/user/list` | `GET /admin/user/:id` | `POST /admin/user` | `PUT /admin/user/:id` | `DELETE /admin/user/:id` |
| 用户会话 | `GET /admin/user_session/list` | `GET /admin/user_session/:id` | `POST /admin/user_session` | `PUT /admin/user_session/:id` | `DELETE /admin/user_session/:id` |
| 用户参数 | `GET /admin/user_parameter/list` | `GET /admin/user_parameter/:id` | `POST /admin/user_parameter` | `PUT /admin/user_parameter/:id` | `DELETE /admin/user_parameter/:id` |
| 邀请码 | `GET /admin/invitation_code/list` | `GET /admin/invitation_code/:id` | `POST /admin/invitation_code` | `PUT /admin/invitation_code/:id` | `DELETE /admin/invitation_code/:id` |
| 邀请关系 | `GET /admin/invitation_relation/list` | `GET /admin/invitation_relation/:id` | `POST /admin/invitation_relation` | `PUT /admin/invitation_relation/:id` | `DELETE /admin/invitation_relation/:id` |
| 佣金记录 | `GET /admin/commission_record/list` | `GET /admin/commission_record/:id` | `POST /admin/commission_record` | `PUT /admin/commission_record/:id` | `DELETE /admin/commission_record/:id` |

## 测试示例

### 使用 curl 测试

```bash
# 1. 获取用户列表
curl -X GET "http://localhost:8080/admin/user/list?page=1&page_size=10" \
  -H "Authorization: Bearer <your_jwt_token>"

# 2. 搜索用户
curl -X GET "http://localhost:8080/admin/user/list?search=test" \
  -H "Authorization: Bearer <your_jwt_token>"

# 3. 获取用户详情
curl -X GET "http://localhost:8080/admin/user/123" \
  -H "Authorization: Bearer <your_jwt_token>"

# 4. 创建用户
curl -X POST "http://localhost:8080/admin/user" \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com"
  }'

# 5. 更新用户
curl -X PUT "http://localhost:8080/admin/user/123" \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "新昵称"
  }'

# 6. 删除用户
curl -X DELETE "http://localhost:8080/admin/user/123" \
  -H "Authorization: Bearer <your_jwt_token>"
```

### 使用 Postman 测试

1. 设置环境变量 `base_url` = `http://localhost:8080`
2. 设置环境变量 `jwt_token` = `<your_jwt_token>`
3. 在请求头中添加：`Authorization: Bearer {{jwt_token}}`
4. 按照上述接口列表创建请求

