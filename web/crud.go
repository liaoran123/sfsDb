package web

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/util"
)

// TableRequest 表操作请求
type TableRequest struct {
	Name   string         `json:"name"`
	Fields map[string]any `json:"fields"`
}

// RecordRequest 记录操作请求
type RecordRequest struct {
	Fields map[string]any `json:"fields"`
}

// BatchRecordRequest 批量记录操作请求
type BatchRecordRequest struct {
	Records []map[string]any `json:"records"`
}

// convertFieldTypes 根据表结构转换字段类型
func convertFieldTypes(table *engine.Table, fields map[string]any) (map[string]any, error) {
	// 获取表字段定义
	tableFields := table.GetAllFields()

	// 转换每个字段的类型
	for field, value := range fields {
		if fieldValue, exists := tableFields[field]; exists {
			if value != nil {
				// 尝试转换类型
				convertedValue, err := convertType(value, fieldValue)
				if err != nil {
					return nil, fmt.Errorf("字段 '%s' 的类型转换失败: %v", field, err)
				}
				fields[field] = convertedValue
			}
		}
	}

	return fields, nil
}

// convertType 尝试将值转换为目标类型
func convertType(value, targetType any) (any, error) {
	// 如果类型已经匹配，直接返回
	if reflect.TypeOf(value) == reflect.TypeOf(targetType) {
		return value, nil
	}

	// 如果值是字符串，使用 util.StrToAny 函数进行转换
	if strValue, ok := value.(string); ok {
		return util.StrToAny(strValue, targetType)
	}

	// 尝试不同类型的转换
	switch targetType := targetType.(type) {
	case int:
		return convertToInt(value)
	case int8:
		return convertToInt8(value)
	case int16:
		return convertToInt16(value)
	case int32:
		return convertToInt32(value)
	case int64:
		return convertToInt64(value)
	case uint:
		return convertToUint(value)
	case uint8:
		return convertToUint8(value)
	case uint16:
		return convertToUint16(value)
	case uint32:
		return convertToUint32(value)
	case uint64:
		return convertToUint64(value)
	case float32:
		return convertToFloat32(value)
	case float64:
		return convertToFloat64(value)
	case bool:
		return convertToBool(value)
	case string:
		return convertToString(value)
	default:
		// 其他类型不支持转换
		return nil, fmt.Errorf("不支持的类型转换: %T 到 %T", value, targetType)
	}
}

// convertToInt 尝试将值转换为int
func convertToInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("无法转换为int: %T", value)
	}
}

// convertToInt8 尝试将值转换为int8
func convertToInt8(value any) (int8, error) {
	switch v := value.(type) {
	case int:
		return int8(v), nil
	case int8:
		return v, nil
	case int16:
		return int8(v), nil
	case int32:
		return int8(v), nil
	case int64:
		return int8(v), nil
	case uint:
		return int8(v), nil
	case uint8:
		return int8(v), nil
	case uint16:
		return int8(v), nil
	case uint32:
		return int8(v), nil
	case uint64:
		return int8(v), nil
	case float32:
		return int8(v), nil
	case float64:
		return int8(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return int8(i), nil
	default:
		return 0, fmt.Errorf("无法转换为int8: %T", value)
	}
}

// convertToInt16 尝试将值转换为int16
func convertToInt16(value any) (int16, error) {
	switch v := value.(type) {
	case int:
		return int16(v), nil
	case int8:
		return int16(v), nil
	case int16:
		return v, nil
	case int32:
		return int16(v), nil
	case int64:
		return int16(v), nil
	case uint:
		return int16(v), nil
	case uint8:
		return int16(v), nil
	case uint16:
		return int16(v), nil
	case uint32:
		return int16(v), nil
	case uint64:
		return int16(v), nil
	case float32:
		return int16(v), nil
	case float64:
		return int16(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return int16(i), nil
	default:
		return 0, fmt.Errorf("无法转换为int16: %T", value)
	}
}

// convertToInt32 尝试将值转换为int32
func convertToInt32(value any) (int32, error) {
	switch v := value.(type) {
	case int:
		return int32(v), nil
	case int8:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case uint:
		return int32(v), nil
	case uint8:
		return int32(v), nil
	case uint16:
		return int32(v), nil
	case uint32:
		return int32(v), nil
	case uint64:
		return int32(v), nil
	case float32:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return int32(i), nil
	default:
		return 0, fmt.Errorf("无法转换为int32: %T", value)
	}
}

// convertToInt64 尝试将值转换为int64
func convertToInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("无法转换为int64: %T", value)
	}
}

