package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

// TestTableSearch 测试使用加密数据库表遍历数据和Search方法的功能
func TestTableSearch1(t *testing.T) {
	// 生成测试密钥
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}
	// 创建加密配置
	encryptConfig := &storage.EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: masterKey,
	}
	// 生成唯一的数据库路径和表名，避免测试之间的数据冲突
	dbPath := "./test_encrypted_table_db_" + t.Name()
	tableName := "test_search_" + t.Name()
	// 初始化加密的全局KVDb
	_, err := storage.GetDBManager().OpenDBWithEncryption(dbPath, encryptConfig)
	//_, err := storage.OpenDefaultDbWithEncryption(dbPath, encryptConfig)
	if err != nil {
		t.Fatalf("Failed to open encrypted database: %v", err)
	}
	// 创建测试表
	table, err := TableNew(tableName)
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}
	if table == nil {
		t.Fatal("TableNew 失败")
	}
	// 必须先为表预设字段和数据类型
	fields := map[string]any{"id": 0, "name": "", "age": uint8(0), "description": ""}
	table.SetFields(fields)

	PrimaryKeys, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew 失败: %v", err)
	}
	PrimaryKeys.AddFields("id")    //创建一个id的组合主键
	table.CreateIndex(PrimaryKeys) //将组合主键设置到表中

	fullText, err := DefaultFullTextIndexNew("ft")
	if err != nil {
		t.Fatalf("DefaultFullTextIndexNew 失败: %v", err)
	}
	//全文索引正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义。
	fullText.AddFields("description", "id") //创建一个description的组合全文索引
	//指定description为全文索引字段，长度为5
	//如果没有指定，则等同一般索引
	err = fullText.SetFullField("description", 5) //添加description全文索引字段，长度为5
	if err != nil {
		t.Fatalf("SetFullField 失败: %v", err)
	}
	table.CreateIndex(fullText) //将组合全文索引设置到表中

	normalIndex, err := DefaultNormalIndexNew("idx")
	if err != nil {
		t.Fatalf("DefaultNormalIndexNew 失败: %v", err)
	}
	normalIndex.AddFields("name", "age") //创建一个name, age的组合普通索引
	table.CreateIndex(normalIndex)       //将组合普通索引设置到表中

	// 插入测试数据
	data := []map[string]any{
		{"id": 1, "name": "六月", "age": uint8(25), "description": "古木阴阴六月凉，幽花藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
		{"id": 2, "name": "Bob", "age": uint8(30), "description": "Bob is a product manager"},
		{"id": 3, "name": "Charlie", "age": uint8(35), "description": "Charlie is1 a designer"},
		{"id": 4, "name": "David", "age": uint8(40), "description": "David isnot a developer"},
		{"id": 5, "name": "Eve", "age": uint8(45), "description": "Eve is an manager"},
		{"id": 6, "name": "Alice", "age": uint8(27), "description": "Alice is2 a software engineer"},
		{"id": 7, "name": "Eve 49", "age": uint8(49), "description": "Eve is3 a manager 49"}, //"id": nil 使用自动增值
		{"id": 8, "name": "Eve 55", "age": uint8(55), "description": "Eve is4 a manager 55"}, //"id": nil 使用自动增值
	}
	for _, item := range data {
		fields := table.GetAllFields()
		fields["id"] = item["id"]
		fields["name"] = item["name"]
		fields["age"] = item["age"]
		fields["description"] = item["description"]
		_, err := table.Insert(&fields)
		if err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
		/*
				if item["id"] == nil {
					continue
				}
			if currentID != util.AnyToInt(item["id"]) {
				t.Errorf("插入测试数据后，当前ID应为%v，实际: %d", util.AnyToInt(item["id"]), currentID)
			}
		*/
	}
	//测试遍历表所有kv
	t.Run("For", func(t *testing.T) {
		dataIter := table.For()
		for dataIter.Next() {
			//fmt.Printf("dataIter.Key(): %v\n", dataIter.Key())
			fmt.Printf("key: %s, value: %s\n", dataIter.Key(), dataIter.Value())
			//val := table.ParseValue(dataIter.Value())
			//fmt.Printf("val: %v\n", val)
		}
		dataIter.Release()
	})
	fmt.Println("-----------------")
	// 测试遍历表所有数据
	t.Run("ForData", func(t *testing.T) {
		// 使用ForData方法遍历所有数据
		dataIter := table.ForData()
		defer dataIter.Release()
		//defer GlobalTableIterPool.Put(dataIter)
		rs := dataIter.GetRecords(true)
		defer record.PutRecords(rs)

		for _, item := range rs {
			fmt.Printf("records: %v\n", item)
		}
	})
	fmt.Println("-----------------")

	// 测试1: 主键搜索
	t.Run("PrimaryKeySearch", func(t *testing.T) {
		// 使用Search方法搜索主键为3的记录
		fields := map[string]any{
			"id": 1,
		}
		dataIter, _ := table.Search(&fields)
		defer GlobalTableIterPool.Put(dataIter)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer GlobalTableIterPool.Put(dataIter)
		records := dataIter.GetRecords(true)
		defer record.PutRecords(records)
		fmt.Printf("records: %v\n", records)
		//判断data[0]和records是否相等

		if records[0]["name"] != data[0]["name"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["name"], records[0]["name"])
		}
		if records[0]["age"] != data[0]["age"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["age"], records[0]["age"])
		}
		if records[0]["description"] != data[0]["description"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["description"], records[0]["description"])
		}

	})

	// 测试2: 索引字段搜索
	t.Run("IndexSearch", func(t *testing.T) {
		// 使用Search方法搜索name为"Charlie"的记录
		fields := map[string]any{
			"name": "Charlie",
		}
		dataIter, _ := table.Search(&fields)
		defer GlobalTableIterPool.Put(dataIter)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer GlobalTableIterPool.Put(dataIter)
		records := dataIter.GetRecords(true)
		defer record.PutRecords(records)
		for _, item := range records.Select("name", "age", "description") {
			fmt.Printf("records: %v\n", item)
		}

		if records[0]["age"] != data[2]["age"] {
			t.Errorf("搜索name为Charlie的记录错误，期望: %v, 实际: %v", data[2]["age"], records[0]["age"])
		}
	})

	// 测试3: 全文索引搜索
	t.Run("FullTextSearch", func(t *testing.T) {
		// 使用Search方法搜索description包含"Bob"的记录
		fields := map[string]any{
			"description": "Bob",
		}
		dataIter, _ := table.Search(&fields)
		defer GlobalTableIterPool.Put(dataIter)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer GlobalTableIterPool.Put(dataIter)
		records := dataIter.GetRecords(true)
		defer record.PutRecords(records)
		fmt.Printf("records: %v\n", records)
		for _, item := range records.Select("name", "age", "description") {
			fmt.Printf("records: %v\n", item)
		}

		sdata := []map[string]any{
			{"description": "古木阴阴六月凉，幽花藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "六月凉，幽花藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "幽花藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "藉藉四时香。——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "——裘万顷《次余仲庸松风阁韵十九首其三》"},
			{"description": "次余仲庸松风阁韵十九首其三》"},
			{"description": "十九首其三》"},
			{"description": "。——裘万顷《次余仲庸松风阁韵十九首其三》"},
		}
		for _, item := range sdata {
			fields := map[string]any{
				"description": item["description"],
			}
			dataIter, _ := table.Search(&fields)
			defer GlobalTableIterPool.Put(dataIter)
			if dataIter.iter == nil {
				t.Fatalf("Search 失败")
			}
			//defer GlobalTableIterPool.Put(dataIter)

			records := dataIter.GetRecords(true)
			defer record.PutRecords(records)
			for _, item := range records.Select("name", "age", "description") {
				fmt.Printf("搜索:%v -》 records: %v\n", fields["description"], item)
			}

			//判断data[1]和records是否相等
			if records[0]["name"] != data[0]["name"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["name"], records[0]["name"])
			}
			if records[0]["age"] != data[0]["age"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["age"], records[0]["age"])
			}
			if records[0]["description"] != data[0]["description"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["description"], records[0]["description"])
			}
		}

		sdata1 := []map[string]any{
			{"description": "Bob is a product manager"},
			{"description": "is a product manager"},
			{"description": "a product manager"},
			{"description": "product manager"},
			{"description": "ct ma"},
			//{"description": "is"},
			{"description": " a product manager"},
			{"description": " product"},
		}
		for _, item := range sdata1 {
			fields := map[string]any{
				"description": item["description"],
			}
			dataIter, _ := table.Search(&fields)
			defer GlobalTableIterPool.Put(dataIter)
			if dataIter.iter == nil {
				t.Fatalf("Search 失败")
			}
			//defer GlobalTableIterPool.Put(dataIter)

			records := dataIter.GetRecords(true)
			defer record.PutRecords(records)
			for _, item := range records.Select("name", "age", "description") {
				fmt.Printf("搜索:%v -》 records: %v\n", fields["description"], item)
			}
			if records[0]["id"] != data[1]["id"] {
				fmt.Printf("查询结果可能是多个: %v\n。但是测试并没有错误。", records)
				//t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[1]["id"], records[0][table.GetPrimaryKey().GetFields()[0]])
			}
		}

		indexStats := GetIndexMatchCacheStats()

		fmt.Printf("索引匹配缓存 - 大小: %d, 访问: %d, 命中: %d, 命中率: %.2f%%\n",
			indexStats.Size, indexStats.Accesses, indexStats.Hits, indexStats.HitRate*100)

	})

	// 测试4: 搜索不存在的数据
	t.Run("SearchNonExistent", func(t *testing.T) {
		fields := map[string]any{
			"id": 100,
		}
		// 使用Search方法搜索不存在的id
		dataIter, _ := table.Search(&fields)
		defer GlobalTableIterPool.Put(dataIter)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer GlobalTableIterPool.Put(dataIter)
		records := dataIter.GetRecords(true)
		defer record.PutRecords(records)
		//判断data[1]和records是否相等
		if len(records) > 0 {
			t.Errorf("搜索不存在的记录错误，期望: 空结果, 实际: %v", records)
		} else {
			t.Logf("搜索不存在的记录成功，返回空结果")
		}
	})

	//打开所有记录
	t.Run("SearchAll", func(t *testing.T) {
		// 第一次搜索，缓存结果
		fields := map[string]any{
			"id": nil, // id=nil或空，将获取所有表记录
		}
		dataIter, _ := table.Search(&fields)
		defer GlobalTableIterPool.Put(dataIter)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer GlobalTableIterPool.Put(dataIter)
		records := dataIter.GetRecords(true)
		defer record.PutRecords(records)
		for i, item := range records.Select() {
			fmt.Printf("item %d: %v\n", i, item)
		}
	})

	// 强制垃圾回收，确保 finalizer 被执行
	//runtime.GC()
	//runtime.Gosched() // 让出CPU时间，让 finalizer 有机会执行

	/*
		// 强制垃圾回收，确保 finalizer 被执行
		runtime.GC()
		runtime.Gosched() // 让出CPU时间，让 finalizer 有机会执行
	*/
}

