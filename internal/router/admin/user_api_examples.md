# 分销商管理接口请求示例

所有接口都需要在请求头中携带管理员认证 Token：
```
Authorization: Bearer <your_admin_token>
```

---

## 1. 设置分销商身份

### 接口信息
- **方法**: `POST`
- **路径**: `/api/v1/admin/user/{id}/distributor`
- **说明**: 将指定用户设置为分销商，如果已是分销商则更新信息

### cURL 示例

```bash
# 使用默认佣金比例（0.2）
curl -X POST "http://localhost:8080/api/v1/admin/user/user123/distributor" \
  -H "Authorization: Bearer your_admin_token" \
  -H "Content-Type: application/json"

# 指定佣金比例
curl -X POST "http://localhost:8080/api/v1/admin/user/user123/distributor" \
  -H "Authorization: Bearer your_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "commission_rate": 0.25
  }'

# 完整参数（佣金比例 + 额外参数）
curl -X POST "http://localhost:8080/api/v1/admin/user/user123/distributor" \
  -H "Authorization: Bearer your_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "commission_rate": 0.3,
    "extra_params": "{\"region\":\"华东\",\"contact\":\"张三\",\"phone\":\"13800138000\"}"
  }'
```

### JavaScript/Fetch 示例

```javascript
// 设置分销商（使用默认佣金比例）
async function setDistributor(userId) {
  const response = await fetch(`http://localhost:8080/api/v1/admin/user/${userId}/distributor`, {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer your_admin_token',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({})
  });
  
  const data = await response.json();
  console.log(data);
}

// 设置分销商（指定佣金比例）
async function setDistributorWithRate(userId, commissionRate) {
  const response = await fetch(`http://localhost:8080/api/v1/admin/user/${userId}/distributor`, {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer your_admin_token',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      commission_rate: commissionRate,
      extra_params: JSON.stringify({
        region: '华东',
        contact: '张三',
        phone: '13800138000'
      })
    })
  });
  
  const data = await response.json();
  console.log(data);
}

// 调用示例
setDistributor('user123');
setDistributorWithRate('user123', 0.3);
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 路径参数，用户ID |
| `commission_rate` | float | 否 | 佣金比例，范围 0-1，默认 0.2 |
| `extra_params` | string | 否 | 额外参数，JSON字符串格式 |

### 成功响应示例

```json
{
  "code": 200,
  "msg": "设置分销商身份成功",
  "data": {
    "distributor_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "user123",
    "commission_rate": 0.3,
    "extra_params": "{\"region\":\"华东\",\"contact\":\"张三\",\"phone\":\"13800138000\"}"
  }
}
```

### 错误响应示例

```json
{
  "code": 400,
  "msg": "佣金比例必须在0到1之间"
}
```

---

## 2. 移除分销商身份

### 接口信息
- **方法**: `DELETE`
- **路径**: `/api/v1/admin/user/{id}/distributor`
- **说明**: 移除指定用户的分销商身份，删除分销商记录并恢复用户角色为普通用户

### cURL 示例

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/user/user123/distributor" \
  -H "Authorization: Bearer your_admin_token"
```

### JavaScript/Fetch 示例

```javascript
async function removeDistributor(userId) {
  const response = await fetch(`http://localhost:8080/api/v1/admin/user/${userId}/distributor`, {
    method: 'DELETE',
    headers: {
      'Authorization': 'Bearer your_admin_token'
    }
  });
  
  const data = await response.json();
  console.log(data);
}

// 调用示例
removeDistributor('user123');
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 路径参数，用户ID |

### 成功响应示例

```json
{
  "code": 200,
  "msg": "移除分销商身份成功",
  "data": {
    "user_id": "user123"
  }
}
```

### 错误响应示例

```json
{
  "code": 404,
  "msg": "该用户不是分销商"
}
```

---

## 3. 获取分销商信息

### 接口信息
- **方法**: `GET`
- **路径**: `/api/v1/admin/user/{id}/distributor`
- **说明**: 获取指定用户的分销商详细信息

### cURL 示例

```bash
curl -X GET "http://localhost:8080/api/v1/admin/user/user123/distributor" \
  -H "Authorization: Bearer your_admin_token"
```

### JavaScript/Fetch 示例

```javascript
async function getDistributorInfo(userId) {
  const response = await fetch(`http://localhost:8080/api/v1/admin/user/${userId}/distributor`, {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer your_admin_token'
    }
  });
  
  const data = await response.json();
  console.log(data);
}

// 调用示例
getDistributorInfo('user123');
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 路径参数，用户ID |

### 成功响应示例

```json
{
  "code": 200,
  "msg": "获取分销商信息成功",
  "data": {
    "distributor_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "user123",
    "commission_rate": 0.3,
    "extra_params": "{\"region\":\"华东\",\"contact\":\"张三\",\"phone\":\"13800138000\"}",
    "created_at": "2025-01-15 10:30:00",
    "updated_at": "2025-01-15 14:20:00"
  }
}
```

### 错误响应示例

```json
{
  "code": 404,
  "msg": "该用户不是分销商"
}
```

---

## 4. 获取分销商列表

### 接口信息
- **方法**: `GET`
- **路径**: `/api/v1/admin/distributor/list`
- **说明**: 获取所有分销商列表，支持分页查询

### cURL 示例

