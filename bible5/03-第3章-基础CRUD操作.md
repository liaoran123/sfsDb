# 第 3 章：基础 CRUD 操作

---

## 3.1 完整的可运行示例

让我们通过一个完整的示例来演示 sfsDb 的基础 CRUD 操作：

```go
package main

import (
	"fmt"
	"os"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

func main() {
	fmt.Println("=== sfsDb 数据库实战 - 第 3 章：基础 CRUD 操作示例 ===")
	fmt.Println()

	dbPath := "./bible_examples_chapter03_db"
	cleanup(dbPath)

	fmt.Println("1. 初始化数据库")
	_, err := storage.GetDBManager().OpenDB(dbPath)
	if err != nil {
		fmt.Printf("数据库打开失败: %v\n", err)
		return
	}
	defer storage.GetDBManager().CloseDB()
	fmt.Println("   ✓ 数据库打开成功")
	fmt.Println()

	fmt.Println("2. 创建表并设置字段")
	table, err := engine.TableNew("employees")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	fields := map[string]any{
		"id":     0,
		"name":   "",
		"age":    0,
		"department": "",
		"salary": 0,
	}
	err = table.SetFields(fields)
	if err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	primaryKey, err := engine.DefaultPrimaryKeyNew("pk_id")
	primaryKey.AddFields("id")
	table.CreateIndex(primaryKey)
	fmt.Println("   ✓ 表创建成功")
	fmt.Println()

	fmt.Println("3. 插入数据 (Create)")
	employees := []map[string]any{
		{"id": 1, "name": "张三", "age": 28, "department": "技术部", "salary": 15000},
		{"id": 2, "name": "李四", "age": 32, "department": "市场部", "salary": 12000},
		{"id": 3, "name": "王五", "age": 25, "department": "技术部", "salary": 13000},
	}

	for _, emp := range employees {
		_, err = table.Insert(&amp;emp)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
	}
	fmt.Println("   ✓ 成功插入", len(employees), "条员工数据")
	fmt.Println()

	fmt.Println("4. 查询数据 (Read)")
	searchAll := map[string]any{"id": nil}
	dataIter, _ := table.Search(&amp;searchAll)
	defer dataIter.Release()
	
	records := dataIter.GetRecords(true)
	defer records.Release()
	
	fmt.Println("   所有员工列表:")
	for _, emp := range records {
		fmt.Printf("   - ID: %v, 姓名: %v, 年龄: %v, 部门: %v, 薪资: %v\n",
			emp["id"], emp["name"], emp["age"], emp["department"], emp["salary"])
	}
	fmt.Println()

	fmt.Println("5. 更新数据 (Update)")
	updateEmp := map[string]any{
		"id":     1,
		"salary": 18000,
	}
	err = table.Update(&amp;updateEmp)
	if err != nil {
		fmt.Printf("更新数据失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ 成功更新员工 ID=1 的薪资")
	
	searchEmp := map[string]any{"id": 1}
	dataIter2, _ := table.Search(&amp;searchEmp)
	defer dataIter2.Release()
	updatedRecord := dataIter2.GetRecords(true)
	defer updatedRecord.Release()
	fmt.Printf("   更新后的数据: %+v\n", updatedRecord[0])
	fmt.Println()

	fmt.Println("6. 删除数据 (Delete)")
	deleteEmp := map[string]any{"id": 3}
	err = table.Delete(&amp;deleteEmp)
	if err != nil {
		fmt.Printf("删除数据失败: %v\n", err)
		return
	}
	fmt.Println("   ✓ 成功删除员工 ID=3")
	
	dataIter3, _ := table.Search(&amp;searchAll)
	defer dataIter3.Release()
	finalRecords := dataIter3.GetRecords(true)
	defer finalRecords.Release()
	fmt.Println("   删除后的员工列表:")
	for _, emp := range finalRecords {
		fmt.Printf("   - ID: %v, 姓名: %v\n", emp["id"], emp["name"])
	}
	fmt.Println()

	fmt.Println("=== CRUD 操作演示完成！===")
}

func cleanup(path string) {
	os.RemoveAll(path)
}
```

这段代码展示了 sfsDb 的完整 CRUD 操作流程：

1. **初始化数据库**：通过 `storage.GetDBManager().OpenDB()` 打开数据库
2. **创建表**：使用 `engine.TableNew()` 创建表
3. **设置字段**：通过 `SetFields()` 定义表结构
4. **创建索引**：使用 `CreateIndex()` 创建主键索引
5. **插入数据**：使用 `Insert()` 插入多条员工数据
6. **查询数据**：使用 `Search()` 查询所有记录
7. **更新数据**：使用 `Update()` 更新指定记录
8. **删除数据**：使用 `Delete()` 删除指定记录

---

## 3.2 小结

本章我们通过项目的实际代码学习了 sfsDb 的基础 CRUD 操作：

- **创建表**：使用 `TableNew()` 和 `SetFields()`
- **插入数据**：使用 `Insert()`，支持自动递增 ID
- **查询数据**：使用 `Search()` 和 `ForData()`
- **更新数据**：使用 `Update()`，通过主键定位
- **删除数据**：使用 `Delete()`，通过主键删除
- **资源管理**：及时释放迭代器和记录

在下一章中，我们将深入学习索引与查询优化。


---

**本书版本**：1.0.0  
**最后更新**：2026-03-11  
**sfsDb** - 以工业物联网边缘计算为核心场景的高性能嵌入式数据库！🚀  
**技术栈** - Go、leveldb。纯golang实现。
**项目地址**：[GitHub](https://github.com/liaoran123/sfsDb)  
**GitCode 镜像**：[GitCode](https://gitcode.com/liuyun258369/sfsDb)