// 测试添加、删除、修改操作
func TestAddDeleteUpdate(t *testing.T) {
	// 创建表
	table, err := TableNew("test_add_delete_update")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 定义表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	// 设置表字段
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 测试1: 添加单条记录
	t.Log("测试1: 添加单条记录")
	insertRecord := map[string]any{
		"id":   1,
		"name": "测试用户",
		"age":  30,
	}

	_, err = table.Insert(&insertRecord)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// 验证记录存在
	searchFields := map[string]any{"id": 1}
	dataIter, _ := table.Search(&searchFields)
	defer GlobalTableIterPool.Put(dataIter)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	records := dataIter.GetRecords(true)
	defer record.PutRecords(records)
	if len(records) == 0 {
		t.Error("Record not found after adding")
	} else {
		if records[0]["name"] != "测试用户" || records[0]["age"] != 30 {
			t.Errorf("Record data mismatch: got %v, expected name=测试用户, age=30", records[0])
		} else {
			t.Logf("添加记录成功: %v", records[0])
		}
	}

	// 测试2: 修改记录
	t.Log("测试2: 修改记录")
	updateRecord := map[string]any{
		"id":   1,
		"name": "修改后的测试用户",
		"age":  31,
	}

	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 验证修改成功
	searchFields = map[string]any{"id": 1}
	dataIter, _ = table.Search(&searchFields)
	defer GlobalTableIterPool.Put(dataIter)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords := dataIter.GetRecords(true)
	defer record.PutRecords(updatedRecords)
	if len(updatedRecords) == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords[0]["name"] != "修改后的测试用户" || updatedRecords[0]["age"] != 31 {
			t.Errorf("Record update failed: got %v, expected name=修改后的测试用户, age=31", updatedRecords[0])
		} else {
			t.Logf("修改记录成功: %v", updatedRecords[0])
		}
	}

	// 测试3: 删除记录
	t.Log("测试3: 删除记录")
	deleteFields := map[string]any{"id": 1}
	err = table.Delete(&deleteFields)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证记录已删除
	searchFields = map[string]any{"id": 1}
	dataIter, _ = table.Search(&searchFields)
	defer GlobalTableIterPool.Put(dataIter)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords := dataIter.GetRecords(true)
	defer record.PutRecords(deletedRecords)
	if len(deletedRecords) == 0 {
		t.Log("删除记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords)
	}

	t.Log("所有添加、删除、修改测试通过")
}

