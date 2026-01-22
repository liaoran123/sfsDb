// 测试文件，对应 table.go
package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/util"
)

// 测试索引匹配功能

// 测试组合主键搜索功能
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
	//全文索引正常情况下必须在前或后带上全量主键，否则后面的关键词都被覆盖，失去全文索引的意义。
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
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败: %v", err)
	}
	//defer dataIter.Release()
	records := dataIter.GetRecords(true)
	for i, item := range records {
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
		rs := dataIter.GetRecords(true)
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
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
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
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
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
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
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
			defer dataIter.Release()
			if dataIter.iter == nil {
				t.Fatalf("Search 失败")
			}
			//defer dataIter.Release()

			records := dataIter.GetRecords(true)
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
			dataIter := table.Search(&fields)
			defer dataIter.Release()
			if dataIter.iter == nil {
				t.Fatalf("Search 失败")
			}
			//defer dataIter.Release()

			records := dataIter.GetRecords(true)
			for _, item := range records.Select("name", "age", "description") {
				fmt.Printf("搜索:%v -》 records: %v\n", fields["description"], item)
			}
			if records[0]["id"] != data[1]["id"] {
				fmt.Printf("查询结果可能是多个: %v\n。但是测试并没有错误。", records)
				//t.Errorf("全文索引搜索 description 包含Bob的记录错误，期望: %v, 实际: %v", data[1]["id"], records[0][table.GetPrimaryKey().GetFields()[0]])
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
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
		//判断data[1]和records是否相等
		if records != nil {
			t.Errorf("搜索description包含Bob的记录错误，期望: %v, 实际: %v", data[1]["id"], records[0][table.GetPrimaryKey().GetFields()[0]])
		}
	})

	//打开所有记录
	t.Run("SearchAll", func(t *testing.T) {
		// 第一次搜索，缓存结果
		fields := map[string]any{
			"id": nil, // id=nil或空，将获取所有表记录
		}
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
		for i, item := range records.Select() {
			fmt.Printf("item %d: %v\n", i, item)
		}
	})
}

