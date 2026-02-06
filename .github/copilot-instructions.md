## 目标

为 AI 编码助理提供尽快上手本仓库的关键上下文：总体架构、常见开发/调试流程、项目特定约定、以及容易踩的坑。指令尽量短小精悍，可直接用于自动化提示模板。

## 一句话概览

sfsDb 是用 Go 实现的轻量级嵌入式关系型数据库，核心分层为 `storage`（持久化引擎接口与 LevelDB/RocksDB 实现）和 `engine`（表/索引/查询/事务逻辑）。项目大量使用对象池、迭代器模式与可替换的存储后端。

## 必读位置（优先级排序）

- `README.md`（项目概览与快速示例）
- `docs/README.md`、`docs/en/README.md`、`docs/zh/README.md`（详细使用与设计说明）
- `storage/README.md`（存储接口与引擎实现，关键抽象：`Store`, `Batch`, `Iterator`, `Snapshot`）
- `engine/` 目录（表、索引、查询、事务的实现与测试）
- `examples/`（示例代码）

## 关键架构要点（可用作检索/补全上下文）

- 存储层抽象：`storage.Store` 接口是系统的持久层契约（`Get/Put/Delete/Batch/Iterator/Snapshot/Close`）。任何自定义存储引擎都通过实现该接口接入（参考 `storage/leveldb.go`）。
- 引擎层：`engine` 提供 `TableNew`, `SetFields`, `CreateIndex`, `Insert`, `Search`, `Update`, `Delete` 等高层 API。索引类型包括主键/普通索引和全文索引。
- 迭代器与对象池：查询返回迭代器（需 `Release()`），并经常配合对象池（例如 `GlobalTableIterPool`）与 `PutRecords`/`PutRecordSet` 回收记录对象。忘记 Release/Put 会导致资源泄露。
- 存储注入点：使用 `storage.SetStore()` 可以替换后端存储（便于测试或集成自定义引擎）。

## 项目约定与常见模式

- 使用 Go 的 `any` 类型表示任意字段值（见 `engine` 示例）。
- 严格资源管理：凡是带 `Close()` / `Release()` 的对象都必须在使用后清理（DB、迭代器、快照、池中对象）。
- 批量写入通过 `Batch()` / `Commit()` 实现，用于高吞吐写入场景（参考 `storage/README.md` 的示例）。
- 测试分布在各个包（大量在 `engine/`），项目同时包含 `ginkgo`/`gomega` 测试依赖，但 `go test` 仍可运行。

## 开发 / 构建 / 测试 快速命令（Windows PowerShell）

- 使用的 Go 版本：参考 `go.mod` 中的 `go 1.25.3`。
- 构建整个仓库（在仓库根目录）：

```powershell
go build ./...
```

- 运行所有测试（并发执行包内测试）：

```powershell
go test ./... -v
```

- 运行单个包测试（例如 engine）：

```powershell
go test ./engine -v
```

- 运行带名测试或调试单测：

```powershell
go test ./engine -run TestName -v
```

如果需要使用 `ginkgo` 的 BDD 风格测试，请先安装 `ginkgo`，但大部分 CI 与本地检查直接使用 `go test` 即可。

## 配置与集成点

- 默认持久化：项目默认使用 LevelDB（`github.com/syndtr/goleveldb`），相关实现见 `storage/leveldb.go`。
- 可替换存储：参考 `storage.SetStore()` 与 `storage.NewStore(config)`。要在测试或部署中替换后端，优先通过这些 API。
- Web/服务依赖：仓库包含 `github.com/gin-gonic/gin`（用于示例或管理服务），相关代码散布在 `web_example_db/` 或 `services/`。

## 易犯错误（写给自动化 agent 的提示）

- 在补全或生成修改涉及查询/迭代的代码时，确保在退出路径调用 `iter.Release()` 并正确回收记录池（`PutRecords` / `PutRecordSet`）。
- 不要在生成生产 DB 路径时使用相对路径默认值（CI/测试应使用临时目录）。
- 修改存储接口时，保持 `Store`/`Iterator`/`Snapshot` 行为向后兼容，仓库中多处直接依赖这些语义。

## 可用示例片段（可直接插入补全）

- 打开默认 DB：`storage.OpenDefaultDb("./basic_example_db")` -> `defer storage.CloseDb()`
- 创建表并设置字段/索引：参考 `README.md` 的示例，核心调用 `engine.TableNew`, `SetFields`, `DefaultPrimaryKeyNew`, `CreateIndex`, `Insert`, `Search`。

## 最后——如果不确定，先做小改动并运行本地测试

修改前优先修改小函数/模块并运行 `go test ./...`。把不确定的改动包装成小 PR，附带针对受影响功能的测试。需要我把这些规则合并到 CI 或生成示例脚本时请告诉我。

---

如果要我把某部分展开为可执行脚本（例如 CI 步骤或常用 devcontainer 配置），或把 README 中的示例转换为小的可运行演示程序，告诉我想要的目标（CI/本地/Windows/macOS/Linux）。
