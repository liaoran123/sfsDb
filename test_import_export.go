package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func mainexport() {
	// 创建HTTP客户端
	client := &http.Client{}

	// 测试1: 创建测试表
	fmt.Println("测试1: 创建测试表")
	createTestTable(client)

	// 测试2: 导出表数据为JSON
	fmt.Println("\n测试2: 导出表数据为JSON")
	exportTable(client, "test_table", "json")

	// 测试3: 导出表数据为CSV
	fmt.Println("\n测试3: 导出表数据为CSV")
	exportTable(client, "test_table", "csv")

	// 测试4: 导入表数据
	fmt.Println("\n测试4: 导入表数据")
	importTable(client, "test_table", "./exports/test_table_20260219_123654.json")
}

// 创建测试表
func createTestTable(client *http.Client) {
	data := `{"name": "test_table", "fields": {"id": 0, "name": "", "age": 0, "score": 0}}`
	req, err := http.NewRequest("POST", "http://localhost:8083/api/tables", strings.NewReader(data))
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}
	req.Header.Add("X-API-Key", "test_api_key_for_development")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("创建表响应: %s\n", body)
}

// 插入测试数据
func insertTestData(client *http.Client) {
	// 插入第一条数据
	data1 := `{"fields": {"id": 1, "name": "Alice", "age": 25, "score": 85}}`
	insertRecord(client, "test_table", data1)

	// 插入第二条数据
	data2 := `{"fields": {"id": 2, "name": "Bob", "age": 30, "score": 90}}`
	insertRecord(client, "test_table", data2)

	// 插入第三条数据
	data3 := `{"fields": {"id": 3, "name": "Charlie", "age": 35, "score": 95}}`
	insertRecord(client, "test_table", data3)
}

// 插入记录
func insertRecord(client *http.Client, tableName, data string) {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8083/api/tables/%s/records", tableName), strings.NewReader(data))
	if err != nil {
		fmt.Printf("插入记录失败: %v\n", err)
		return
	}
	req.Header.Add("X-API-Key", "test_api_key_for_development")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("插入记录失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("插入记录响应: %s\n", body)
}

// 导出表数据
func exportTable(client *http.Client, tableName, format string) {
	data := fmt.Sprintf(`{"tableName": "%s", "format": "%s"}`, tableName, format)
	req, err := http.NewRequest("POST", "http://localhost:8083/api/tables/export", strings.NewReader(data))
	if err != nil {
		fmt.Printf("导出表失败: %v\n", err)
		return
	}
	req.Header.Add("X-API-Key", "test_api_key_for_development")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("导出表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("导出表响应: %s\n", body)
}

// 导入表数据
func importTable(client *http.Client, tableName, importPath string) {
	data := fmt.Sprintf(`{"tableName": "%s", "importPath": "%s"}`, tableName, importPath)
	req, err := http.NewRequest("POST", "http://localhost:8083/api/tables/import", strings.NewReader(data))
	if err != nil {
		fmt.Printf("导入表失败: %v\n", err)
		return
	}
	req.Header.Add("X-API-Key", "test_api_key_for_development")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("导入表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("导入表响应: %s\n", body)
}
