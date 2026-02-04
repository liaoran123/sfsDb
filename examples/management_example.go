package main

import (
	"fmt"

	"github.com/liaoran123/sfsDb/test"
)

func main() {
	fmt.Println("开始测试所有管理功能...")
	
	err := test.TestAllManagementFeatures()
	if err != nil {
		fmt.Printf("测试失败: %v\n", err)
		return
	}
	
	fmt.Println("所有管理功能测试成功!")
}