```bash
# 使用默认分页（第1页，每页10条）
curl -X GET "http://localhost:8080/api/v1/admin/distributor/list" \
  -H "Authorization: Bearer your_admin_token"

# 指定分页参数
curl -X GET "http://localhost:8080/api/v1/admin/distributor/list?page=1&page_size=20" \
  -H "Authorization: Bearer your_admin_token"
```

### JavaScript/Fetch 示例

```javascript
async function getDistributorList(page = 1, pageSize = 10) {
  const url = new URL('http://localhost:8080/api/v1/admin/distributor/list');
  url.searchParams.append('page', page);
  url.searchParams.append('page_size', pageSize);
  
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer your_admin_token'
    }
  });
  
  const data = await response.json();
  console.log(data);
}

// 调用示例
getDistributorList(1, 20);
```

### 请求参数（Query）

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码，默认 1 |
| `page_size` | int | 否 | 每页数量，默认 10，最大 100 |

### 成功响应示例

```json
{
  "code": 200,
  "msg": "获取分销商列表成功",
  "data": {
    "items": [
      {
        "distributor_id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "user123",
        "commission_rate": 0.3,
        "extra_params": "{\"region\":\"华东\",\"contact\":\"张三\"}",
        "created_at": "2025-01-15 10:30:00",
        "updated_at": "2025-01-15 14:20:00",
        "username": "zhangsan",
        "nickname": "张三",
        "phone": "13800138000",
        "email": "zhangsan@example.com",
        "avatar": "https://example.com/avatar.jpg"
      },
      {
        "distributor_id": "660e8400-e29b-41d4-a716-446655440001",
        "user_id": "user456",
        "commission_rate": 0.25,
        "extra_params": null,
        "created_at": "2025-01-14 09:15:00",
        "updated_at": "2025-01-14 09:15:00",
        "username": "lisi",
        "nickname": "李四",
        "phone": "13900139000",
        "email": "lisi@example.com",
        "avatar": null
      }
    ],
    "total": 25,
    "page": 1,
    "page_size": 10
  }
}
```

---

## Postman 集合示例

### 1. 设置分销商身份

```
POST http://localhost:8080/api/v1/admin/user/{{id}}/distributor

Headers:
  Authorization: Bearer {{admin_token}}
  Content-Type: application/json

Body (raw JSON):
{
  "commission_rate": 0.3,
  "extra_params": "{\"region\":\"华东\",\"contact\":\"张三\"}"
}
```

### 2. 移除分销商身份

```
DELETE http://localhost:8080/api/v1/admin/user/{{id}}/distributor

Headers:
  Authorization: Bearer {{admin_token}}
```

### 3. 获取分销商信息

```
GET http://localhost:8080/api/v1/admin/user/{{id}}/distributor

Headers:
  Authorization: Bearer {{admin_token}}
```

### 4. 获取分销商列表

```
GET http://localhost:8080/api/v1/admin/distributor/list?page=1&page_size=10

Headers:
  Authorization: Bearer {{admin_token}}
```

---

## Python 请求示例

```python
import requests

BASE_URL = "http://localhost:8080/api/v1/admin"
ADMIN_TOKEN = "your_admin_token"
HEADERS = {
    "Authorization": f"Bearer {ADMIN_TOKEN}",
    "Content-Type": "application/json"
}

# 1. 设置分销商身份
def set_distributor(user_id, commission_rate=0.2, extra_params=None):
    url = f"{BASE_URL}/user/{user_id}/distributor"
    data = {
        "commission_rate": commission_rate
    }
    if extra_params:
        import json
        data["extra_params"] = json.dumps(extra_params)
    
    response = requests.post(url, headers=HEADERS, json=data)
    return response.json()

# 2. 移除分销商身份
def remove_distributor(user_id):
    url = f"{BASE_URL}/user/{user_id}/distributor"
    response = requests.delete(url, headers=HEADERS)
    return response.json()

# 3. 获取分销商信息
def get_distributor_info(user_id):
    url = f"{BASE_URL}/user/{user_id}/distributor"
    response = requests.get(url, headers=HEADERS)
    return response.json()

# 4. 获取分销商列表
def get_distributor_list(page=1, page_size=10):
    url = f"{BASE_URL}/distributor/list"
    params = {"page": page, "page_size": page_size}
    response = requests.get(url, headers=HEADERS, params=params)
    return response.json()

# 使用示例
if __name__ == "__main__":
    # 设置分销商
    result = set_distributor(
        user_id="user123",
        commission_rate=0.3,
        extra_params={"region": "华东", "contact": "张三"}
    )
    print(result)
    
    # 获取分销商信息
    info = get_distributor_info("user123")
    print(info)
    
    # 获取分销商列表
    distributor_list = get_distributor_list(page=1, page_size=20)
    print(distributor_list)
    
    # 移除分销商
    remove_result = remove_distributor("user123")
    print(remove_result)
```

---

## 注意事项

1. **认证**: 所有接口都需要管理员权限，必须在请求头中携带有效的管理员 Token
2. **佣金比例**: 必须在 0 到 1 之间（例如：0.2 表示 20%）
3. **额外参数**: `extra_params` 必须是有效的 JSON 字符串格式
4. **用户角色**: 设置分销商时，用户角色会自动更新为 4（分销商）；移除时恢复为 1（普通用户）
5. **事务处理**: 设置和移除操作都使用数据库事务，确保数据一致性