// 测试添加Table.Insert，删除Table.Delete，修改Table.Update，添加一条记录，通过主键进行修改和删除
func TestTableCRUD1(t *testing.T) {

	// 创建表
	table, err := TableNew("test_table_CRUD1")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 定义表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	// 设置表字段 - 使用正确的SetFields方法
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 测试1: 添加单条记录
	t.Log("测试1: 添加单条记录")
	insertRecord := map[string]any{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	_, err = table.Insert(&insertRecord)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// 验证记录存在
	searchFields := map[string]any{"id": 1}
	dataIter, _ := table.Search(&searchFields)
	defer GlobalTableIterPool.Put(dataIter)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	records := dataIter.GetRecords(true)
	defer records.Release()
	if len(records) == 0 {
		t.Error("Record not found after adding")
	} else {
		if records[0]["name"] != "张三" || records[0]["age"] != 25 {
			t.Errorf("Record data mismatch: got %v, expected name=张三, age=25", records[0])
		} else {
			t.Logf("添加记录成功: %v", records[0])
		}
	}

	// 测试2: 修改记录
	t.Log("测试2: 修改记录")
	updateRecord := map[string]any{
		"id":   1,
		"name": "张三修改",
		"age":  26,
	}

	err = table.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// 验证修改成功
	searchFields = map[string]any{"id": 1}
	dataIter, _ = table.Search(&searchFields)
	defer GlobalTableIterPool.Put(dataIter)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords := dataIter.GetRecords(true)
	defer record.PutRecords(updatedRecords)
	if len(updatedRecords) == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords[0]["name"] != "张三修改" || updatedRecords[0]["age"] != 26 {
			t.Errorf("Record update failed: got %v, expected name=张三修改, age=26", updatedRecords[0])
		} else {
			t.Logf("修改记录成功: %v", updatedRecords[0])
		}
	}

	// 测试3: 删除记录
	t.Log("测试3: 删除记录")
	deleteFields := map[string]any{"id": 1}
	err = table.Delete(&deleteFields)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// 验证记录已删除
	searchFields = map[string]any{"id": 1}
	dataIter, _ = table.Search(&searchFields)
	defer GlobalTableIterPool.Put(dataIter)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords := dataIter.GetRecords(true)
	defer record.PutRecords(deletedRecords)
	if len(deletedRecords) == 0 {
		t.Log("删除记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords)
	}

	t.Log("所有CRUD测试通过")

}

// 测试添加Table.Insert，删除Table.Delete，修改Table.Update，添加一条记录，通过主键进行修改和删除
func TestTableCRUD2(t *testing.T) {
	// 创建带索引的表
	tableWithIndex, err := TableNew("test_table_CRUD2")
	if err != nil {
		t.Fatalf("Failed to create table with index: %v", err)
	}

	// 定义带索引的表字段
	indexFields := map[string]any{
		"id":      0,
		"title":   "",
		"content": "",
		"author":  "",
		"views":   0,
	}

	// 设置表字段 - 使用正确的SetFields方法
	err = tableWithIndex.SetFields(indexFields)
	if err != nil {
		t.Fatalf("Failed to set fields for table with index: %v", err)
	}

	// 创建普通索引
	// 创建标题索引
	titleIndex, err := DefaultNormalIndexNew("title_index")
	if err != nil {
		t.Fatalf("Failed to create title index instance: %v", err)
	}
	titleIndex.AddFields("title")
	err = tableWithIndex.CreateIndex(titleIndex)
	//err = tableWithIndex.indexs.CreateIndex(titleIndex)
	if err != nil {
		t.Fatalf("Failed to create title index: %v", err)
	}

	// 创建复合索引（作者-浏览量）
	authorViewsIndex, err := DefaultNormalIndexNew("author_views_index")
	if err != nil {
		t.Fatalf("Failed to create author_views index instance: %v", err)
	}
	authorViewsIndex.AddFields("author", "views")
	err = tableWithIndex.CreateIndex(authorViewsIndex)
	if err != nil {
		t.Fatalf("Failed to create author_views index: %v", err)
	}
	// 添加多条记录
	indexRecords := []map[string]any{
		{
			"id":      1,
			"title":   "Go语言入门",
			"content": "Go语言是一种开源的编程语言，它具有高效、简洁、并发等特点。",
			"author":  "张三",
			"views":   100,
		},

		{
			"id":      2,
			"title":   "Go语言进阶",
			"content": "Go语言的并发模型是其一大特色，使用goroutine和channel可以轻松实现高效的并发编程。",
			"author":  "李四",
			"views":   200,
		},
		{
			"id":      3,
			"title":   "Go语言实战",
			"content": "通过实际项目学习Go语言，可以更好地掌握其特性和最佳实践。",
			"author":  "张三",
			"views":   150,
		},
	}

	for _, record := range indexRecords {
		_, err = tableWithIndex.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to add record to indexed table: %v", err)
		}
	}

	// 测试通过普通索引查询
	t.Log("测试通过普通索引查询")
	searchByTitle := map[string]any{"title": "Go语言入门"}
	dataIter, _ := tableWithIndex.Search(&searchByTitle)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	titleRecords := dataIter.GetRecords(true)
	defer record.PutRecords(titleRecords)
	if len(titleRecords) != 1 {
		t.Errorf("Expected 1 record for title 'Go语言入门', got %d", len(titleRecords))
	} else {
		t.Logf("通过标题索引查询成功: %v", titleRecords[0])
	}

	// 测试通过复合索引查询
	t.Log("测试通过复合索引查询")
	searchByAuthor := map[string]any{"author": "张三"}
	dataIter, _ = tableWithIndex.Search(&searchByAuthor)
	defer GlobalTableIterPool.Put(dataIter)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecords := dataIter.GetRecords(true)
	if len(authorRecords) != 2 {
		t.Errorf("Expected 2 records for author '张三', got %d", len(authorRecords))
	} else {
		t.Logf("通过作者索引查询成功，找到 %d 条记录", len(authorRecords))
	}
	defer record.PutRecords(authorRecords)

	// 测试修改记录
	t.Log("测试修改带索引的记录")
	updateRecord := map[string]any{
		"id":    1,
		"title": "Go语言入门教程",
		"views": 120,
	}

	err = tableWithIndex.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update indexed record: %v", err)
	}

	// 验证修改成功
	searchUpdated := map[string]any{"id": 1}
	dataIter, _ = tableWithIndex.Search(&searchUpdated)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords := dataIter.GetRecords(true)
	defer updatedRecords.Release()
	//defer record.PutRecords(updatedRecords)
	if len(updatedRecords) == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords[0]["title"] != "Go语言入门教程" || updatedRecords[0]["views"] != 120 {
			t.Errorf("Record update failed: got %v, expected title=Go语言入门教程, views=120", updatedRecords[0])
		} else {
			t.Logf("修改带索引记录成功: %v", updatedRecords[0])
		}
	}

}

