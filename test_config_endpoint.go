package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func mainclient() {
	// 创建HTTP客户端
	client := &http.Client{}

	// 创建请求
	req, err := http.NewRequest("GET", "http://localhost:8081/api/config", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// 添加认证头
	req.Header.Add("X-API-Key", "test_api_key_for_development")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// 打印响应
	fmt.Printf("Status code: %d\n", resp.StatusCode)
	fmt.Printf("Response body: %s\n", body)
}
