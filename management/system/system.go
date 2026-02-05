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

	// 使用范围查询，只遍历系统键
	sysPrefix := []byte("sys-")
	iter := sm.store.Iterator(sysPrefix, nil)
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

	// 使用范围查询，只遍历指定表的字段键
	prefix := []byte(fmt.Sprintf("sys-%d-field-", tableID))
	iter := sm.store.Iterator(prefix, nil)
	defer iter.Release()

	for iter.First(); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		if strings.HasPrefix(key, fmt.Sprintf("sys-%d-field-", tableID)) {
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

	// 使用范围查询，只遍历指定表的索引键
	prefix := []byte(fmt.Sprintf("sys-%d-idx-", tableID))
	iter := sm.store.Iterator(prefix, nil)
	defer iter.Release()

	for iter.First(); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		if strings.HasPrefix(key, fmt.Sprintf("sys-%d-idx-", tableID)) {
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
	var tables []TableInfo
	tableDetails := make(map[uint8]map[string]interface{})
	fieldsMap := make(map[uint8][]FieldInfo)
	indexesMap := make(map[uint8][]IndexInfo)

	// 使用范围查询，只遍历系统键
	sysPrefix := []byte("sys-")
	iter := sm.store.Iterator(sysPrefix, nil)
	defer iter.Release()

	for iter.First(); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		value := iter.Value()

		// 跳过计数器键（不包含名称部分）
		if strings.Count(key, "-") < 2 {
			continue
		}

		parts := strings.Split(key, "-")
		if len(parts) < 3 {
			continue
		}

		// 处理表信息: sys-table-name
		if parts[0] == "sys" && parts[1] == "table" && len(parts) >= 3 {
			tableName := strings.Join(parts[2:], "-")
			tableID := uint8(value[0])
			tables = append(tables, TableInfo{
				Name: tableName,
				ID:   tableID,
			})
			// 初始化表详情
			if _, ok := tableDetails[tableID]; !ok {
				tableDetails[tableID] = map[string]interface{}{
					"name":    tableName,
					"fields":  []FieldInfo{},
					"indexes": []IndexInfo{},
				}
				fieldsMap[tableID] = []FieldInfo{}
				indexesMap[tableID] = []IndexInfo{}
			}
		} else if parts[0] == "sys" && len(parts) >= 4 {
			// 处理字段信息: sys-tableid-field-name
			if parts[2] == "field" {
				var tableID uint8
				fmt.Sscanf(parts[1], "%d", &tableID)
				fieldName := strings.Join(parts[3:], "-")
				fieldID := uint8(value[0])
				fieldsMap[tableID] = append(fieldsMap[tableID], FieldInfo{
					Name:    fieldName,
					ID:      fieldID,
					TableID: tableID,
				})
			}
			// 处理索引信息: sys-tableid-idx-name
			if parts[2] == "idx" {
				var tableID uint8
				fmt.Sscanf(parts[1], "%d", &tableID)
				indexName := strings.Join(parts[3:], "-")
				indexID := uint8(value[0])
				indexesMap[tableID] = append(indexesMap[tableID], IndexInfo{
					Name:    indexName,
					ID:      indexID,
					TableID: tableID,
				})
			}
		}
	}

	// 组装结果
	for tableID, details := range tableDetails {
		details["fields"] = fieldsMap[tableID]
		details["indexes"] = indexesMap[tableID]
	}

	result := make(map[string]interface{})
	result["tables"] = tables
	result["tableDetails"] = tableDetails

	return result, nil
}
