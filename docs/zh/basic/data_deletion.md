# 删除记录

## 8.1 单个记录删除

```go
// 删除单个记录
recordToDelete := map[string]any{
    "id": 1, // 必须包含主键
}
err = table.Delete(&recordToDelete)
if err != nil {
    panic(err)
}
fmt.Println("记录删除成功")
```

## 8.2 批量删除

```go
// 批量删除记录：删除年龄小于25的用户

// 首先搜索符合条件的记录
searchCriteria := map[string]any{
    "age": map[util.ComparisonOperator]any{util.LessThan: 25},
}
iter, err := table.Search(&searchCriteria)
if err != nil {
    panic(err)
}
defer GlobalTableIterPool.Put( iter)

// 使用迭代器的 Delete 方法批量删除符合条件的记录
// 直接删除所有符合条件的记录（无限制）
iter.Delete()

// 或者限制删除数量：删除前2条符合条件的记录
// iter.Delete(2)
```

## 8.3 删除表的所有数据

sfsDb 提供了 `DeleteAll()` 方法，用于删除表中的所有数据。这个方法会遍历表的所有键值对，并使用批量操作删除所有数据。

```go
// 删除表的所有数据
func (t *Table) DeleteAll() error {
	// 获取表的所有kv键值对迭代器
	iter := t.For()
	defer GlobalTableIterPool.Put( iter)
	
	// 创建批量操作
	batch := t.kvStore.GetBatch()
	if batch == nil {
		return fmt.Errorf("failed to get batch")
	}
	
	// 定义批量操作的大小限制
	const batchSizeLimit = 1000
	
	// 遍历并删除所有键值对
	count := 0
	for iter.Next() {
		key := iter.Key()
		batch.Delete(key)
		count++
		
		// 当批量操作的大小达到限制时，执行批量操作并重置批量操作对象
		if count >= batchSizeLimit {
			// 提交批量操作
			if err := t.kvStore.WriteBatch(batch); err != nil {
				return err
			}
			
			// 重置计数器和批量操作对象
			count = 0
			batch = t.kvStore.GetBatch()
			if batch == nil {
				return fmt.Errorf("failed to get batch")
			}
		}
	}
	
	// 执行剩余的批量操作
	if count > 0 {
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}
	
	return nil
}
```

**使用示例**：

```go
// 删除表的所有数据
err = table.DeleteAll()
if err != nil {
    panic(err)
}
fmt.Println("表数据删除成功")
```

**工作原理**：
1. `DeleteAll()` 方法首先调用 `For()` 方法获取表的所有键值对迭代器
2. 然后创建批量操作对象，用于批量删除键值对
3. 定义批量操作的大小限制（默认为 1000），当达到限制时执行批量操作并重置
4. 遍历迭代器，将每个键添加到批量操作中
5. 当批量操作达到限制时，执行批量操作并重置
6. 最后执行剩余的批量操作

**注意事项**：
- `DeleteAll()` 方法会删除表中的所有数据，包括索引数据
- 该方法使用分批处理机制，确保在处理大量数据时不会超过批量操作的大小限制
- 删除操作不会影响其他表的数据
- 删除后表仍然存在，只是数据被清空
- 可以在删除后重新向表中插入数据