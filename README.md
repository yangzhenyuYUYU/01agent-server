# Gin Web Project

一个基于Gin框架的企业级Web服务项目，包含完整的用户认证、数据库操作、Redis缓存、日志管理等功能。

## 🚀 特性

- **Web框架**: Gin - 高性能的Go web框架
- **数据库**: 支持MySQL和PostgreSQL，使用GORM ORM
- **缓存**: Redis支持
- **认证**: JWT token认证
- **日志**: 结构化日志，支持文件和控制台输出
- **配置**: 基于YAML的配置管理
- **中间件**: CORS、日志、JWT认证等
- **API**: RESTful API设计
- **Docker**: 开发环境容器化

## 📁 项目结构

```
gin_web/
├── cmd/                    # 主程序入口
├── configs/               # 配置文件
│   └── config.yaml
├── internal/              # 内部应用代码
│   ├── config/           # 配置管理
│   ├── database/         # 数据库连接和迁移
│   ├── handler/          # HTTP处理器
│   ├── logger/           # 日志管理
│   ├── middleware/       # 中间件
│   ├── models/           # 数据模型
│   ├── redis/            # Redis连接
│   ├── router/           # 路由配置
│   ├── service/          # 业务逻辑
│   └── utils/            # 工具函数
├── logs/                 # 日志文件目录
├── docker-compose.yml    # Docker开发环境
├── go.mod               # Go模块依赖
├── go.sum
├── main.go              # 程序入口
└── README.md
```

## 🛠️ 安装和运行

### 前置要求

- Go 1.21+
- MySQL 8.0+ 或 PostgreSQL 15+
- Redis 6.0+

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd gin_web
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 启动开发环境（推荐）

使用Docker Compose启动所有依赖服务：

```bash
docker-compose up -d
```

这将启动：
- MySQL (端口: 3306)
- Redis (端口: 6379)
- PostgreSQL (端口: 5432)
- Redis Commander Web UI (端口: 8081)

### 4. 配置数据库

项目默认使用MySQL，配置文件在 `configs/config.yaml`。
如需使用PostgreSQL，请修改 `internal/database/database.go` 中的数据库初始化代码。

### 5. 运行项目

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动

## 📚 API文档

### 基础接口

- `GET /` - 欢迎页面
- `GET /api/v1/hello` - Hello接口
- `GET /health` - 健康检查

### 用户认证

- `POST /api/v1/register` - 用户注册
- `POST /api/v1/login` - 用户登录
- `POST /api/v1/logout` - 用户登出（需认证）

### 用户管理

- `GET /api/v1/users` - 获取用户列表（分页）
- `GET /api/v1/users/:id` - 获取指定用户信息
- `GET /api/v1/profile` - 获取当前用户信息（需认证）
- `PUT /api/v1/profile` - 更新当前用户信息（需认证）

### 用户参数

- `GET /api/v1/user/parameters` - 获取用户参数配置（需认证）
- `PUT /api/v1/user/parameters` - 更新用户参数配置（需认证）

### 会话管理

- `GET /api/v1/user/sessions` - 获取用户活跃会话（需认证）

### 配置管理

- `GET /api/v1/config` - 获取系统配置信息（隐藏敏感信息）
- `GET /api/v1/config/themes` - 获取主题配置
- `GET /api/v1/config/credits` - 获取积分配置

### 并发测试

- `GET /api/v1/concurrent/serial?tasks=5` - 串行执行任务测试
- `GET /api/v1/concurrent/parallel?tasks=5` - 并发执行任务测试
- `GET /api/v1/concurrent/compare?tasks=5` - 串行vs并发对比测试
- `GET /api/v1/concurrent/stress?goroutines=100` - 并发压力测试

### API示例

#### 用户注册
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "nickname": "测试用户"
  }'
```

#### 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

#### 获取用户信息（需要token）
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 更新用户信息
```bash
curl -X PUT http://localhost:8080/api/v1/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "nickname": "新昵称",
    "phone": "13888888888"
  }'
```

#### 获取用户参数
```bash
curl -X GET http://localhost:8080/api/v1/user/parameters \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 用户登出
```bash
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 获取系统配置
```bash
curl -X GET http://localhost:8080/api/v1/config
```

#### 获取主题配置
```bash
curl -X GET http://localhost:8080/api/v1/config/themes
```

#### 获取积分配置
```bash
curl -X GET http://localhost:8080/api/v1/config/credits
```

#### 串行执行测试（执行5个任务）
```bash
curl -X GET "http://localhost:8080/api/v1/concurrent/serial?tasks=5"
```

#### 并发执行测试（执行5个任务）
```bash
curl -X GET "http://localhost:8080/api/v1/concurrent/parallel?tasks=5"
```

#### 串行vs并发对比测试
```bash
curl -X GET "http://localhost:8080/api/v1/concurrent/compare?tasks=10"
```

#### 并发压力测试（100个协程）
```bash
curl -X GET "http://localhost:8080/api/v1/concurrent/stress?goroutines=100"
```

## ⚙️ 配置说明

主要配置文件位于 `configs/config.yaml`：

### 基础配置
- `server`: 服务器配置（端口、模式）
- `database`: 数据库配置（MySQL/PostgreSQL）
- `redis`: Redis配置
- `jwt`: JWT token配置
- `log`: 日志配置
- `cors`: CORS跨域配置

### 第三方服务配置
- `doubao`: 豆包AI配置（API密钥、模型配置）
- `wxpay`: 微信支付配置
- `alipay`: 支付宝支付配置
- `wechat`: 微信小程序配置
- `wechatGzh`: 微信公众号配置
- `wechatPlatform`: 微信三方平台配置
- `sms`: 短信服务配置
- `oss`: 阿里云OSS存储配置
- `email`: 邮件服务配置

### 业务配置
- `bp`: BP文档配置
- `credits`: 积分系统配置
- `verifyCode`: 验证码配置
- `themes`: 主题映射配置

## 🔧 开发

### 添加新的API

1. 在 `internal/models/` 中定义数据模型
2. 在 `internal/service/` 中实现业务逻辑
3. 在 `internal/handler/` 中添加HTTP处理器
4. 在 `internal/router/` 中注册路由

### 数据库迁移

项目启动时会自动执行数据库迁移。要添加新的表：

1. 在 `internal/models/` 中定义模型
2. 在 `internal/database/migrate.go` 中添加模型到AutoMigrate列表

### 中间件

项目包含以下中间件：
- JWT认证中间件
- CORS跨域中间件
- 日志中间件
- 恢复中间件（错误恢复）

## 📝 日志

日志配置支持：
- 多种日志级别：debug, info, warn, error
- 多种输出格式：json, text
- 多种输出方式：console, file, both
- 日志轮转和压缩

日志文件默认保存在 `logs/` 目录下。

## 🐳 Docker

### 开发环境

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 生产环境

可以创建Dockerfile来构建应用镜像：

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
CMD ["./main"]
```

## 🤝 贡献

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 🆘 支持

如有问题，请提交 [Issue](https://github.com/your-username/gin_web/issues) 或联系维护者。 