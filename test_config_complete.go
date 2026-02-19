package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func mainconfig() {
	// 创建HTTP客户端
	client := &http.Client{}

	// 测试1: 获取配置
	fmt.Println("测试1: 获取配置")
	config, err := getConfig(client)
	if err != nil {
		fmt.Printf("获取配置失败: %v\n", err)
	} else {
		fmt.Printf("获取配置成功: %s\n", config)
	}

	// 测试2: 设置场景配置
	fmt.Println("\n测试2: 设置场景配置")
	scenario := "iot"
	err = setScenarioConfig(client, scenario)
	if err != nil {
		fmt.Printf("设置场景配置失败: %v\n", err)
	} else {
		fmt.Printf("设置场景配置成功: %s\n", scenario)
	}

	// 测试3: 更新单个配置项
	fmt.Println("\n测试3: 更新单个配置项")
	err = setConfig(client, "write_buffer", "16MB")
	if err != nil {
		fmt.Printf("更新配置失败: %v\n", err)
	} else {
		fmt.Println("更新配置成功")
	}

	// 测试4: 重置配置
	fmt.Println("\n测试4: 重置配置")
	err = resetConfig(client)
	if err != nil {
		fmt.Printf("重置配置失败: %v\n", err)
	} else {
		fmt.Println("重置配置成功")
	}

	// 测试5: 再次获取配置，验证重置是否成功
	fmt.Println("\n测试5: 验证重置后的配置")
	config, err = getConfig(client)
	if err != nil {
		fmt.Printf("获取配置失败: %v\n", err)
	} else {
		fmt.Printf("获取配置成功: %s\n", config)
	}
}

func getConfig(client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", "http://localhost:8081/api/config", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-API-Key", "test_api_key_for_development")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func setScenarioConfig(client *http.Client, scenario string) error {
	data := fmt.Sprintf(`{"scenario": "%s"}`, scenario)
	req, err := http.NewRequest("POST", "http://localhost:8081/api/config/scenario", strings.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Add("X-API-Key", "test_api_key_for_development")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func setConfig(client *http.Client, key, value string) error {
	data := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value)
	req, err := http.NewRequest("POST", "http://localhost:8081/api/config", strings.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Add("X-API-Key", "test_api_key_for_development")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func resetConfig(client *http.Client) error {
	req, err := http.NewRequest("POST", "http://localhost:8081/api/config/reset", nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-API-Key", "test_api_key_for_development")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
