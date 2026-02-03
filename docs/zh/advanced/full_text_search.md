# 全文搜索

## 5.1 创建全文索引

```go
// 创建全文索引
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_desc")
if err != nil {
    panic(err)
}
// 添加字段，最后一个字段必须是主键
fullTextIndex.AddFields("description", "id")
// 设置全文索引字段和长度
fullTextIndex.SetFullField("description", 5) // 5表示全文索引的长度
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

## 5.2 插入带全文索引的记录

```go
// 插入带全文索引的记录
contentRecords := []map[string]any{
    {"id": 1, "name": "商品1", "description": "这是一款高性能的笔记本电脑，适合编程和游戏"},
    {"id": 2, "name": "商品2", "description": "智能手机，拥有强大的摄像头和长续航电池"},
    {"id": 3, "name": "商品3", "description": "无线耳机，提供沉浸式音频体验"},
    {"id": 4, "name": "商品4", "description": "智能手表，可监测健康数据和接收通知"},
}

for _, record := range contentRecords {
    _, err = table.Insert(&record)
    if err != nil {
        panic(err)
    }
}
```

## 5.3 执行全文搜索

```go
// 全文搜索示例
fmt.Println("=== 全文搜索示例 ===")

// 搜索包含"笔记本"的记录
fmt.Println("\n1. 搜索 '笔记本':")
search1 := map[string]any{"description": "笔记本"}
iter1,_ := table.Search(&search1)
defer GlobalTableIterPool.Put(iter1)
records1 := iter1.GetRecords(true)
defer record.PutRecords(records1)   
for _, record := range records1 {
    fmt.Printf("   - %s: %s\n", record["name"], record["description"])
}

// 搜索包含"智能"的记录
fmt.Println("\n2. 搜索 '智能':")
search2 := map[string]any{"description": "智能"}
iter2,_ := table.Search(&search2)
defer GlobalTableIterPool.Put(iter2)
records2 := iter2.GetRecords(true)
defer record.PutRecords(records2)   
for _, record := range records2 {
    fmt.Printf("   - %s: %s\n", record["name"], record["description"])
}

// 3. 搜索结果字段选择示例
fmt.Println("\n3. 搜索结果字段选择:")
search3 := map[string]any{"description": "智能"}
iter3,_ := table.Search(&search3)
defer GlobalTableIterPool.Put(iter3)
records3 := iter3.GetRecords(true)
defer record.PutRecords(records3)   
// 使用 Select 方法只选择 name 字段
selectedNames := records3.Select("name")
fmt.Println("只显示匹配记录的名称:")
for _, record := range selectedNames {
    fmt.Printf("   - %s\n", record["name"])
}
```