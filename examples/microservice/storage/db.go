package storage

import (
	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

var (
	localDB *engine.Table
	dbPath  string
)

// InitLocalStorage 初始化本地状态存储
func InitLocalStorage(serviceName string) error {
	// 使用DBManager打开数据库
	dbPath = "./" + serviceName + "_local_db"
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB(dbPath)
	if err != nil {
		return err
	}
	
	// 创建状态表
	table, err := engine.TableNew("service_state")
	if err != nil {
		return err
	}
	
	// 设置字段
	fields := map[string]any{
		"key":    "",  // 状态键
		"value":  "",  // 状态值
		"expire": int64(0),  // 过期时间（ Unix 时间戳）
	}
	err = table.SetFields(fields)
	if err != nil {
		return err
	}
	
	// 创建主键索引
	pk, _ := engine.DefaultPrimaryKeyNew("pk")
	pk.AddFields("key")
	err = table.CreateIndex(pk)
	if err != nil {
		return err
	}
	
	localDB = table
	return nil
}

// GetState 获取本地状态
func GetState(key string) (string, error) {
	fields := map[string]any{"key": key}
	iter, err := localDB.Search(&fields)
	if err != nil {
		return "", err
	}
	defer iter.Release()
	
	records := iter.GetRecords(true)
	defer records.Release()
	
	if len(records) == 0 {
		return "", nil
	}
	
	return records[0]["value"].(string), nil
}

// SetState 设置本地状态
func SetState(key, value string, expire int64) error {
	fields := map[string]any{
		"key":    key,
		"value":  value,
		"expire": expire,
	}
	_, err := localDB.Insert(&fields)
	return err
}

// CloseLocalStorage 关闭本地存储
func CloseLocalStorage() error {
	dbManager := storage.GetDBManager()
	return dbManager.CloseDB()
}

// GetDBPath 获取数据库路径
func GetDBPath() string {
	return dbPath
}
