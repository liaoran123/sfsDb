package main

import (
	"fmt"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

func main() {
	fmt.Println("SFSDB Advanced Example")
	fmt.Println("=====================")

	// 1. 初始化数据库
	fmt.Println("\n1. 初始化数据库")
	storage.OpenDefaultDb("./advanced_example_db")
	defer storage.CloseDb()

	// 2. 创建文章表（使用组合主键）
	fmt.Println("\n2. 创建文章表（使用组合主键）")
	articleTable, err := engine.TableNew("articles")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 3. 设置字段
	fmt.Println("\n3. 设置字段")
	articleFields := map[string]any{
		"article_id":   0,  // 文章ID
		"section_no":   0,  // 章节序号
		"title":        "", // 文章标题
		"content":      "", // 文章内容
		"author":       "", // 作者
		"word_count":   0,  // 字数
		"publish_date": "", // 发布日期
	}
	err = articleTable.SetFields(articleFields)
	if err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	// 4. 创建组合主键索引
	fmt.Println("\n4. 创建组合主键索引")
	compositePk, err := engine.DefaultPrimaryKeyNew("composite_pk")
	if err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}
	compositePk.AddFields("article_id", "section_no") // 组合主键
	err = articleTable.CreateIndex(compositePk)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 5. 创建全文索引
	fmt.Println("\n5. 创建全文索引")
	fullTextIdx, err := engine.DefaultFullTextIndexNew("content_ft")
	if err != nil {
		fmt.Printf("创建全文索引失败: %v\n", err)
		return
	}
	fullTextIdx.AddFields("content", "article_id", "section_no")
	err = fullTextIdx.SetFullField("content", 5) // 设置content为全文索引字段
	if err != nil {
		fmt.Printf("设置全文索引字段失败: %v\n", err)
		return
	}
	err = articleTable.CreateIndex(fullTextIdx)
	if err != nil {
		fmt.Printf("创建全文索引失败: %v\n", err)
		return
	}

	// 6. 插入文章数据
	fmt.Println("\n6. 插入文章数据")
	articles := []map[string]any{
		{"article_id": 1, "section_no": 1, "title": "Go语言入门", "content": "Go语言是一种开源的编程语言，它能让构造简单、可靠且高效的软件变得容易。", "author": "张三", "word_count": 50},
		{"article_id": 1, "section_no": 2, "title": "Go语言入门", "content": "Go语言的语法接近C语言，但对于变量的声明有所不同。Go语言支持垃圾回收功能。", "author": "张三", "word_count": 60},
		{"article_id": 2, "section_no": 1, "title": "数据库基础", "content": "数据库是按照数据结构来组织、存储和管理数据的仓库。", "author": "李四", "word_count": 40},
		{"article_id": 2, "section_no": 2, "title": "数据库基础", "content": "关系型数据库是基于关系模型的数据库，比如MySQL、PostgreSQL等。", "author": "李四", "word_count": 55},
		{"article_id": 3, "section_no": 1, "title": "网络编程", "content": "网络编程是指编写运行在多个设备（计算机）之间的程序。", "author": "王五", "word_count": 45},
	}

	for _, article := range articles {
		_, err := articleTable.Insert(&article)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
		fmt.Printf("插入文章章节成功, ArticleID: %d, SectionNo: %d\n", article["article_id"], article["section_no"])
	}

	// 7. 组合主键查询
	fmt.Println("\n7. 组合主键查询")
	// 查询特定文章的特定章节
	iter := articleTable.Search(&map[string]any{"article_id": 1, "section_no": 1})
	defer iter.Release()
	records := iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("查询结果: %v\n", records[0])
	}

	// 8. 全文搜索
	fmt.Println("\n8. 全文搜索")
	// 搜索包含"Go语言"的内容
	fullTextIter := articleTable.Search(&map[string]any{"content": "Go语言"})
	defer fullTextIter.Release()
	fullTextRecords := fullTextIter.GetRecords(true)
	fmt.Printf("全文搜索结果数: %d\n", len(fullTextRecords))
	for i, record := range fullTextRecords {
		fmt.Printf("结果 %d: 文章ID=%d, 章节=%d, 内容: %s\n", i+1, record["article_id"], record["section_no"], record["content"])
	}

	// 9. 批量操作
	fmt.Println("\n9. 批量操作")
	// 创建批量更新数据
	updateData := map[string]any{
		"article_id": 1,      // 用于定位文章
		"section_no": 1,      // 用于定位章节
		"author":     "张三教授", // 更新作者
		"word_count": 55,     // 更新字数
	}
	err = articleTable.Update(&updateData)
	if err != nil {
		fmt.Printf("更新数据失败: %v\n", err)
		return
	}
	fmt.Println("批量更新成功")

	// 验证更新
	iter = articleTable.Search(&map[string]any{"article_id": 1, "section_no": 1})
	defer iter.Release()
	records = iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("更新后的数据: %v\n", records[0])
	}

	// 10. 范围查询
	fmt.Println("\n10. 范围查询")
	// 查询文章ID大于1的所有章节
	rangeIter := articleTable.Search(&map[string]any{"article_id": 1}, util.GreaterThan)
	defer rangeIter.Release()
	rangeRecords := rangeIter.GetRecords(true)
	fmt.Printf("文章ID大于1的章节数: %d\n", len(rangeRecords))
	for i, record := range rangeRecords {
		fmt.Printf("章节 %d: 文章ID=%d, 章节=%d, 标题=%s\n", i+1, record["article_id"], record["section_no"], record["title"])
	}

}
