分布式短链接系统

## 📁 项目结构phase 1

```
shorturl-system/
├── docker-compose.yml          # Docker编排文件
├── scripts/
│   └── init.sql               # 数据库初始化脚本
└── go-services/
    ├── shortener-service/     # 短链生成服务 (端口: 8001)
    ├── redirect-service/      # 重定向服务 (端口: 8002)
    ├── gateway/               # API网关 (待实现)
    └── analytics-service/     # 数据分析服务 (待实现)
```

### 第一步：启动基础设施

在项目根目录下创建 `docker-compose.yml` 和 `scripts/init.sql` 文件，然后运行：

```bash
# 启动MySQL、Redis、Kafka
docker-compose up -d

# 查看容器状态
docker-compose ps

# 查看MySQL日志（确保数据库初始化完成）
docker-compose logs mysql
```

### 第二步：配置Shortener Service

```bash
cd go-services/shortener-service

# 初始化Go模块
go mod init shortener-service
go mod tidy

# 下载依赖
go get github.com/zeromicro/go-zero@latest
go get github.com/go-redis/redis/v8
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/bwmarrin/snowflake
```

### 第三步：启动Shortener Service

```bash
# 在 go-services/shortener-service 目录下
go run cmd/main.go

# 或者编译后运行
go build -o shortener cmd/main.go
./shortener
```

预期输出：
```
Starting server at 0.0.0.0:8001...
```

### 第四步：启动Redirect Service

```bash
cd go-services/redirect-service

# 初始化Go模块
go mod init redirect-service
go mod tidy

# 下载依赖
go get github.com/go-redis/redis/v8

# 运行服务
go run cmd/main.go
```

预期输出：
```
Redirect service starting on :8002...
```

## 🧪 测试API

### 1. 创建短链接

```bash
curl -X POST http://localhost:8001/api/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://www.google.com",
    "title": "Google搜索",
    "description": "世界上最受欢迎的搜索引擎"
  }'
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "short_code": "aBc123",
    "short_url": "http://localhost:8002/aBc123",
    "original_url": "https://www.google.com",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

### 2. 查询短链详情

```bash
curl http://localhost:8001/api/links/aBc123
```

### 3. 批量创建短链接

```bash
curl -X POST http://localhost:8001/api/batch/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://www.github.com",
      "https://www.stackoverflow.com",
      "https://www.reddit.com"
    ]
  }'
```

### 4. 测试重定向

在浏览器中访问：
```
http://localhost:8002/aBc123
```

应该会重定向到原始URL。

## 📊 数据库查看

```bash
# 连接到MySQL
docker exec -it shorturl_mysql mysql -uroot -proot123

# 切换数据库
use shorturl;

# 查看短链接表
select * from short_links;

# 查看访问记录
select * from visit_logs;
```

## 🔧 常见问题

### 1. 端口被占用

修改配置文件中的端口：
- `shortener-service/internal/config/config.yaml` 中的 `Port`
- `redirect-service/cmd/main.go` 中的 `:8002`

### 2. Redis连接失败

确保Docker容器正在运行：
```bash
docker-compose ps
docker-compose logs redis
```

### 3. MySQL连接失败

检查数据库是否初始化完成：
```bash
docker-compose logs mysql | grep "ready for connections"
```
## 📝 下一步开发计划

- [ ] **阶段三**：完善Redirect Service，添加访问统计
- [ ] **阶段四**：实现Gateway服务，统一API入口
- [ ] **阶段五**：实现Analytics Service，接入Kafka消费访问日志
- [ ] **阶段六**：开发Vue前端管理界面

**当前完成度**：✅ 阶段一 & 阶段二完成