// convertToUint 尝试将值转换为uint
func convertToUint(value any) (uint, error) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint: %d", v)
		}
		return uint(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint: %d", v)
		}
		return uint(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint: %d", v)
		}
		return uint(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint: %d", v)
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint: %d", v)
		}
		return uint(v), nil
	case uint:
		return v, nil
	case uint8:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint: %f", v)
		}
		return uint(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint: %f", v)
		}
		return uint(v), nil
	case string:
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(u), nil
	default:
		return 0, fmt.Errorf("无法转换为uint: %T", value)
	}
}

// convertToUint8 尝试将值转换为uint8
func convertToUint8(value any) (uint8, error) {
	switch v := value.(type) {
	case int:
		if v < 0 || v > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case int8:
		if v < 0 || int(v) > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case int16:
		if v < 0 || int(v) > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case int32:
		if v < 0 || int(v) > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case int64:
		if v < 0 || int(v) > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case uint:
		if v > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case uint8:
		return v, nil
	case uint16:
		if v > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case uint32:
		if v > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case uint64:
		if v > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %d", v)
		}
		return uint8(v), nil
	case float32:
		if v < 0 || v > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %f", v)
		}
		return uint8(v), nil
	case float64:
		if v < 0 || v > 0xFF {
			return 0, fmt.Errorf("值超出uint8范围: %f", v)
		}
		return uint8(v), nil
	case string:
		u, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(u), nil
	default:
		return 0, fmt.Errorf("无法转换为uint8: %T", value)
	}
}

// convertToUint16 尝试将值转换为uint16
func convertToUint16(value any) (uint16, error) {
	switch v := value.(type) {
	case int:
		if v < 0 || v > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %d", v)
		}
		return uint16(v), nil
	case int8:
		if v < 0 || int(v) > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %d", v)
		}
		return uint16(v), nil
	case int16:
		if v < 0 || int(v) > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %d", v)
		}
		return uint16(v), nil
	case int32:
		if v < 0 || int(v) > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %d", v)
		}
		return uint16(v), nil
	case int64:
		if v < 0 || int(v) > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %d", v)
		}
		return uint16(v), nil
	case uint:
		if v > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %d", v)
		}
		return uint16(v), nil
	case uint8:
		return uint16(v), nil
	case uint16:
		return v, nil
	case uint32:
		if v > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %d", v)
		}
		return uint16(v), nil
	case uint64:
		if v > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %d", v)
		}
		return uint16(v), nil
	case float32:
		if v < 0 || v > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %f", v)
		}
		return uint16(v), nil
	case float64:
		if v < 0 || v > 0xFFFF {
			return 0, fmt.Errorf("值超出uint16范围: %f", v)
		}
		return uint16(v), nil
	case string:
		u, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(u), nil
	default:
		return 0, fmt.Errorf("无法转换为uint16: %T", value)
	}
}

// convertToUint32 尝试将值转换为uint32
func convertToUint32(value any) (uint32, error) {
	switch v := value.(type) {
	case int:
		if v < 0 || v > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %d", v)
		}
		return uint32(v), nil
	case int8:
		if v < 0 || int(v) > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %d", v)
		}
		return uint32(v), nil
	case int16:
		if v < 0 || int(v) > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %d", v)
		}
		return uint32(v), nil
	case int32:
		if v < 0 || int(v) > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %d", v)
		}
		return uint32(v), nil
	case int64:
		if v < 0 || int(v) > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %d", v)
		}
		return uint32(v), nil
	case uint:
		if v > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %d", v)
		}
		return uint32(v), nil
	case uint8:
		return uint32(v), nil
	case uint16:
		return uint32(v), nil
	case uint32:
		return v, nil
	case uint64:
		if v > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %d", v)
		}
		return uint32(v), nil
	case float32:
		if v < 0 || v > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %f", v)
		}
		return uint32(v), nil
	case float64:
		if v < 0 || v > 0xFFFFFFFF {
			return 0, fmt.Errorf("值超出uint32范围: %f", v)
		}
		return uint32(v), nil
	case string:
		u, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(u), nil
	default:
		return 0, fmt.Errorf("无法转换为uint32: %T", value)
	}
}

// convertToUint64 尝试将值转换为uint64
func convertToUint64(value any) (uint64, error) {
	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint64: %d", v)
		}
		return uint64(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint64: %d", v)
		}
		return uint64(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint64: %d", v)
		}
		return uint64(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint64: %d", v)
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint64: %d", v)
		}
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint64: %f", v)
		}
		return uint64(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("负数不能转换为uint64: %f", v)
		}
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("无法转换为uint64: %T", value)
	}
}

// convertToFloat32 尝试将值转换为float32
func convertToFloat32(value any) (float32, error) {
	switch v := value.(type) {
	case int:
		return float32(v), nil
	case int8:
		return float32(v), nil
	case int16:
		return float32(v), nil
	case int32:
		return float32(v), nil
	case int64:
		return float32(v), nil
	case uint:
		return float32(v), nil
	case uint8:
		return float32(v), nil
	case uint16:
		return float32(v), nil
	case uint32:
		return float32(v), nil
	case uint64:
		return float32(v), nil
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return 0, err
		}
		return float32(f), nil
	default:
		return 0, fmt.Errorf("无法转换为float32: %T", value)
	}
}