// 测试添加Table.Insert，删除Table.Delete，修改Table.Update，添加一条记录，通过主键进行修改和删除
func TestTableCRUD3(t *testing.T) {

	// 创建带索引的表
	tableWithIndex, err := TableNew("test_table_with_index3")
	if err != nil {
		t.Fatalf("Failed to create table with index: %v", err)
	}

	// 定义带索引的表字段
	indexFields := map[string]any{
		"id":      0,
		"title":   "",
		"content": "",
		"author":  "",
		"views":   0,
	}

	// 设置表字段 - 使用正确的SetFields方法
	err = tableWithIndex.SetFields(indexFields)
	if err != nil {
		t.Fatalf("Failed to set fields for table with index: %v", err)
	}

	// 创建普通索引
	// 创建标题索引
	titleIndex, err := DefaultNormalIndexNew("title_index")
	if err != nil {
		t.Fatalf("Failed to create title index instance: %v", err)
	}
	titleIndex.AddFields("title")
	err = tableWithIndex.CreateIndex(titleIndex)
	//err = tableWithIndex.indexs.CreateIndex(titleIndex)
	if err != nil {
		t.Fatalf("Failed to create title index: %v", err)
	}

	// 创建复合索引（作者-浏览量）
	authorViewsIndex, err := DefaultNormalIndexNew("author_views_index")
	if err != nil {
		t.Fatalf("Failed to create author_views index instance: %v", err)
	}
	authorViewsIndex.AddFields("author", "views")
	err = tableWithIndex.CreateIndex(authorViewsIndex)
	if err != nil {
		t.Fatalf("Failed to create author_views index: %v", err)
	}

	// 创建全文索引
	contentFulltextIndex, err := DefaultFullTextIndexNew("content_fulltext")
	if err != nil {
		t.Fatalf("Failed to create content fulltext index instance: %v", err)
	}
	//全文索引正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义。
	contentFulltextIndex.AddFields("content", "id")
	//指定description为全文索引字段，长度为5
	//如果没有指定，则等同一般索引
	err = contentFulltextIndex.SetFullField("content", 5)
	if err != nil {
		t.Fatalf("Failed to set fulltext index field: %v", err)
	}

	err = tableWithIndex.CreateIndex(contentFulltextIndex)
	if err != nil {
		t.Fatalf("Failed to create content fulltext index: %v", err)
	}

	// 添加多条记录
	indexRecords := []map[string]any{
		{
			"id":      1,
			"title":   "Go语言入门",
			"content": "Go语言是一种开源的编程语言，它具有高效、简洁、并发等特点。",
			"author":  "张三",
			"views":   100,
		},
		{
			"id":      2,
			"title":   "Go语言进阶",
			"content": "Go语言的并发模型是其一大特色，使用goroutine和channel可以轻松实现高效的并发编程。",
			"author":  "李四",
			"views":   200,
		},
		{
			"id":      3,
			"title":   "Go语言实战",
			"content": "通过实际项目学习Go语言，可以更好地掌握其特性和最佳实践。",
			"author":  "张三",
			"views":   150,
		},
	}

	for _, record := range indexRecords {
		_, err = tableWithIndex.Insert(&record)
		if err != nil {
			t.Fatalf("Failed to add record to indexed table: %v", err)
		}
	}

	// 测试通过普通索引查询
	t.Log("测试通过普通索引查询")
	searchByTitle := map[string]any{"title": "Go语言入门"}
	dataIter, _ := tableWithIndex.Search(&searchByTitle)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	titleRecords := dataIter.GetRecords(true)
	defer record.PutRecords(titleRecords)
	if len(titleRecords) != 1 {
		t.Errorf("Expected 1 record for title 'Go语言入门', got %d", len(titleRecords))
	} else {
		t.Logf("通过标题索引查询成功: %v", titleRecords[0])
	}

	// 测试通过复合索引查询
	t.Log("测试通过复合索引查询")
	searchByAuthor := map[string]any{"author": "张三"}
	dataIter, _ = tableWithIndex.Search(&searchByAuthor)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecords := dataIter.GetRecords(true)
	if len(authorRecords) != 2 {
		t.Errorf("Expected 2 records for author '张三', got %d", len(authorRecords))
	} else {
		t.Logf("通过作者索引查询成功，找到 %d 条记录", len(authorRecords))
	}
	defer record.PutRecords(authorRecords)

	// 测试修改记录
	t.Log("测试修改带索引的记录")
	updateRecord := map[string]any{
		"id":    1,
		"title": "Go语言入门教程",
		"views": 120,
	}

	err = tableWithIndex.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update indexed record: %v", err)
	}

	// 验证修改成功
	searchUpdated := map[string]any{"id": 1}
	dataIter1, _ := tableWithIndex.Search(&searchUpdated)
	defer dataIter1.Release()
	if dataIter1.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords1 := dataIter1.GetRecords(true)
	defer updatedRecords1.Release()
	//defer record.PutRecords(updatedRecords1)
	if len(updatedRecords1) == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords1[0]["title"] != "Go语言入门教程" || updatedRecords1[0]["views"] != 120 {
			t.Errorf("Record update failed: got %v, expected title=Go语言入门教程, views=120", updatedRecords1[0])
		} else {
			t.Logf("修改带索引记录成功: %v", updatedRecords1[0])
		}
	}

	// 测试删除记录
	t.Log("测试删除带索引的记录")
	deleteFields := map[string]any{"id": 3}
	err = tableWithIndex.Delete(&deleteFields)
	if err != nil {
		t.Fatalf("Failed to delete indexed record: %v", err)
	}

	// 验证记录已删除
	searchDeleted := map[string]any{"id": 3}
	dataIter, _ = tableWithIndex.Search(&searchDeleted)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords := dataIter.GetRecords(true)
	defer deletedRecords.Release()
	if len(deletedRecords) == 0 {
		t.Log("删除带索引记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords[0])
	}

	// 验证索引仍然有效
	searchByAuthorAfterDelete := map[string]any{"author": "张三"}
	dataIter, _ = tableWithIndex.Search(&searchByAuthorAfterDelete)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecordsAfterDelete := dataIter.GetRecords(true)
	defer authorRecordsAfterDelete.Release()
	if len(authorRecordsAfterDelete) != 1 {
		t.Errorf("Expected 1 record for author '张三' after delete, got %d", len(authorRecordsAfterDelete))
	} else {
		t.Logf("删除后通过作者索引查询成功，找到 %d 条记录", len(authorRecordsAfterDelete))
	}

	t.Log("所有带索引的CRUD测试通过")
}
