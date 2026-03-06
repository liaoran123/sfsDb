package engine

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/liaoran123/sfsDb/util"
)

// ExportToCSV 将表数据导出为CSV格式
func (t *Table) ExportToCSV(filePath string) error {
	// 打开文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建CSV文件失败: %v", err)
	}
	defer file.Close()

	// 创建CSV写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 获取所有记录
	iter := t.ForData()
	defer GlobalTableIterPool.Put(iter)

	// 获取字段列表
	fields := make([]string, 0, len(t.fields))
	for field := range t.fields {
		fields = append(fields, field)
	}

	// 写入CSV表头
	if err := writer.Write(fields); err != nil {
		return fmt.Errorf("写入CSV表头失败: %v", err)
	}

	// 获取所有记录
	records := iter.GetRecords(true)

	// 遍历记录并写入CSV
	for _, rec := range records {
		// 将记录转换为CSV行
		row := make([]string, 0, len(fields))
		for _, field := range fields {
			fieldValue, ok := rec[field]
			if !ok {
				row = append(row, "")
				continue
			}

			// 使用util.AnyToStr转换字段值为字符串，保持格式一致性
			strValue := util.AnyToStr(fieldValue)
			row = append(row, strValue)
		}

		// 写入CSV行
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("写入CSV行失败: %v", err)
		}
	}

	return nil
}

// ImportFromCSV 从CSV文件导入数据到表
func (t *Table) ImportFromCSV(filePath string, batchSize int) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	// 创建CSV读取器
	reader := csv.NewReader(file)

	// 读取CSV表头
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("读取CSV表头失败: %v", err)
	}

	// 验证字段名是否存在于表中
	for _, field := range header {
		if _, exists := t.fields[field]; !exists {
			return fmt.Errorf("CSV文件中包含表中不存在的字段: %s", field)
		}
	}

	// 设置默认批量大小
	if batchSize <= 0 {
		batchSize = 100
	}

	// 读取CSV数据并导入
	var records []map[string]any
	for {
		// 读取一行数据
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取CSV行失败: %v", err)
		}

		// 转换为记录
		record := make(map[string]any)
		for i, value := range row {
			if i >= len(header) {
				continue
			}

			field := header[i]
			// 获取字段的默认值作为类型模板
			fieldTemplate := t.fields[field]

			// 使用util.StrToAny转换值为正确的类型
			typedValue, err := util.StrToAny(value, fieldTemplate)
			if err != nil {
				// 如果转换失败，使用默认值
				typedValue = fieldTemplate
			}

			record[field] = typedValue
		}

		records = append(records, record)

		// 批量导入
		if len(records) >= batchSize {
			if err := t.batchInsert(records); err != nil {
				return fmt.Errorf("批量导入数据失败: %v", err)
			}
			records = records[:0] // 清空记录切片
		}
	}

	// 导入剩余数据
	if len(records) > 0 {
		if err := t.batchInsert(records); err != nil {
			return fmt.Errorf("批量导入剩余数据失败: %v", err)
		}
	}

	return nil
}

// ExportToJSON 将表数据导出为JSON格式
func (t *Table) ExportToJSON(filePath string) error {
	// 打开文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建JSON文件失败: %v", err)
	}
	defer file.Close()

	// 创建JSON编码器
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// 获取所有记录
	iter := t.ForData()
	defer GlobalTableIterPool.Put(iter)

	// 开始JSON数组
	if _, err := file.WriteString("[\n"); err != nil {
		return fmt.Errorf("写入JSON数组开始失败: %v", err)
	}

	// 获取所有记录
	records := iter.GetRecords(true)

	// 遍历记录并写入JSON
	firstRecord := true
	for _, rec := range records {
		// 转换记录为map[string]any
		recordMap := make(map[string]any)
		for field, fieldValue := range rec {
			// 转换[]byte为string
			if bytesValue, ok := fieldValue.([]byte); ok {
				recordMap[field] = string(bytesValue)
			} else {
				recordMap[field] = fieldValue
			}
		}

		// 写入JSON记录
		if !firstRecord {
			if _, err := file.WriteString(",\n"); err != nil {
				return fmt.Errorf("写入JSON分隔符失败: %v", err)
			}
		} else {
			firstRecord = false
		}

		if err := encoder.Encode(recordMap); err != nil {
			return fmt.Errorf("写入JSON记录失败: %v", err)
		}
	}

	// 结束JSON数组
	if _, err := file.WriteString("]\n"); err != nil {
		return fmt.Errorf("写入JSON数组结束失败: %v", err)
	}

	return nil
}