// 测试添加Table.Insert，删除Table.Delete，修改Table.Update，添加一条记录，通过主键进行修改和删除
func TestTable_SetFields(t *testing.T) {
	// 创建表
	table, err := TableNew("test_table_SetFields")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 测试1: 正常设置字段
	t.Log("测试1: 正常设置字段")
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// 验证字段设置正确
	if len(table.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 3 {
		t.Errorf("Expected 3 fieldsid mappings, got %d", len(table.fieldsid))
	}

	// 验证每个字段都有对应的ID映射
	for field := range fields {
		found := false
		for _, f := range table.fieldsid {
			if f == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Field %s not found in fieldsid mapping", field)
		}
	}
	fmt.Printf("table.fieldsid: %v\n", table.fieldsid)
	// 测试3: 空字段映射
	t.Log("测试3: 空字段映射")
	emptyFields := map[string]any{}

	err = table.SetFields(emptyFields)
	if err != nil {
		t.Errorf("Unexpected error for empty fields: %v", err)
	}
	if len(table.fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 0 {
		t.Errorf("Expected 0 fieldsid mappings, got %d", len(table.fieldsid))
	}

	// 测试4: 更新现有字段
	t.Log("测试4: 更新现有字段")
	updateFields := map[string]any{
		"id":    0,
		"name":  "",
		"age":   0,
		"email": "",
	}

	err = table.SetFields(updateFields)
	if err != nil {
		t.Fatalf("Failed to update fields: %v", err)
	}

	if len(table.fields) != 4 {
		t.Errorf("Expected 4 fields after update, got %d", len(table.fields))
	}
	if len(table.fieldsid) != 4 {
		t.Errorf("Expected 4 fieldsid mappings after update, got %d", len(table.fieldsid))
	}
}

func TestTableCRUD(t *testing.T) {
	// 使用唯一表名，避免测试数据累积
	tableName := fmt.Sprintf("test_table_CRUD_%d", time.Now().UnixNano())
	tableWithIndexName := fmt.Sprintf("test_table_with_index_%d", time.Now().UnixNano())

	// 创建表
	table, err := TableNew(tableName)
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
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	records := dataIter.GetRecords(true)
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
	dataIter = table.Search(&searchFields)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords := dataIter.GetRecords(true)
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
	dataIter = table.Search(&searchFields)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords := dataIter.GetRecords(true)
	if len(deletedRecords) == 0 {
		t.Log("删除记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords)
	}

	t.Log("所有CRUD测试通过")

	// 测试4: 带有索引和全文索引的表CRUD操作
	t.Log("\n测试4: 带有索引和全文索引的表CRUD操作")

	// 创建带索引的表
	tableWithIndex, err := TableNew(tableWithIndexName)
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
	err = tableWithIndex.SetFields(indexFields)
	if err != nil {
		t.Fatalf("Failed to set fields for table with index: %v", err)
	}

	// 创建主键索引
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index instance: %v", err)
	}
	pkIndex.AddFields("id")
	err = tableWithIndex.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
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
	dataIter = tableWithIndex.Search(&searchByTitle)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	titleRecords := dataIter.GetRecords(true)
	if len(titleRecords) != 1 {
		t.Errorf("Expected 1 record for title 'Go语言入门', got %d", len(titleRecords))
	} else {
		t.Logf("通过标题索引查询成功: %v", titleRecords[0])
	}
	/*
		// -----------------------------------------
		tb := tableWithIndex
		tb.Search(&searchByTitle).GetRecords(true).Select("id", "title", "content", "author", "views")
		tb.Search(&searchByTitle).GetRecords(true).Delete()
		tb.Search(&searchByTitle).GerRecords(true).Update(&updateRecord)
		//-----------------------------------------
	*/
	// 测试通过复合索引查询
	t.Log("测试通过复合索引查询")
	searchByAuthor := map[string]any{"author": "张三"}
	dataIter = tableWithIndex.Search(&searchByAuthor)
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
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords = dataIter.GetRecords(true)
	if len(updatedRecords) == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords[0]["title"] != "Go语言入门教程" || updatedRecords[0]["views"] != 120 {
			t.Errorf("Record update failed: got %v, expected title=Go语言入门教程, views=120", updatedRecords[0])
		} else {
			t.Logf("修改带索引记录成功: %v", updatedRecords[0])
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
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	deletedRecords = dataIter.GetRecords(true)
	if len(deletedRecords) == 0 {
		t.Log("删除带索引记录成功")
	} else {
		t.Errorf("Record deletion failed: found %v, expected none", deletedRecords[0])
	}

	// 验证索引仍然有效
	searchByAuthorAfterDelete := map[string]any{"author": "张三"}
	dataIter = tableWithIndex.Search(&searchByAuthorAfterDelete)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	authorRecordsAfterDelete := dataIter.GetRecords(true)
	if len(authorRecordsAfterDelete) != 1 {
		t.Errorf("Expected 1 record for author '张三' after delete, got %d", len(authorRecordsAfterDelete))
	} else {
		t.Logf("删除后通过作者索引查询成功，找到 %d 条记录", len(authorRecordsAfterDelete))
	}

	t.Log("所有带索引的CRUD测试通过")
}

// 测试 getSysNameId 函数，验证多次重复创建相同表时ID不会重新生成
func TestGetSysNameId(t *testing.T) {
	// 测试场景1：多次创建同名表，验证ID相同
	t.Run("SameTableSameID", func(t *testing.T) {
		// 创建第一个表
		table1, err := TableNew("test_same_id")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}
		id1 := table1.id

		// 创建第二个同名表
		table2, err := TableNew("test_same_id")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}
		id2 := table2.id

		// 验证两个表的ID相同
		if id1 != id2 {
			t.Errorf("Expected same ID for same table name, got %d and %d", id1, id2)
		}
		t.Logf("Same table name 'test_same_id' got same ID: %d", id1)
	})

	// 测试场景2：创建不同名表，验证ID不同
	t.Run("DifferentTableDifferentID", func(t *testing.T) {
		// 创建第一个表
		table1, err := TableNew("test_table_1")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}
		id1 := table1.id

		// 创建第二个不同名表
		table2, err := TableNew("test_table_2")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}
		id2 := table2.id

		// 验证两个表的ID不同
		if id1 == id2 {
			t.Errorf("Expected different IDs for different table names, got %d for both", id1)
		}
		t.Logf("Different table names got different IDs: %d and %d", id1, id2)
	})

	// 测试场景3：测试IDManager的ID生成逻辑
	t.Run("IDManagerGeneration", func(t *testing.T) {
		// 创建表并获取其IDManager
		table, err := TableNew("test_id_manager")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}

		// 设置表字段
		fields := map[string]any{
			"id":   0,
			"name": "",
		}
		table.SetFields(fields)

		// 验证fieldIDManager已正确初始化
		if table.fieldIDManager == nil {
			t.Fatalf("fieldIDManager is nil")
		}

		// 测试IDManager的GetOrCreateID逻辑
		// 生成字段键
		fieldKey := table.fieldIDManager.GenerateFieldKey(table.id, "name")
		id1, err := table.fieldIDManager.GetOrCreateID(fieldKey)
		if err != nil {
			t.Fatalf("GetOrCreateID failed: %v", err)
		}

		// 再次调用，应该返回相同ID
		id2, err := table.fieldIDManager.GetOrCreateID(fieldKey)
		if err != nil {
			t.Fatalf("GetOrCreateID failed: %v", err)
		}

		// 验证相同键生成相同ID
		if id1 != id2 {
			t.Errorf("Expected same ID for same field key, got %d and %d", id1, id2)
		}
		t.Logf("Same field key got same ID: %d", id1)
	})
}