// convertToFloat64 尝试将值转换为float64
func convertToFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("无法转换为float64: %T", value)
	}
}

// convertToBool 尝试将值转换为bool
func convertToBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case float32:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("无法转换为bool: %T", value)
	}
}

// convertToString 尝试将值转换为string
func convertToString(value any) (string, error) {
	return util.AnyToStr(value), nil
}

// handleCreateTable 处理创建表请求
func (s *Server) handleCreateTable(c *gin.Context) {
	var req TableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	if req.Name == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	if len(req.Fields) == 0 {
		s.sendError(c, http.StatusBadRequest, "missing fields", "Table fields are required")
		return
	}

	// 创建表
	table, err := engine.TableNew(req.Name)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table creation failed", "Failed to create table", err.Error())
		return
	}

	// 设置字段
	if err := table.SetFields(req.Fields); err != nil {
		s.sendError(c, http.StatusInternalServerError, "field setting failed", "Failed to set table fields", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"name":    req.Name,
		"id":      table.GetId(),
		"message": "Table created successfully",
	})
}

// handleGetTables 处理获取所有表请求
func (s *Server) handleGetTables(c *gin.Context) {
	systemMgr := s.manager.SystemManager()
	tables, err := systemMgr.GetAllTables()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "tables retrieval failed", "Failed to get tables", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"tables":  tables,
		"message": "Tables retrieved successfully",
	})
}

// handleGetTable 处理获取表信息请求
func (s *Server) handleGetTable(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(name)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 获取表字段
	fields := table.GetFieldsName()

	s.sendSuccess(c, gin.H{
		"name":    name,
		"id":      table.GetId(),
		"fields":  fields,
		"message": "Table retrieved successfully",
	})
}

// handleUpdateTable 处理修改表结构请求
func (s *Server) handleUpdateTable(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	var req TableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(name)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 更新字段
	if err := table.SetFields(req.Fields); err != nil {
		s.sendError(c, http.StatusInternalServerError, "field update failed", "Failed to update table fields", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"name":    name,
		"message": "Table updated successfully",
	})
}

// handleDeleteTable 处理删除表请求
func (s *Server) handleDeleteTable(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(name)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 删除表的所有数据
	if err := table.DeleteAll(); err != nil {
		s.sendError(c, http.StatusInternalServerError, "table deletion failed", "Failed to delete table data", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"name":    name,
		"message": "Table deleted successfully",
	})
}

// handleInsertRecord 处理插入记录请求
func (s *Server) handleInsertRecord(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	var req RecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	if len(req.Fields) == 0 {
		s.sendError(c, http.StatusBadRequest, "missing fields", "Record fields are required")
		return
	}

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(tableName)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 转换字段类型
	convertedFields, err := convertFieldTypes(table, req.Fields)
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "type conversion failed", "Failed to convert field types", err.Error())
		return
	}

	// 插入记录
	id, err := table.Insert(&convertedFields)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "record insertion failed", "Failed to insert record", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"id":      id,
		"message": "Record inserted successfully",
	})
}

// handleGetRecords 处理查询记录请求
func (s *Server) handleGetRecords(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(tableName)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 解析查询参数
	query := make(map[string]any)

	// 处理单个值的查询参数
	if err := c.ShouldBindQuery(&query); err == nil && len(query) > 0 {
		// 使用查询参数构建查询条件
		for k, v := range query {
			query[k] = v
		}
	}

	// 处理多个相同名称的查询参数（如多个device_id或sensor_type）
	for key, values := range c.Request.URL.Query() {
		if len(values) > 1 {
			// 如果有多个相同名称的参数，使用第一个值作为查询条件
			// 注意：这里需要根据实际的索引字段类型来处理
			// 如果索引字段支持数组类型，可以修改为使用数组
			query[key] = values[0]
		}
	}

	// 转换查询参数类型
	if len(query) > 0 {
		convertedQuery, err := convertFieldTypes(table, query)
		if err != nil {
			s.sendError(c, http.StatusBadRequest, "type conversion failed", "Failed to convert query parameter types", err.Error())
			return
		}
		query = convertedQuery
	}

	// 查询记录
	var iter *engine.TableIter
	if len(query) > 0 {
		iter, err = table.Search(&query)
		if err != nil {
			s.sendError(c, http.StatusInternalServerError, "record search failed", "Failed to search records", err.Error())
			return
		}
	} else {
		// 如果没有查询参数，返回所有记录
		iter = table.ForData()
	}
	defer iter.Release()

	// 获取记录集
	records := iter.GetRecords(true)
	defer records.Release()

	s.sendSuccess(c, gin.H{
		"records": records,
		"count":   len(records),
		"message": "Records retrieved successfully",
	})
}

