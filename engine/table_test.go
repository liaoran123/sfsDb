// 测试文件，对应 table.go
package engine

import (
	"fmt"
	"maps"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// 测试索引匹配功能

// 测试组合主键搜索
func TestCompositePrimaryKeySearch(t *testing.T) {
	table, err := TableNew("art")
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}
	// 必须先为表预设字段和数据类型
	fields := map[string]any{
		"mid":     0,  //文章ID或目录ID
		"secNo":   0,  //文章句子序号
		"title":   "", //文章标题
		"content": "", //文章内容
	}
	table.SetFields(fields)
	PrimaryKeys, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew 失败: %v", err)
	}
	PrimaryKeys.AddFields("mid", "secNo") //创建一个mid, secNo的组合主键
	if err := table.CreateIndex(PrimaryKeys); err != nil {
		t.Fatalf("CreateIndex 失败: %v", err)
	}

	fullText, err := DefaultFullTextIndexNew("ft")
	if err != nil {
		t.Fatalf("DefaultFullTextIndexNew 失败: %v", err)
	}
	//全文索引正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义。
	fullText.AddFields("content", "mid", "secNo")
	//指定content为全文索引字段，长度为5
	//如果没有指定，则等同一般索引
	err = fullText.SetFullField("content", 5) //添加content全文索引字段，长度为10
	if err != nil {
		t.Fatalf("SetFullField 失败: %v", err)
	}
	if err = table.CreateIndex(fullText); err != nil {
		t.Fatalf("CreateIndex 失败: %v", err)
	}

	normalIndex, err := DefaultNormalIndexNew("idx")
	if err != nil {
		t.Fatalf("DefaultNormalIndexNew 失败: %v", err)
	}
	//这个是重复索引，不会添加成功
	normalIndex.AddFields("mid", "secNo") //创建一个普通组合索引
	if err = table.CreateIndex(normalIndex); err != nil {
		fmt.Printf("重复索引，不能创建: %v", err)
	}
	//------------------------------

	table.Insert(&map[string]any{
		"mid":     1,
		"secNo":   1,
		"title":   "文章标题11",
		"content": "文章内容11，从三个接口中提取了公共方法，避免了重复定义",
	})
	table.Insert(&map[string]any{
		"mid":     1,
		"secNo":   2,
		"title":   "文章标题12",
		"content": "文章内容12，清晰的层次结构 ：基础接口 + 具体索引类型接口的设计，层次分明",
	})
	table.Insert(&map[string]any{
		"mid":     2,
		"secNo":   1,
		"title":   "文章标题21",
		"content": "文章内容21，更好的可扩展性 ：新索引类型只需嵌入 IndexBase 接口，即可继承公共方法",
	})
	table.Insert(&map[string]any{
		"mid":     2,
		"secNo":   2,
		"title":   "文章标题22",
		"content": "文章内容22，高度可定制化 ：每个索引类型都可以根据需求定制索引字段和行为",
	})

	// 搜索指定文章的所有句子
	fields1 := map[string]any{
		"mid":   nil, // id=nil或空，将获取所有表记录
		"secNo": nil, // id=nil或空，将获取所有表记录
	}

	iter := table.For()
	for iter.Next() {
		fmt.Printf("iter.Key(): %v,iter.Value(): %v\n", string(iter.Key()), string(iter.Value()))
	}

	dataIter := table.Search(&fields1)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败: %v", err)
	}
	defer dataIter.Release()
	records := dataIter.GerRecords(true)
	for i, item := range records.Select() {
		fmt.Printf("结果集 %d: %v\n", i, item)
	}

}

