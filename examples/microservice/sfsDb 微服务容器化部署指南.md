# sfsDb 微服务容器化部署指南，解决了传统数据库的依赖地狱问题

## 一、部署目的

基于 sfsDb 的核心技术优势，我们将其部署为容器化微服务，以充分发挥其在微服务架构中的价值：

### 微服务与容器化环境优势匹配度：★★★★★
- **技术栈完美契合**：Go 语言在微服务领域占据主导地位，sfsDb 作为纯 Go 实现的数据库，与微服务技术栈天然匹配
- **部署优势显著**：纯 Go 二进制部署无需依赖，非常适合容器化环境，解决了传统数据库的依赖地狱问题
- **无需 CGO 的价值**：避免了跨平台编译的复杂性，容器镜像更小，启动更快
- **复杂查询能力**：微服务架构中经常需要在本地进行数据聚合和分析，sfsDb 的复杂查询能力正好满足这一需求

## 二、项目结构

```
examples/microservice/
├── Dockerfile          # Docker 构建文件
├── go.mod              # Go 模块文件
├── main.go             # 微服务主程序
├── storage/            # 存储模块
│   ├── db.go           # 数据库操作封装
│   └── db_test.go      # 测试文件
└── 容器化部署指南.md    # 本指南
```

## 三、核心功能

- **状态管理**：存储和获取微服务本地状态
- **计数器**：实现基于本地存储的计数器功能
- **健康检查**：提供服务健康状态接口
- **轻量级存储**：利用 sfsDb 实现本地轻量级数据存储

## 四、示例代码

### main.go 微服务主程序

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/liaoran123/sfsDb/examples/microservice/storage"
)