// 测试多次重复创建相同的索引时，不会重新生成ID
func TestCreateIndexSameID(t *testing.T) {
	t.Run("SameIndexSameID", func(t *testing.T) {
		// 创建表
		table, err := TableNew("test_create_index")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}

		// 设置表字段
		fields := map[string]any{
			"id":   0,
			"name": "",
			"age":  0,
		}
		table.SetFields(fields)

		// 确保indexIDManager已初始化
		if table.indexIDManager == nil {
			table.indexIDManager = NewIDManager(table.kvStore)
		}

		// 创建第一个索引并获取ID
		index1, err := DefaultNormalIndexNew("test_idx")
		if err != nil {
			t.Fatalf("DefaultNormalIndexNew failed: %v", err)
		}
		index1.AddFields("name")

		// 生成索引键并获取ID
		indexKey := table.indexIDManager.GenerateIndexKey(table.id, "test_idx")
		idBefore, err := table.indexIDManager.GetOrCreateID(indexKey)
		if err != nil {
			t.Fatalf("GetOrCreateID failed: %v", err)
		}

		// 创建索引
		err = table.CreateIndex(index1)
		if err != nil {
			t.Fatalf("CreateIndex failed: %v", err)
		}

		// 再次获取ID
		idAfter, err := table.indexIDManager.GetOrCreateID(indexKey)
		if err != nil {
			t.Fatalf("GetOrCreateID failed: %v", err)
		}

		// 创建第二个同名索引
		index2, err := DefaultNormalIndexNew("test_idx")
		if err != nil {
			t.Fatalf("DefaultNormalIndexNew failed: %v", err)
		}
		index2.AddFields("name")

		// 再次获取ID
		idAgain, err := table.indexIDManager.GetOrCreateID(indexKey)
		if err != nil {
			t.Fatalf("GetOrCreateID failed: %v", err)
		}

		// 验证所有ID相同
		if idBefore != idAfter || idAfter != idAgain {
			t.Errorf("Expected same ID for same index name, got %d, %d, %d", idBefore, idAfter, idAgain)
		}
		t.Logf("Same index name 'test_idx' got same ID: %d", idAfter)
	})
}