// TestTableSearch 测试表遍历数据和Search方法的功能
func TestTableSearch(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_search")
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
		{"id": nil, "name": "Eve 49", "age": uint8(49), "description": "Eve is3 a manager 49"}, //"id": nil 使用自动增值
		{"id": nil, "name": "Eve 55", "age": uint8(55), "description": "Eve is4 a manager 55"}, //"id": nil 使用自动增值
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
			}*/
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
		rs := dataIter.GerRecords(true)
		for _, item := range rs.Select() {
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
		dataIter := table.Search(&fields)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		defer dataIter.Release()
		records := dataIter.GerRecords(true)
		fmt.Printf("records: %v\n", records)
		//判断data[0]和records是否相等

		if records.Get(0)["name"] != data[0]["name"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["name"], records.Get(0)["name"])
		}
		if records.Get(0)["age"] != data[0]["age"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["age"], records.Get(0)["age"])
		}
		if records.Get(0)["description"] != data[0]["description"] {
			t.Errorf("搜索主键为1的记录错误，期望: %v, 实际: %v", data[0]["description"], records.Get(0)["description"])
		}

	})

	// 测试2: 索引字段搜索
	t.Run("IndexSearch", func(t *testing.T) {
		// 使用Search方法搜索name为"Charlie"的记录
		fields := map[string]any{
			"name": "Charlie",
		}
		dataIter := table.Search(&fields)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		defer dataIter.Release()
		records := dataIter.GerRecords(true)
		for _, item := range records.Select("name", "age", "description") {
			fmt.Printf("records: %v\n", item)
		}

		if records.Get(0)["age"] != data[2]["age"] {
			t.Errorf("搜索name为Charlie的记录错误，期望: %v, 实际: %v", data[2]["age"], records.Get(0)["age"])
		}
	})

	// 测试3: 全文索引搜索
	t.Run("FullTextSearch", func(t *testing.T) {
		// 使用Search方法搜索description包含"Bob"的记录
		fields := map[string]any{
			"description": "Bob",
		}
		dataIter := table.Search(&fields)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		defer dataIter.Release()
		records := dataIter.GerRecords(true)
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
			dataIter := table.Search(&fields)
			if dataIter.iter == nil {
				t.Fatalf("Search 失败")
			}
			defer dataIter.Release()

			records := dataIter.GerRecords(true)
			for _, item := range records.Select("name", "age", "description") {
				fmt.Printf("搜索:%v -》 records: %v\n", fields["description"], item)
			}

			//判断data[1]和records是否相等
			if records.Get(0)["name"] != data[0]["name"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["name"], records.Get(0)["name"])
			}
			if records.Get(0)["age"] != data[0]["age"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["age"], records.Get(0)["age"])
			}
			if records.Get(0)["description"] != data[0]["description"] {
				t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[0]["description"], records.Get(0)["description"])
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
			dataIter := table.Search(&fields)
			if dataIter.iter == nil {
				t.Fatalf("Search 失败")
			}
			defer dataIter.Release()

			records := dataIter.GerRecords(true)
			for _, item := range records.Select("name", "age", "description") {
				fmt.Printf("搜索:%v -》 records: %v\n", fields["description"], item)
			}
			if records.Get(0)[table.GetPrimaryKey().GetFields()[0]] != data[1]["id"] {
				fmt.Printf("查询结果可能是多个: %v\n。但是测试并没有错误。", records)
				//t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[1]["id"], records.Get(0)[table.indexs.GetPrimaryKey().GetFields()[0]])
			}
		}
	})

	// 测试4: 搜索不存在的数据
	t.Run("SearchNonExistent", func(t *testing.T) {
		fields := map[string]any{
			"id": 100,
		}
		// 使用Search方法搜索不存在的id
		dataIter := table.Search(&fields)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		defer dataIter.Release()
		records := dataIter.GerRecords(true)
		//判断data[1]和records是否相等
		if records != nil {
			t.Errorf("搜索description包含Bob的记录错误，期望: %v, 实际: %v", data[1]["id"], records.Get(0)[table.GetPrimaryKey().GetFields()[0]])
		}
	})

	//打开所有记录
	t.Run("SearchAll", func(t *testing.T) {
		// 第一次搜索，缓存结果
		fields := map[string]any{
			"id": nil, // id=nil或空，将获取所有表记录
		}
		dataIter := table.Search(&fields)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		defer dataIter.Release()
		records := dataIter.GerRecords(true)
		for i, item := range records.Select() {
			fmt.Printf("item %d: %v\n", i, item)
		}
	})
}