func main() {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "default-service"
	}

	// 初始化本地存储
	err := storage.InitLocalStorage(serviceName)
	if err != nil {
		log.Fatalf("初始化本地存储失败: %v", err)
	}
	defer storage.CloseLocalStorage()

	// 注册路由
	http.HandleFunc("/api/state", handleState)
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/counter", handleCounter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务启动在端口 %s", port)
	log.Printf("本地存储路径: %s", storage.GetDBPath())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleState(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "缺少key参数", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// 获取状态
		value, err := storage.GetState(key)
		if err != nil {
			http.Error(w, fmt.Sprintf("获取状态失败: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"key": "%s", "value": "%s"}`, key, value)
	case http.MethodPost:
		// 设置状态
		value := r.FormValue("value")
		expireStr := r.FormValue("expire")
		expire := int64(0)
		if expireStr != "" {
			expireInt, err := strconv.ParseInt(expireStr, 10, 64)
			if err == nil {
				expire = expireInt
			}
		}
		err := storage.SetState(key, value, expire)
		if err != nil {
			http.Error(w, fmt.Sprintf("设置状态失败: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"key": "%s", "status": "success"}`, key)
	default:
		http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": %d}`, time.Now().Unix())
}

func handleCounter(w http.ResponseWriter, r *http.Request) {
	counterKey := "request_counter"

	switch r.Method {
	case http.MethodGet:
		// 获取计数器值
		value, err := storage.GetState(counterKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("获取计数器失败: %v", err), http.StatusInternalServerError)
			return
		}

		count := 0
		if value != "" {
			count, _ = strconv.Atoi(value)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"counter": %d}`, count)
	case http.MethodPost:
		// 增加计数器
		value, err := storage.GetState(counterKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("获取计数器失败: %v", err), http.StatusInternalServerError)
			return
		}

		count := 0
		if value != "" {
			count, _ = strconv.Atoi(value)
		}
		count++

		err = storage.SetState(counterKey, strconv.Itoa(count), 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("更新计数器失败: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"counter": %d, "status": "success"}`, count)
	default:
		http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
	}
}
```

### storage/db.go 存储模块

```go
package storage

import (
	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

var (
	localDB *engine.Table
	dbPath  string
)

// InitLocalStorage 初始化本地状态存储
func InitLocalStorage(serviceName string) error {
	// 使用DBManager打开数据库
	dbPath = "./" + serviceName + "_local_db"
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB(dbPath)
	if err != nil {
		return err
	}
	
	// 创建状态表
	table, err := engine.TableNew("service_state")
	if err != nil {
		return err
	}
	
	// 设置字段
	fields := map[string]any{
		"key":    "",  // 状态键
		"value":  "",  // 状态值
		"expire": int64(0),  // 过期时间（ Unix 时间戳）
	}
	err = table.SetFields(fields)
	if err != nil {
		return err
	}
	
	// 创建主键索引
	pk, _ := engine.DefaultPrimaryKeyNew("pk")
	pk.AddFields("key")
	err = table.CreateIndex(pk)
	if err != nil {
		return err
	}
	
	localDB = table
	return nil
}

// GetState 获取本地状态
func GetState(key string) (string, error) {
	fields := map[string]any{"key": key}
	iter, err := localDB.Search(&fields)
	if err != nil {
		return "", err
	}
	defer iter.Release()
	
	records := iter.GetRecords(true)
	defer records.Release()
	
	if len(records) == 0 {
		return "", nil
	}
	
	return records[0]["value"].(string), nil
}

// SetState 设置本地状态
func SetState(key, value string, expire int64) error {
	fields := map[string]any{
		"key":    key,
		"value":  value,
		"expire": expire,
	}
	_, err := localDB.Insert(&fields)
	return err
}

// CloseLocalStorage 关闭本地存储
func CloseLocalStorage() error {
	dbManager := storage.GetDBManager()
	return dbManager.CloseDB()
}

// GetDBPath 获取数据库路径
func GetDBPath() string {
	return dbPath
}
```

## 五、容器化部署步骤

### 1. 准备构建环境

确保本地安装了以下工具：
- Docker
- Go 1.20+

### 2. 创建 Dockerfile

```dockerfile
# 第一阶段：构建
FROM golang:1.20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o microservice-app main.go

# 第二阶段：运行
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=builder /app/microservice-app .

# 创建数据目录
RUN mkdir -p /app/data

# 设置环境变量
ENV SERVICE_NAME=microservice-demo
ENV PORT=8080

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./microservice-app"]
```

### 3. 创建 go.mod 文件

```go
module github.com/liaoran123/sfsDb/examples/microservice

go 1.20

require (
	github.com/liaoran123/sfsDb v0.0.0
)

replace github.com/liaoran123/sfsDb => ../..
```

### 4. 构建容器镜像

```bash
# 进入微服务目录
cd examples/microservice

# 构建镜像
docker build -t sfsdb-microservice:latest .
```

### 5. 运行容器

#### 5.1 基本运行（临时存储）

```bash
docker run -d \
  --name sfsdb-microservice \
  -p 8080:8080 \
  -e SERVICE_NAME=my-microservice \
  sfsdb-microservice:latest
```

#### 5.2 持久化存储运行

```bash
docker run -d \
  --name sfsdb-microservice \
  -p 8080:8080 \
  -e SERVICE_NAME=my-microservice \
  -v sfsdb-data:/app/data \
  sfsdb-microservice:latest
```

#### 5.3 多实例运行（负载均衡）

```bash
# 创建网络
docker network create sfsdb-network

# 运行多个实例
docker run -d --name sfsdb-microservice-1 --network sfsdb-network -e SERVICE_NAME=microservice-1 sfsdb-microservice:latest
docker run -d --name sfsdb-microservice-2 --network sfsdb-network -e SERVICE_NAME=microservice-2 sfsdb-microservice:latest

# 创建负载均衡器
docker run -d \
  --name sfsdb-lb \
  --network sfsdb-network \
  -p 8080:80 \
  -e BACKEND_HOSTS=sfsdb-microservice-1:8080,sfsdb-microservice-2:8080 \
  nginx:alpine
```

## 六、Kubernetes 部署

### 1. 创建部署配置文件

```yaml
# kubernetes-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sfsdb-microservice
  labels:
    app: sfsdb-microservice
spec:
  replicas: 2
  selector:
    matchLabels:
      app: sfsdb-microservice
  template:
    metadata:
      labels:
        app: sfsdb-microservice
    spec:
      containers:
      - name: sfsdb-microservice
        image: sfsdb-microservice:latest
        ports:
        - containerPort: 8080
        env:
        - name: SERVICE_NAME
          value: sfsdb-microservice
        - name: PORT
          value: "8080"
        volumeMounts:
        - name: local-db
          mountPath: /app/data
      volumes:
      - name: local-db
        emptyDir: {}

---
apiVersion: v1
kind: Service
metadata:
  name: sfsdb-microservice
spec:
  selector:
    app: sfsdb-microservice
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 2. 部署到 Kubernetes

```bash
# 应用部署配置
kubectl apply -f kubernetes-deployment.yaml

# 查看部署状态
kubectl get pods
kubectl get services
```

## 七、API 接口测试

### 1. 健康检查

```bash
curl http://localhost:8080/api/health
# 预期响应：{"status": "healthy", "timestamp": 1678901234}
```

### 2. 设置状态

```bash
curl -X POST http://localhost:8080/api/state?key=test -d "value=hello"
# 预期响应：{"key": "test", "status": "success"}
```

### 3. 获取状态

```bash
curl http://localhost:8080/api/state?key=test
# 预期响应：{"key": "test", "value": "hello"}
```

### 4. 增加计数器

```bash
curl -X POST http://localhost:8080/api/counter
# 预期响应：{"counter": 1, "status": "success"}
```

### 5. 获取计数器值

```bash
curl http://localhost:8080/api/counter
# 预期响应：{"counter": 1}
```

## 八、性能优势

### 1. 启动速度

- **传统数据库**：需要启动数据库服务，加载配置，初始化连接池，通常需要数秒到数十秒
- **sfsDb**：作为嵌入式数据库，与应用同进程启动，无需额外服务，启动时间毫秒级

### 2. 资源占用

- **传统数据库**：通常需要数百 MB 内存，CPU 占用较高
- **sfsDb**：内存占用低，适合资源受限的容器环境，最小仅需几十 MB 内存

### 3. 部署复杂度

- **传统数据库**：需要单独部署、配置、监控，增加了系统复杂度
- **sfsDb**：与应用一体化部署，无需额外配置，简化了部署流程

## 九、最佳实践

### 1. 存储策略

- **临时数据**：使用 emptyDir 存储，适合无需持久化的数据
- **重要数据**：使用 PersistentVolumeClaim 存储，确保数据持久化
- **数据清理**：定期清理过期数据，避免存储膨胀

### 2. 安全建议

- **非 root 用户**：在 Dockerfile 中使用非 root 用户运行应用
- **最小权限**：只授予应用必要的文件系统权限
- **网络隔离**：使用 Kubernetes NetworkPolicy 限制网络访问

### 3. 监控与可观测性

- **健康检查**：配置 Kubernetes 就绪探针和存活探针
- **日志管理**：使用结构化日志，便于收集和分析
- **指标监控**：暴露 Prometheus 指标，监控存储使用情况

## 十、故障排查

### 1. 常见问题

| 问题 | 可能原因 | 解决方案 |
|------|---------|----------|
| 容器启动失败 | 端口冲突 | 检查端口占用，修改环境变量 PORT |
| 数据丢失 | 使用临时存储 | 配置持久化存储卷 |
| 性能下降 | 数据量过大 | 实现数据清理策略，定期压缩数据库 |
| 连接拒绝 | 网络配置问题 | 检查网络策略，确保服务可访问 |

### 2. 日志查看

```bash
# Docker 容器日志
docker logs sfsdb-microservice

# Kubernetes Pod 日志
kubectl logs pod/sfsdb-microservice-xxx
```

## 十一、总结

通过容器化部署，sfsDb 可以充分发挥其在微服务架构中的优势：

1. **轻量级部署**：与应用一体化部署，无需额外数据库服务
2. **高性能**：本地存储，低延迟，适合微服务的实时性要求
3. **可靠性**：支持事务，确保数据一致性
4. **灵活性**：纯 Go 实现，跨平台部署简单
5. **成本效益**：减少了基础设施成本，简化了运维复杂度

sfsDb 作为微服务的本地存储解决方案，为微服务架构提供了一种轻量级、高性能的存储选择，特别适合边缘计算、IoT 网关、Serverless 函数等资源受限的场景。

---

**部署成功后，您的微服务将具备本地状态管理能力，同时享受容器化带来的一致性、可移植性和易于管理的优势。**