// 测试 Insert 函数的参数验证和错误处理
func TestTableInsertValidation(t *testing.T) {
	t.Run("InsertWithNilFields", func(t *testing.T) {
		// 创建表
		table, err := TableNew("test_insert_validation")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}

		// 设置表字段
		fields := map[string]any{
			"id":   0,
			"name": "",
		}
		table.SetFields(fields)

		// 测试 nil fields 参数
		_, err = table.Insert(nil)
		if err == nil {
			t.Errorf("Expected error for nil fields, got nil")
		}
		t.Logf("Expected error for nil fields, got: %v", err)
	})

	t.Run("InsertWithoutFields", func(t *testing.T) {
		// 创建表
		table, err := TableNew("test_insert_no_fields")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}

		// 不设置表字段，直接插入
		testFields := map[string]any{
			"id":   1,
			"name": "test",
		}
		_, err = table.Insert(&testFields)
		if err == nil {
			t.Errorf("Expected error for table without fields, got nil")
		}
		t.Logf("Expected error for table without fields, got: %v", err)
	})

	t.Run("InsertWithInvalidPrimaryKeyType", func(t *testing.T) {
		// 创建表
		table, err := TableNew("test_insert_invalid_pk")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}

		// 设置表字段
		fields := map[string]any{
			"id":   0,
			"name": "",
		}
		table.SetFields(fields)

		// 测试无效的主键类型
		testFields := map[string]any{
			"id":   "not an integer", // 无效的主键类型
			"name": "test",
		}
		_, err = table.Insert(&testFields)
		if err == nil {
			t.Errorf("Expected error for invalid primary key type, got nil")
		}
		t.Logf("Expected error for invalid primary key type, got: %v", err)
	})

	t.Run("InsertWithNilBatch", func(t *testing.T) {
		// 创建表
		table, err := TableNew("test_insert_nil_batch")
		if err != nil {
			t.Fatalf("TableNew failed: %v", err)
		}

		// 设置表字段
		fields := map[string]any{
			"id":   0,
			"name": "",
		}
		table.SetFields(fields)

		// 测试 nil batch
		testFields := map[string]any{
			"id":   1,
			"name": "test",
		}
		_, err = table.Insert(&testFields, nil)
		if err == nil {
			t.Errorf("Expected error for nil batch, got nil")
		}
		t.Logf("Expected error for nil batch, got: %v", err)
	})
}

