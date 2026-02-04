package cmd

import (
	"fmt"

	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/storage"
)

var (
	dbPath string
	store  storage.Store
	manager *management.Manager
)

// initStore 初始化数据库连接
func initStore() error {
	var err error
	store, err = storage.OpenDefaultDb(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	manager = management.NewManager(store)
	return nil
}

// closeStore 关闭数据库连接
func closeStore() {
	if store != nil {
		storage.CloseDb()
	}
}