// ImportFromJSON 从JSON文件导入数据到表
func (t *Table) ImportFromJSON(filePath string, batchSize int) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开JSON文件失败: %v", err)
	}
	defer file.Close()

	// 解析JSON数组
	var records []map[string]any
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&records); err != nil {
		return fmt.Errorf("解析JSON文件失败: %v", err)
	}

	// 设置默认批量大小
	if batchSize <= 0 {
		batchSize = 100
	}

	// 转换记录中的数值类型为正确的类型
	for i, record := range records {
		for field, value := range record {
			// 获取字段的默认值作为类型模板
			fieldTemplate, exists := t.fields[field]
			if !exists {
				continue
			}

			// 将float64转换为正确的数值类型
			if floatVal, ok := value.(float64); ok {
				// 根据字段模板的类型进行转换
				switch fieldTemplate.(type) {
				case int:
					record[field] = int(floatVal)
				case int8:
					record[field] = int8(floatVal)
				case int16:
					record[field] = int16(floatVal)
				case int32:
					record[field] = int32(floatVal)
				case int64:
					record[field] = int64(floatVal)
				case uint:
					record[field] = uint(floatVal)
				case uint8:
					record[field] = uint8(floatVal)
				case uint16:
					record[field] = uint16(floatVal)
				case uint32:
					record[field] = uint32(floatVal)
				case uint64:
					record[field] = uint64(floatVal)
				case float32:
					record[field] = float32(floatVal)
					// float64类型不需要转换
				}
			}
		}
		records[i] = record
	}

	// 批量导入数据
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		if err := t.batchInsert(batch); err != nil {
			return fmt.Errorf("批量导入数据失败: %v", err)
		}
	}

	return nil
}

// batchInsert 批量插入记录
func (t *Table) batchInsert(records []map[string]any) error {
	// 创建批量操作
	batch := t.kvStore.GetBatch()
	defer batch.Reset()

	for i, record := range records {
		// 检查类型是否匹配
		if err := t.CheckType(&record); err != nil {
			return fmt.Errorf("记录 %d 类型检查失败: %v", i, err)
		}

		// 插入记录
		_, err := t.Insert(&record, batch)
		if err != nil {
			return fmt.Errorf("插入记录 %d 失败: %v", i, err)
		}
	}

	// 提交批量操作
	if err := t.kvStore.WriteBatch(batch); err != nil {
		return fmt.Errorf("提交批量操作失败: %v", err)
	}

	return nil
}

// ExportToSQL 将表数据导出为SQL格式
func (t *Table) ExportToSQL(filePath string) error {
	// 打开文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建SQL文件失败: %v", err)
	}
	defer file.Close()

	// 获取所有记录
	iter := t.ForData()
	defer GlobalTableIterPool.Put(iter)

	// 写入创建表语句
	createSQL := fmt.Sprintf("CREATE TABLE %s (\n", t.name)
	fieldsSQL := make([]string, 0, len(t.fields))
	for field, value := range t.fields {
		var fieldType string
		switch value.(type) {
		case string:
			fieldType = "VARCHAR(255)"
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			fieldType = "INT"
		case float32, float64:
			fieldType = "FLOAT"
		case bool:
			fieldType = "BOOLEAN"
		case time.Time:
			fieldType = "TIMESTAMP"
		default:
			fieldType = "JSON"
		}

		fieldsSQL = append(fieldsSQL, fmt.Sprintf("  %s %s", field, fieldType))
	}

	createSQL += strings.Join(fieldsSQL, ",\n") + "\n);\n\n"
	if _, err := file.WriteString(createSQL); err != nil {
		return fmt.Errorf("写入创建表语句失败: %v", err)
	}

	// 获取所有记录
	records := iter.GetRecords(true)

	// 写入插入语句
	for _, rec := range records {
		// 构建插入语句
		insertSQL := fmt.Sprintf("INSERT INTO %s VALUES (", t.name)
		values := make([]string, 0, len(rec))
		for _, fieldValue := range rec {
			var strValue string
			switch v := fieldValue.(type) {
			case []byte:
				strValue = fmt.Sprintf("'%s'", string(v))
			case string:
				strValue = fmt.Sprintf("'%s'", v)
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				strValue = fmt.Sprintf("%v", v)
			case bool:
				strValue = fmt.Sprintf("%t", v)
			case time.Time:
				strValue = fmt.Sprintf("'%s'", v.Format(time.RFC3339))
			default:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					strValue = fmt.Sprintf("'%v'", v)
				} else {
					strValue = fmt.Sprintf("'%s'", string(jsonBytes))
				}
			}

			values = append(values, strValue)
		}

		insertSQL += strings.Join(values, ", ") + ");\n"
		if _, err := file.WriteString(insertSQL); err != nil {
			return fmt.Errorf("写入插入语句失败: %v", err)
		}
	}

	return nil
}
