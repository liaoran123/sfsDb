package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/liaoran123/sfsDb/examples/microservice/storage"
)

func main() {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "default-service"
	}

	// 初始化本地存储
	err := storage.InitLocalStorage(serviceName)
	if err != nil {
		log.Fatalf("初始化本地存储失败: %v", err)
	}
	defer storage.CloseLocalStorage()

	// 注册路由
	http.HandleFunc("/api/state", handleState)
	http.HandleFunc("/api/health", handleHealth)
	http.HandleFunc("/api/counter", handleCounter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务启动在端口 %s", port)
	log.Printf("本地存储路径: %s", storage.GetDBPath())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleState(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "缺少key参数", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// 获取状态
		value, err := storage.GetState(key)
		if err != nil {
			http.Error(w, fmt.Sprintf("获取状态失败: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"key": "%s", "value": "%s"}`, key, value)
	case http.MethodPost:
		// 设置状态
		value := r.FormValue("value")
		expireStr := r.FormValue("expire")
		expire := int64(0)
		if expireStr != "" {
			expireInt, err := strconv.ParseInt(expireStr, 10, 64)
			if err == nil {
				expire = expireInt
			}
		}
		err := storage.SetState(key, value, expire)
		if err != nil {
			http.Error(w, fmt.Sprintf("设置状态失败: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"key": "%s", "status": "success"}`, key)
	default:
		http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": %d}`, time.Now().Unix())
}

func handleCounter(w http.ResponseWriter, r *http.Request) {
	counterKey := "request_counter"

	switch r.Method {
	case http.MethodGet:
		// 获取计数器值
		value, err := storage.GetState(counterKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("获取计数器失败: %v", err), http.StatusInternalServerError)
			return
		}

		count := 0
		if value != "" {
			count, _ = strconv.Atoi(value)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"counter": %d}`, count)
	case http.MethodPost:
		// 增加计数器
		value, err := storage.GetState(counterKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("获取计数器失败: %v", err), http.StatusInternalServerError)
			return
		}

		count := 0
		if value != "" {
			count, _ = strconv.Atoi(value)
		}
		count++

		err = storage.SetState(counterKey, strconv.Itoa(count), 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("更新计数器失败: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"counter": %d, "status": "success"}`, count)
	default:
		http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
	}
}
