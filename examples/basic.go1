package main

import (
	"fmt"
	"github.com/liaoran123/sfsDb/api"
	"github.com/liaoran123/sfsDb/engine/types"
)

func main() {
	// 配置数据库
	config := api.Config{
		Path: "./test_db",
	}

	// 打开数据库
	db, err := api.Open(config)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 定义表结构
	userSchema := &types.TableSchema{
		Name: "users",
		Fields: []*types.Field{
			{Name: "id", Type: types.TypeInt64, Nullable: false, Comment: "用户ID"},
			{Name: "name", Type: types.TypeString, Nullable: false, Comment: "用户名"},
			{Name: "email", Type: types.TypeString, Nullable: false, Comment: "邮箱"},
			{Name: "age", Type: types.TypeInt32, Nullable: true, Comment: "年龄"},
		},
		PrimaryKey: []string{"id"},
		Indexes: []*types.IndexDef{
			{
				Name:   "idx_email",
				Type:   types.IndexTypeBTree,
				Fields: []string{"email"},
			},
		},
	}

	// 创建表
	_, err = db.CreateTable("users", userSchema)
	if err != nil {
		panic(err)
	}

	// 列出所有表
	tableNames, err := db.ListTables()
	if err != nil {
		panic(err)
	}

	fmt.Println("Tables:", tableNames)

	// TODO: 实现插入、查询等操作示例
	fmt.Println("Database setup completed successfully!")
}
