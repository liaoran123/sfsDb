package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func mainresp() {
	// 发送 GET 请求到配置端点
	resp, err := http.Get("http://localhost:8080/api/config")
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	// 解析响应体
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	// 打印完整响应
	fmt.Println("Complete response:")
	if data, err := json.MarshalIndent(response, "", "  "); err == nil {
		fmt.Println(string(data))
	}

	// 检查 Scenarios 字段
	if scenarios, ok := response["Scenarios"]; ok {
		fmt.Println("\nScenarios field found:")
		fmt.Printf("Type: %T\n", scenarios)
		fmt.Printf("Value: %v\n", scenarios)
	} else {
		fmt.Println("\nScenarios field not found in response")
	}
}
