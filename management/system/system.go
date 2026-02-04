package system

import (
	"fmt"
	"strings"

	"github.com/liaoran123/sfsDb/storage"
)

// TableInfo 表信息
type TableInfo struct {
	Name string // 表名
	ID   uint8  // 表ID
}

// FieldInfo 字段信息
type FieldInfo struct {
	Name    string // 字段名
	ID      uint8  // 字段ID
	TableID uint8  // 所属表ID
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name    string // 索引名
	ID      uint8  // 索引ID
	TableID uint8  // 所属表ID
}

// SystemManager 系统信息管理器
type SystemManager struct {
	store storage.Store
}

// NewSystemManager 创建系统信息管理器
// 参数:
//   store: 存储实例
// 返回:
//   *SystemManager: 系统信息管理器实例

func NewSystemManager(store storage.Store) *SystemManager {
	return &SystemManager{
		store: store,
	}
}

// GetAllTables 获取所有表信息
// 返回:
//   []TableInfo: 表信息列表
//   error: 错误信息

func (sm *SystemManager) GetAllTables() ([]TableInfo, error) {
	var tables []TableInfo

	// 遍历存储中的所有键，查找表相关的键
	iter := sm.store.Iterator(nil, nil)
	defer iter.Release()

	for iter.First(); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		if strings.HasPrefix(key, "sys-table-") {
			// 提取表名和ID
			parts := strings.Split(key, "-")
			if len(parts) >= 3 {
				tableName := strings.Join(parts[2:], "-")
				tableID := uint8(iter.Value()[0])
				tables = append(tables, TableInfo{
					Name: tableName,
					ID:   tableID,
				})
			}
		}
	}

	return tables, nil
}

// GetTableFields 获取指定表的所有字段信息
// 参数:
//   tableID: 表ID
// 返回:
//   []FieldInfo: 字段信息列表
//   error: 错误信息

func (sm *SystemManager) GetTableFields(tableID uint8) ([]FieldInfo, error) {
	var fields []FieldInfo

	// 遍历存储中的所有键，查找指定表的字段相关键
	iter := sm.store.Iterator(nil, nil)
	defer iter.Release()

	prefix := fmt.Sprintf("sys-%d-field-", tableID)
	for iter.First(); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		if strings.HasPrefix(key, prefix) {
			// 提取字段名和ID
			parts := strings.Split(key, "-")
			if len(parts) >= 4 {
				fieldName := strings.Join(parts[3:], "-")
				fieldID := uint8(iter.Value()[0])
				fields = append(fields, FieldInfo{
					Name:    fieldName,
					ID:      fieldID,
					TableID: tableID,
				})
			}
		}
	}

	return fields, nil
}

// GetTableIndexes 获取指定表的所有索引信息
// 参数:
//   tableID: 表ID
// 返回:
//   []IndexInfo: 索引信息列表
//   error: 错误信息

func (sm *SystemManager) GetTableIndexes(tableID uint8) ([]IndexInfo, error) {
	var indexes []IndexInfo

	// 遍历存储中的所有键，查找指定表的索引相关键
	iter := sm.store.Iterator(nil, nil)
	defer iter.Release()

	prefix := fmt.Sprintf("sys-%d-idx-", tableID)
	for iter.First(); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		if strings.HasPrefix(key, prefix) {
			// 提取索引名和ID
			parts := strings.Split(key, "-")
			if len(parts) >= 4 {
				indexName := strings.Join(parts[3:], "-")
				indexID := uint8(iter.Value()[0])
				indexes = append(indexes, IndexInfo{
					Name:    indexName,
					ID:      indexID,
					TableID: tableID,
				})
			}
		}
	}

	return indexes, nil
}

// GetAllSystemInfo 获取所有系统信息（表、字段、索引）
// 返回:
//   map[string]interface{}: 系统信息
//   error: 错误信息

func (sm *SystemManager) GetAllSystemInfo() (map[string]interface{}, error) {
	tables, err := sm.GetAllTables()
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	result["tables"] = tables

	// 获取每个表的字段和索引信息
	tableDetails := make(map[uint8]map[string]interface{})
	for _, table := range tables {
		fields, err := sm.GetTableFields(table.ID)
		if err != nil {
			return nil, err
		}

		indexes, err := sm.GetTableIndexes(table.ID)
		if err != nil {
			return nil, err
		}

		tableDetails[table.ID] = map[string]interface{}{
			"name":    table.Name,
			"fields":  fields,
			"indexes": indexes,
		}
	}

	result["tableDetails"] = tableDetails
	return result, nil
}