// handleSearchRange 处理范围查询请求
func (s *Server) handleSearchRange(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	// 获取查询参数
	field := c.Query("field")
	if field == "" {
		s.sendError(c, http.StatusBadRequest, "missing field parameter", "Field name is required for range search")
		return
	}

	start := c.Query("start")
	end := c.Query("end")

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(tableName)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 尝试转换范围查询参数类型
	if start != "" || end != "" {
		// 获取表字段定义
		tableFields := table.GetAllFields()

		// 检查字段是否存在
		if fieldValue, exists := tableFields[field]; exists {
			// 转换start值
			if start != "" {
				convertedStart, err := convertType(start, fieldValue)
				if err != nil {
					s.sendError(c, http.StatusBadRequest, "type conversion failed", "Failed to convert start parameter type", err.Error())
					return
				}
				// 将转换后的值转换回字符串，因为SearchRange期望字符串参数
				start = util.AnyToStr(convertedStart)
			}

			// 转换end值
			if end != "" {
				convertedEnd, err := convertType(end, fieldValue)
				if err != nil {
					s.sendError(c, http.StatusBadRequest, "type conversion failed", "Failed to convert end parameter type", err.Error())
					return
				}
				// 将转换后的值转换回字符串，因为SearchRange期望字符串参数
				end = util.AnyToStr(convertedEnd)
			}
		}
	}

	// 执行范围查询
	// 使用SearchRange方法，funIter=nil时会使用默认的kvStore.Iterator
	// 构建Start和Limit参数
	startMap := map[string]any{field: start}
	endMap := map[string]any{field: end}
	tableIter, err := table.SearchRange(nil, &startMap, &endMap)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "range search failed", "Failed to perform range search", err.Error())
		return
	}
	defer tableIter.Release()

	// 获取记录集
	records := tableIter.GetRecords(true)
	defer records.Release()

	s.sendSuccess(c, gin.H{
		"records": records,
		"count":   len(records),
		"field":   field,
		"start":   start,
		"end":     end,
		"message": "Range search completed successfully",
	})
}

// handleUpdateRecord 处理更新记录请求
func (s *Server) handleUpdateRecord(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	var req RecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	if len(req.Fields) == 0 {
		s.sendError(c, http.StatusBadRequest, "missing fields", "Record fields are required")
		return
	}

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(tableName)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 转换字段类型
	convertedFields, err := convertFieldTypes(table, req.Fields)
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "type conversion failed", "Failed to convert field types", err.Error())
		return
	}

	// 更新记录
	if err := table.Update(&convertedFields); err != nil {
		s.sendError(c, http.StatusInternalServerError, "record update failed", "Failed to update record", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"message": "Record updated successfully",
	})
}

// handleDeleteRecord 处理删除记录请求
func (s *Server) handleDeleteRecord(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	var req RecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	if len(req.Fields) == 0 {
		s.sendError(c, http.StatusBadRequest, "missing fields", "Record fields are required")
		return
	}

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(tableName)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 转换字段类型
	convertedFields, err := convertFieldTypes(table, req.Fields)
	if err != nil {
		s.sendError(c, http.StatusBadRequest, "type conversion failed", "Failed to convert field types", err.Error())
		return
	}

	// 删除记录
	if err := table.Delete(&convertedFields); err != nil {
		s.sendError(c, http.StatusInternalServerError, "record deletion failed", "Failed to delete record", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"message": "Record deleted successfully",
	})
}

// handleBatchInsertRecords 处理批量插入记录请求
func (s *Server) handleBatchInsertRecords(c *gin.Context) {
	tableName := c.Param("name")
	if tableName == "" {
		s.sendError(c, http.StatusBadRequest, "missing table name", "Table name is required")
		return
	}

	var req BatchRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	if len(req.Records) == 0 {
		s.sendError(c, http.StatusBadRequest, "missing records", "Records are required")
		return
	}

	// 获取表
	table := &engine.Table{}
	err := table.OpenTable(tableName)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "table retrieval failed", "Failed to get table", err.Error())
		return
	}

	// 转换记录类型并进行字段类型转换
	recordsPtr := make([]*map[string]any, len(req.Records))
	for i, record := range req.Records {
		// 转换字段类型
		convertedRecord, err := convertFieldTypes(table, record)
		if err != nil {
			s.sendError(c, http.StatusBadRequest, "type conversion failed", "Failed to convert field types for record", err.Error())
			return
		}
		recordsPtr[i] = &convertedRecord
	}

	// 批量插入记录
	ids, err := table.BatchInsert(recordsPtr)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "batch insertion failed", "Failed to batch insert records", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"ids":     ids,
		"count":   len(ids),
		"message": "Records inserted successfully",
	})
}