// 测试添加Table.Insert，删除Table.Delete，修改Table.Update，添加一条记录，通过主键进行修改和删除
func TestTableCRUD(t *testing.T) {

	// 创建表
	table, err := TableNew("test_table_CRUD")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 定义表字段
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	// 设置表字段 - 逐个字段复制
	maps.Copy(table.fields, fields)

	// 测试1: 添加单条记录
	t.Log("测试1: 添加单条记录")
	record := map[string]any{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	_, err = table.Insert(&record)
	if err != nil {
		t.Fatalf("Failed to add record: %v", err)
	}

	// 验证记录存在
	searchFields := map[string]any{"id": 1}
	dataIter := table.Search(&searchFields)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	records := dataIter.GerRecords(true)
	if records.Len() == 0 {
		t.Error("Record not found after adding")
	} else {
		if records.Get(0)["name"] != "张三" || records.Get(0)["age"] != 25 {
			t.Errorf("Record data mismatch: got %v, expected name=张三, age=25", records.Get(0))
		} else {
			t.Logf("添加记录成功: %v", records.Get(0))
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
	dataIter = table.Search(&searchFields)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords := dataIter.GerRecords(true)
	if updatedRecords.Len() == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords.Get(0)["name"] != "张三修改" || updatedRecords.Get(0)["age"] != 26 {
			t.Errorf("Record update failed: got %v, expected name=张三修改, age=26", updatedRecords.Get(0))
		} else {
			t.Logf("修改记录成功: %v", updatedRecords.Get(0))
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
	dataIter = table.Search(&searchFields)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords := dataIter.GerRecords(true)
	if deletedRecords == nil {
		t.Log("删除记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords)
	}

	t.Log("所有CRUD测试通过")

	// 测试4: 带有索引和全文索引的表CRUD操作
	t.Log("\n测试4: 带有索引和全文索引的表CRUD操作")

	// 创建带索引的表
	tableWithIndex, err := TableNew("test_table_with_index")
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

	// 设置表字段
	maps.Copy(tableWithIndex.fields, indexFields)

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
	dataIter = tableWithIndex.Search(&searchByTitle)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	titleRecords := dataIter.GerRecords(true)
	if titleRecords.Len() != 1 {
		t.Errorf("Expected 1 record for title 'Go语言入门', got %d", titleRecords.Len())
	} else {
		t.Logf("通过标题索引查询成功: %v", titleRecords.Get(0))
	}
	/*
		// -----------------------------------------
		tb := tableWithIndex
		tb.Search(&searchByTitle).GerRecords(true).Select("id", "title", "content", "author", "views")
		tb.Search(&searchByTitle).GerRecords(true).Delete()
		tb.Search(&searchByTitle).GerRecords(true).Update(&updateRecord)
		//-----------------------------------------
	*/
	// 测试通过复合索引查询
	t.Log("测试通过复合索引查询")
	searchByAuthor := map[string]any{"author": "张三"}
	dataIter = tableWithIndex.Search(&searchByAuthor)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecords := dataIter.GerRecords(true)
	if authorRecords.Len() != 2 {
		t.Errorf("Expected 2 records for author '张三', got %d", authorRecords.Len())
	} else {
		t.Logf("通过作者索引查询成功，找到 %d 条记录", authorRecords.Len())
	}

	// 测试修改记录
	t.Log("测试修改带索引的记录")
	updateRecord = map[string]any{
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
	dataIter = tableWithIndex.Search(&searchUpdated)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords = dataIter.GerRecords(true)
	if updatedRecords.Len() == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords.Get(0)["title"] != "Go语言入门教程" || updatedRecords.Get(0)["views"] != 120 {
			t.Errorf("Record update failed: got %v, expected title=Go语言入门教程, views=120", updatedRecords.Get(0))
		} else {
			t.Logf("修改带索引记录成功: %v", updatedRecords.Get(0))
		}
	}

	// 测试删除记录
	t.Log("测试删除带索引的记录")
	deleteFields = map[string]any{"id": 3}
	err = tableWithIndex.Delete(&deleteFields)
	if err != nil {
		t.Fatalf("Failed to delete indexed record: %v", err)
	}

	// 验证记录已删除
	searchDeleted := map[string]any{"id": 3}
	dataIter = tableWithIndex.Search(&searchDeleted)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords = dataIter.GerRecords(true)
	if deletedRecords == nil {
		t.Log("删除带索引记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords.Get(0))
	}

	// 验证索引仍然有效
	searchByAuthorAfterDelete := map[string]any{"author": "张三"}
	dataIter = tableWithIndex.Search(&searchByAuthorAfterDelete)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecordsAfterDelete := dataIter.GerRecords(true)
	if authorRecordsAfterDelete.Len() != 1 {
		t.Errorf("Expected 1 record for author '张三' after delete, got %d", authorRecordsAfterDelete.Len())
	} else {
		t.Logf("删除后通过作者索引查询成功，找到 %d 条记录", authorRecordsAfterDelete.Len())
	}

	t.Log("所有带索引的CRUD测试通过")
}

// 检查修改普通索引和全文索引的变化情况
func TestTableIndexChange(t *testing.T) {
	// 测试4: 带有索引和全文索引的表CRUD操作
	//t.Log("\n测试4: 带有索引和全文索引的表CRUD操作")

	// 创建带索引的表
	tableWithIndex, err := TableNew("test_table_with_index_Change")
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

	// 设置表字段
	maps.Copy(tableWithIndex.fields, indexFields)

	// 创建普通索引
	// 创建标题索引
	titleIndex, err := DefaultNormalIndexNew("title_index")
	if err != nil {
		t.Fatalf("Failed to create title index instance: %v", err)
	}
	titleIndex.AddFields("title")
	err = tableWithIndex.CreateIndex(titleIndex)
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
			"content": "qqqqq",
			"author":  "张三",
			"views":   100,
		},
		{
			"id":      2,
			"title":   "Go语言进阶",
			"content": "abcdef",
			"author":  "李四",
			"views":   200,
		},
		{
			"id":      3,
			"title":   "Go语言实战",
			"content": "67890",
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
	dataIter := tableWithIndex.Search(&searchByTitle)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	titleRecords := dataIter.GerRecords(true)
	if titleRecords.Len() != 1 {
		t.Errorf("Expected 1 record for title 'Go语言入门', got %d", titleRecords.Len())
	} else {
		t.Logf("通过标题索引查询成功: %v", titleRecords.Get(0))
	}
	//tb:=tableWithIndex;
	//tb.Search(&searchByTitle).GerRecords(true)
	// 测试通过复合索引查询
	t.Log("测试通过复合索引查询")
	searchByAuthor := map[string]any{"author": "张三"}
	dataIter = tableWithIndex.Search(&searchByAuthor)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecords := dataIter.GerRecords(true)
	if authorRecords.Len() != 2 {
		t.Errorf("Expected 2 records for author '张三', got %d", authorRecords.Len())
	} else {
		t.Logf("通过作者索引查询成功，找到 %d 条记录", authorRecords.Len())
	}
	//检测索引author_views_index的所有键值对
	t.Log("检测索引author_views_index的所有键值对")
	idxkey := authorViewsIndex.Prefix(tableWithIndex.id)
	rangeHelper := storage.NewRangeHelper(nil)
	slice := rangeHelper.FromComparison(storage.Like, idxkey)
	iter := tableWithIndex.kvStore.Iterator(slice)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("修改前", "author_views_index: ", string(key))
	}
	idxkey = titleIndex.Prefix(tableWithIndex.id)
	rangeHelper = storage.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(storage.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("修改前", "title_index: ", string(key))
	}
	iter.Release()
	// 测试修改普通索引记录
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
	fmt.Println("------------------------------------------")
	// 验证修改成功
	idxkey = authorViewsIndex.Prefix(tableWithIndex.id)
	rangeHelper = storage.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(storage.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("修改后", "author_views_index: ", string(key))
	}
	idxkey = titleIndex.Prefix(tableWithIndex.id)
	rangeHelper = storage.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(storage.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("修改后", "title_index: ", string(key))
	}
	iter.Release()

	searchUpdated := map[string]any{"id": 1}
	dataIter = tableWithIndex.Search(&searchUpdated)
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords := dataIter.GerRecords(true)
	if updatedRecords.Len() == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords.Get(0)["title"] != "Go语言入门教程" || updatedRecords.Get(0)["views"] != 120 {
			t.Errorf("Record update failed: got %v, expected title=Go语言入门教程, views=120", updatedRecords.Get(0))
		} else {
			t.Logf("修改带索引记录成功: %v", updatedRecords.Get(0))
		}
	}

	// 检查修改全文索引记录
	//检测索引content_fulltext的所有键值对
	t.Log("检测索引content_fulltext的所有键值对")
	idxkey = contentFulltextIndex.Prefix(tableWithIndex.id)
	rangeHelper = storage.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(storage.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("content_fulltext: ", string(key))
	}
	t.Log("测试修改全文索引记录")
	updateRecord = map[string]any{
		"id":      1,
		"content": "oooooo",
	}
	err = tableWithIndex.Update(&updateRecord)
	if err != nil {
		t.Fatalf("Failed to update indexed record: %v", err)
	}
	// 验证修改成功
	idxkey = contentFulltextIndex.Prefix(tableWithIndex.id)
	rangeHelper = storage.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(storage.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("content_fulltext: ", string(key))
	}
	iter.Release()

}