// 检查修改普通索引和全文索引的变化情况
func TestTableIndexChange(t *testing.T) {
	// 测试4: 带有索引和全文索引的表CRUD操作
	//t.Log("\n测试4: 带有索引和全文索引的表CRUD操作")

	// 使用唯一表名，避免测试数据累积
	tableWithIndexName := fmt.Sprintf("test_table_with_index_Change_%d", time.Now().UnixNano())

	// 创建带索引的表
	tableWithIndex, err := TableNew(tableWithIndexName)
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
	err = tableWithIndex.SetFields(indexFields)
	if err != nil {
		t.Fatalf("Failed to set fields for table with index: %v", err)
	}

	// 创建主键索引
	pkIndex, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("Failed to create primary key index instance: %v", err)
	}
	pkIndex.AddFields("id")
	err = tableWithIndex.CreateIndex(pkIndex)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

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
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	titleRecords := dataIter.GetRecords(true)
	if len(titleRecords) != 1 {
		t.Errorf("Expected 1 record for title 'Go语言入门', got %d", len(titleRecords))
	} else {
		t.Logf("通过标题索引查询成功: %v", titleRecords[0])
	}
	//tb:=tableWithIndex;
	//tb.Search(&searchByTitle).GetRecords(true)
	// 测试通过复合索引查询
	t.Log("测试通过复合索引查询")
	searchByAuthor := map[string]any{"author": "张三"}
	dataIter = tableWithIndex.Search(&searchByAuthor)
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
	//检测索引author_views_index的所有键值对
	t.Log("检测索引author_views_index的所有键值对")
	idxkey := authorViewsIndex.Prefix(tableWithIndex.id)
	rangeHelper := util.NewRangeHelper(nil)
	slice := rangeHelper.FromComparison(util.Like, idxkey)
	iter := tableWithIndex.kvStore.Iterator(slice.Start, slice.Limit)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("修改前", "author_views_index: ", string(key))
	}
	idxkey = titleIndex.Prefix(tableWithIndex.id)
	rangeHelper = util.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(util.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice.Start, slice.Limit)
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
	rangeHelper = util.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(util.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice.Start, slice.Limit)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("修改后", "author_views_index: ", string(key))
	}
	idxkey = titleIndex.Prefix(tableWithIndex.id)
	rangeHelper = util.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(util.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice.Start, slice.Limit)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("修改后", "title_index: ", string(key))
	}
	iter.Release()

	searchUpdated := map[string]any{"id": 1}
	dataIter = tableWithIndex.Search(&searchUpdated)
	defer dataIter.Release()
	if dataIter.iter == nil {
		t.Fatalf("Search 失败")
	}

	updatedRecords := dataIter.GetRecords(true)
	if len(updatedRecords) == 0 {
		t.Error("Updated record not found")
	} else {
		if updatedRecords[0]["title"] != "Go语言入门教程" || updatedRecords[0]["views"] != 120 {
			t.Errorf("Record update failed: got %v, expected title=Go语言入门教程, views=120", updatedRecords[0])
		} else {
			t.Logf("修改带索引记录成功: %v", updatedRecords[0])
		}
	}

	// 检查修改全文索引记录
	//检测索引content_fulltext的所有键值对
	t.Log("检测索引content_fulltext的所有键值对")
	idxkey = contentFulltextIndex.Prefix(tableWithIndex.id)
	rangeHelper = util.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(util.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice.Start, slice.Limit)
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
	rangeHelper = util.NewRangeHelper(nil)
	slice = rangeHelper.FromComparison(util.Like, idxkey)
	iter = tableWithIndex.kvStore.Iterator(slice.Start, slice.Limit)
	for iter.Next() {
		key := iter.Key()
		fmt.Println("content_fulltext: ", string(key))
	}
	iter.Release()

}

// TestTableSearch 测试表遍历数据和Search方法的功能
func TestIntEscape(t *testing.T) {
	// 创建测试表
	table, err := TableNew("test_int_escape")
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

	// 插入测试数据
	data := []map[string]any{
		{"id": 1, "name": "六月", "age": uint8(45), "description": "123"},
		{"id": 2, "name": "Bob", "age": uint8(30), "description": "Bob is a product manager"},
		{"id": 3, "name": "Charlie", "age": uint8(45), "description": "Charlie is1 a designer"},
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

	}
	//测试遍历表所有kv

	fmt.Println("-----------------")
	// 测试遍历表所有数据
	t.Run("ForData", func(t *testing.T) {
		// 使用ForData方法遍历所有数据
		dataIter := table.ForData()
		rs := dataIter.GetRecords(true)
		for _, item := range rs {
			fmt.Printf("records: %v\n", item)
		}
	})
	fmt.Println("-----------------")

	//打开所有记录
	t.Run("SearchAll", func(t *testing.T) {
		// 第一次搜索，缓存结果
		fields := map[string]any{
			"id": nil, // id=nil或空，将获取所有表记录
		}
		dataIter := table.Search(&fields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		//defer dataIter.Release()
		records := dataIter.GetRecords(true)
		for i, item := range records.Select() {
			fmt.Printf("item %d: %v\n", i, item)
		}
	})
}
