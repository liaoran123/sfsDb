# sfsDb 构建指南

## 1. 概述

本指南详细介绍 sfsDb 的构建方法，包括预编译解决方案、交叉编译配置、Docker 构建环境以及 C/C++ 集成步骤。通过本指南，您可以快速搭建 sfsDb 的构建环境，为工业嵌入式系统生成适配的预编译库。

## 2. 预编译解决方案

### 2.1 下载预编译包

sfsDb 提供了完整的预编译包，支持多种架构：

- **x86_64 (amd64)**: 适用于工业工控机
- **ARM64 (arm64)**: 适用于 ARM 开发板和工业网关
- **MIPS64 (mips64)**: 适用于 MIPS 路由器和嵌入式设备

### 2.2 预编译包内容

预编译包包含以下文件：

```
sfsdb-prebuild-v1.0.0/
├── lib/                  # 静态库文件
│   ├── sfsdb_linux_amd64.a     # x86 工控机
│   ├── sfsdb_linux_arm64.a     # ARM64 网关
│   └── sfsdb_linux_mips64.a    # MIPS64 路由器
├── demo/                 # 示例程序
│   ├── demo_linux_amd64        # x86 示例
│   ├── demo_linux_arm64        # ARM64 示例
│   └── demo_linux_mips64       # MIPS64 示例
├── examples/             # 示例源码
│   ├── main.go                 # Go 语言示例
│   ├── c_example/              # C 语言示例
│   │   ├── main.c              # C 示例代码
│   │   ├── Makefile            # Make 配置
│   │   └── CMakeLists.txt      # CMake 配置
│   └── performance_test.go     # 性能测试代码
├── RELEASE_NOTES.md      # 发布说明
└── INTEGRATION_GUIDE.md  # 集成指南
```

## 3. 本地构建

### 3.1 环境要求

- **Go 版本**: 1.20 或更高版本
- **操作系统**: Linux、Windows
- **交叉编译工具链** (可选): 用于 ARM64 和 MIPS64 架构

### 3.2 Linux 构建

#### 步骤 1: 克隆代码库

```bash
git clone https://github.com/liaoran123/sfsDb.git
cd sfsDb
```

#### 步骤 2: 运行构建脚本

```bash
# 运行构建脚本
./build/build.sh

# 构建结果在 build/output 目录
ls -la build/output/
```

### 3.3 Windows 构建

#### 步骤 1: 克隆代码库

```powershell
git clone https://github.com/liaoran123/sfsDb.git
cd sfsDb
```

#### 步骤 2: 运行构建脚本

```powershell
# 运行构建脚本（使用 PowerShell）
powershell -ExecutionPolicy Bypass -File .\build\build.ps1

# 构建结果在 build/output 目录
dir .\build\output\
```

### 3.4 构建依赖

- **静态库构建**: 需要 `ar` 命令（Git for Windows 或 MinGW-w64 提供）
- **交叉编译**: 需要对应架构的交叉编译工具链

## 4. Docker 构建环境

### 4.1 构建 Docker 镜像

```bash
# 构建 Docker 镜像
docker build -t sfsdb-build-env .
```

### 4.2 运行 Docker 容器

```bash
# 运行容器进行构建
docker run -it --rm -v $(pwd):/app sfsdb-build-env bash -c "./build/build.sh"
```

### 4.3 Docker 环境优势

- **一致性**: 确保在任何环境中生成相同的构建结果
- **依赖管理**: 内置所有必要的依赖和工具链
- **隔离性**: 不会影响本地系统环境

## 5. 交叉编译配置

### 5.1 ARM64 交叉编译

```bash
# 设置环境变量
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=0

# 编译静态库
go build -buildmode=c-archive -o lib/sfsdb_linux_arm64.a build/main.go

# 编译示例程序
go build -o demo/demo_linux_arm64 build/main.go
```

### 5.2 MIPS64 交叉编译

```bash
# 设置环境变量
export GOOS=linux
export GOARCH=mips64
export CGO_ENABLED=0

# 编译示例程序（静态库不支持）
go build -o demo/demo_linux_mips64 build/main.go
```

## 6. C/C++ 集成

### 6.1 使用 Makefile 集成

#### 步骤 1: 复制库文件

```bash
# 复制静态库到项目目录
cp sfsdb-prebuild-v1.0.0/lib/sfsdb_linux_arm64.a /path/to/your/project/lib/

# 复制头文件（如果需要）
cp sfsdb-prebuild-v1.0.0/examples/c_example/*.h /path/to/your/project/include/
```

#### 步骤 2: 配置 Makefile

```makefile
# sfsDb 集成配置
SFSDB_LIB_DIR = ./lib
SFSDB_LIB = -lsfsdb_linux_arm64

# 编译选项
CFLAGS += -I./include
LDFLAGS += -L$(SFSDB_LIB_DIR) $(SFSDB_LIB)

# 目标文件
target: main.o
	$(CC) $(LDFLAGS) -o $@ $^
```

#### 步骤 3: 编译项目

```bash
# 编译项目
make

# 运行程序
./target
```

### 6.2 使用 CMake 集成

#### 步骤 1: 复制库文件

```bash
# 复制静态库到项目目录
cp sfsdb-prebuild-v1.0.0/lib/sfsdb_linux_arm64.a /path/to/your/project/lib/
```

