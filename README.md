# NCCL Test Web Service

基于 Go 和 Gin 框架开发的 Web 服务，带有 Svelte 前端界面。

## 项目结构

```
.
├── cmd/
│   └── server/          # 主程序入口
│       └── main.go
├── internal/
│   ├── handlers/        # HTTP 处理器
│   │   ├── health.go    # 健康检查
│   │   └── iplist.go    # IP列表处理
│   └── web/             # 嵌入的前端文件
│       ├── embed.go     # 前端嵌入逻辑
│       └── dist/        # 构建后的前端文件
├── web/                 # Svelte 前端项目
│   ├── src/
│   │   ├── App.svelte   # 主组件
│   │   └── main.js      # 入口文件
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
├── data/                # 数据存储目录
│   └── iplist          # IP列表文件
├── build.sh             # 构建脚本
├── go.mod
└── go.sum
```

## 功能特性

- ✅ RESTful API 接口
- ✅ Svelte 响应式前端界面
- ✅ 前端嵌入到 Go 二进制文件中
- ✅ IP 列表的增删改查功能
- ✅ 批量编辑模式
- ✅ 现代化的 UI 设计

## API 接口

### 1. 健康检查

**请求:**
```
GET /healthz
```

**响应:**
```json
{
  "status": "ok"
}
```

### 2. IP列表管理接口

#### 2.1 保存/创建IP列表（增）

**请求:**
```
POST /iplist
Content-Type: application/json

{
  "iplist": [
    "192.168.1.1",
    "192.168.1.2",
    "192.168.1.3"
  ]
}
```

**响应:**
```json
{
  "message": "IP list saved successfully",
  "count": 3
}
```

#### 2.2 查询IP列表（查）

**请求:**
```
GET /iplist
```

**响应:**
```json
{
  "iplist": [
    "192.168.1.1",
    "192.168.1.2",
    "192.168.1.3"
  ],
  "count": 3
}
```

如果列表为空或文件不存在：
```json
{
  "iplist": [],
  "count": 0
}
```

#### 2.3 更新IP列表（改）

**请求:**
```
PUT /iplist
Content-Type: application/json

{
  "iplist": [
    "10.0.0.1",
    "10.0.0.2"
  ]
}
```

**响应:**
```json
{
  "message": "IP list updated successfully",
  "count": 2
}
```

#### 2.4 删除IP列表（删）

**请求:**
```
DELETE /iplist
```

**响应:**
```json
{
  "message": "IP list deleted successfully"
}
```

## 运行服务

### 方法一：使用构建脚本（推荐）

构建脚本会自动构建前端并将其嵌入到 Go 二进制文件中：

```bash
./build.sh
./nccl-test-web
```

### 方法二：手动构建

#### 1. 构建前端
```bash
cd web
npm install
npm run build
cd ..
```

#### 2. 复制前端构建文件
```bash
cp -r web/dist internal/web/
```

#### 3. 构建 Go 二进制
```bash
go build -o nccl-test-web cmd/server/main.go
```

#### 4. 运行
```bash
./nccl-test-web
```

### 方法三：开发模式

开发时可以分别运行前后端：

**后端：**
```bash
go run cmd/server/main.go
```

**前端（另一个终端）：**
```bash
cd web
npm run dev
```

服务将在以下地址启动：
- **后端 API**: `http://localhost:8080`
- **前端界面**: `http://localhost:8080` (生产模式) 或 `http://localhost:5173` (开发模式)

## 使用前端界面

访问 `http://localhost:8080` 即可使用图形化界面管理 IP 列表：

- **查看列表**：页面加载时自动显示所有 IP
- **添加 IP**：在输入框中输入 IP 地址，点击"Add IP"
- **删除 IP**：点击每个 IP 右侧的"Delete"按钮
- **批量编辑**：点击"Batch Edit"按钮，可以一次性编辑所有 IP（每行一个）
- **删除全部**：点击"Delete All"清空所有 IP
- **刷新**：点击"Refresh"重新加载列表

## 测试接口

### 测试健康检查
```bash
curl http://localhost:8080/healthz
```

### 测试创建/保存IP列表
```bash
curl -X POST http://localhost:8080/iplist \
  -H "Content-Type: application/json" \
  -d '{
    "iplist": [
      "192.168.1.1",
      "192.168.1.2",
      "192.168.1.3"
    ]
  }'
```

### 测试查询IP列表
```bash
curl http://localhost:8080/iplist
```

### 测试更新IP列表
```bash
curl -X PUT http://localhost:8080/iplist \
  -H "Content-Type: application/json" \
  -d '{
    "iplist": [
      "10.0.0.1",
      "10.0.0.2",
      "10.0.0.3",
      "10.0.0.4"
    ]
  }'
```

### 测试删除IP列表
```bash
curl -X DELETE http://localhost:8080/iplist
```

### 查看保存的IP列表文件
```bash
cat data/iplist
```

## 快速开始

使用 Makefile 快速构建和运行：

```bash
# 完整构建（前端 + 后端）
make build

# 运行服务
make run

# 或者一次完成
make all run

# 清理构建文件
make clean

# 开发模式 - 运行后端
make dev-backend

# 开发模式 - 运行前端（需要另一个终端）
make dev-frontend
```

## 技术栈

**后端：**
- Go 1.x
- Gin Web Framework
- embed 包（嵌入静态文件）

**前端：**
- Svelte 4
- Vite 5
- 原生 Fetch API

## 构建说明

项目使用 Go 的 `embed` 功能将前端文件嵌入到二进制文件中：
- 前端文件位于 `web/` 目录
- 构建后的文件会复制到 `internal/web/dist/`
- Go 在编译时会将 `internal/web/dist/` 目录嵌入到二进制文件
- 最终生成的单个二进制文件包含了完整的前后端功能

这意味着：
- ✅ 只需要一个二进制文件即可运行整个应用
- ✅ 不需要额外部署前端静态文件
- ✅ 方便分发和部署

## Docker 部署

使用 Docker 快速部署：

```bash
# 使用 Docker Compose
docker-compose up -d

# 或使用 Docker 直接构建和运行
docker build -t nccl-test-web .
docker run -d -p 8080:8080 -v $(pwd)/data:/app/data nccl-test-web
```

访问 `http://localhost:8080` 即可使用。