#### 步骤 2: 配置 CMakeLists.txt

```cmake
# sfsDb 集成配置
set(SFSDB_LIB_DIR "${CMAKE_SOURCE_DIR}/lib")
set(SFSDB_LIBRARY "sfsdb_linux_arm64")

# 链接 sfsDb 库
target_link_libraries(your_target
    PRIVATE
        -L${SFSDB_LIB_DIR}
        -l${SFSDB_LIBRARY}
)
```

#### 步骤 3: 编译项目

```bash
# 创建构建目录
mkdir -p build && cd build

# 配置 CMake
cmake ..

# 编译项目
cmake --build .

# 运行程序
./your_target
```

## 7. 性能优化

### 7.1 内存优化

- **使用批量操作**: 对于大量数据插入，使用批量操作减少内存占用
- **合理设置缓存大小**: 根据设备内存情况调整缓存参数
- **及时释放资源**: 使用完数据库后及时关闭，避免内存泄漏

### 7.2 速度优化

- **创建合适的索引**: 为频繁查询的字段创建索引
- **使用主键查询**: 主键查询速度最快，尽量使用主键进行查询
- **减少并发操作**: 在资源有限的设备上，合理控制并发数

### 7.3 稳定性优化

- **使用事务**: 对于关键操作，使用事务确保数据一致性
- **定期备份**: 定期备份数据库，防止数据丢失
- **错误处理**: 完善的错误处理机制，提高系统稳定性

## 8. 常见问题解决方案

### 8.1 编译错误

#### 问题：undefined reference to `xxx`

**解决方案**：
- 检查库文件是否正确链接
- 确认使用了正确架构的库文件
- 检查编译命令是否包含 `-L` 和 `-l` 选项

#### 问题：cannot find -lsfsdb_linux_arm64

**解决方案**：
- 确认库文件存在于指定目录
- 检查库文件权限是否正确
- 确认使用了正确的库文件名称

### 8.2 运行错误

#### 问题：数据库初始化失败

**解决方案**：
- 检查数据库路径是否可写
- 确认设备有足够的存储空间
- 检查文件系统权限

#### 问题：内存不足

**解决方案**：
- 减少批量操作的大小
- 增加设备内存（如果可能）
- 优化数据结构，减少内存使用

### 8.3 性能问题

#### 问题：查询速度慢

**解决方案**：
- 为查询字段创建索引
- 使用主键查询
- 减少查询返回的数据量

#### 问题：插入速度慢

**解决方案**：
- 使用批量插入
- 减少索引数量
- 优化数据结构

## 9. CI/CD 集成

### 9.1 GitHub Actions 配置

sfsDb 提供了完整的 GitHub Actions 配置，用于自动构建和发布预编译包：

- **自动交叉编译**：每次代码提交自动执行多架构交叉编译
- **生成预编译包**：自动生成包含所有架构的预编译包
- **发布到 GitHub Release**：自动上传预编译包到 Release 页面

### 9.2 本地 CI 配置

如果需要在本地或私有 CI 系统中配置构建流程，可以参考以下步骤：

1. **安装依赖**：Go 1.20+、交叉编译工具链
2. **配置环境变量**：设置 `GOOS`、`GOARCH` 等环境变量
3. **执行构建脚本**：运行 `./build/build.sh` 或 `./build/build.ps1`
4. **上传构建产物**：将预编译包上传到指定位置

## 10. 版本管理

### 10.1 版本号格式

sfsDb 使用语义化版本号格式：

```
v{主版本}.{次版本}.{补丁版本}
```

- **主版本**：不兼容的 API 变更
- **次版本**：向后兼容的功能添加
- **补丁版本**：向后兼容的 bug 修复

### 10.2 版本升级

- **补丁版本升级**：直接替换库文件即可
- **次版本升级**：可能需要更新头文件，API 保持兼容
- **主版本升级**：需要更新代码以适配新的 API

## 11. 技术支持

### 11.1 问题反馈

如果遇到问题，可以通过以下方式反馈：

- **GitHub Issues**：https://github.com/liaoran123/sfsDb/issues
- **邮件**：support@sfsdb.io
- **社区**：https://github.com/liaoran123/sfsDb/discussions

### 11.2 常见问题

**Q**：sfsDb 支持哪些操作系统？
**A**：sfsDb 主要支持 Linux 系统，包括 x86、ARM64、MIPS64 架构。

**Q**：sfsDb 适合哪些场景？
**A**：sfsDb 适合工业嵌入式系统、边缘计算设备、IoT 网关等资源受限的场景。

**Q**：sfsDb 支持事务吗？
**A**：是的，sfsDb 支持完整的 ACID 事务，包括四种隔离级别。

**Q**：如何获取最新版本的 sfsDb？
**A**：可以从 GitHub Release 页面下载最新的预编译包，或从源码构建。

## 12. 总结

sfsDb 提供了完整的构建解决方案，支持多架构交叉编译、C/C++ 集成和 Docker 构建环境。通过本指南的步骤，您可以快速为工业嵌入式系统构建和集成 sfsDb，享受其带来的高性能和可靠性。

如果您有任何问题或建议，欢迎随时反馈，我们将不断改进 sfsDb，为工业嵌入式领域提供更好的数据库解决方案。