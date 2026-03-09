package engine

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/liaoran123/sfsDb/record"
	"github.com/liaoran123/sfsDb/storage"
)

func TestEdgeXTable(t *testing.T) {
	dbPath := "./test_edgex_table_db"
	os.RemoveAll(dbPath)
	defer os.RemoveAll(dbPath)

	_, err := storage.GetDBManager().OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB 失败: %v", err)
	}
	defer storage.GetDBManager().CloseDB()

	table, err := TableNew("edgex_readings")
	if err != nil {
		t.Fatalf("TableNew 失败: %v", err)
	}

	fields := map[string]any{
		"id":         "",
		"deviceName": "",
		"reading":    "",
		"value":      0.0,
		"valueType":  "",
		"baseType":   "",
		"timestamp":  int64(0),
		"metadata":   "",
	}
	table.SetFields(fields)

	PrimaryKeys, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		t.Fatalf("DefaultPrimaryKeyNew 失败: %v", err)
	}
	PrimaryKeys.AddFields("deviceName", "timestamp")
	table.CreateIndex(PrimaryKeys)

	/*util.SetDefaultTypeSize(7)  //设置所有表。
	 */
	fieldTypeLen := map[string]uint8{
		"deviceName": 7,
	}
	table.GetfieldTypeLen(&fieldTypeLen) //设置单个表

	deviceNameIndex, err := DefaultNormalIndexNew("idx_deviceName")
	if err != nil {
		t.Fatalf("DefaultNormalIndexNew 失败: %v", err)
	}
	deviceNameIndex.AddFields("deviceName")
	err = table.CreateIndex(deviceNameIndex)
	if err != nil {
		t.Fatalf("CreateIndex deviceName 失败: %v", err)
	}

	timestampIndex, err := DefaultNormalIndexNew("idx_timestamp")
	if err != nil {
		t.Fatalf("DefaultNormalIndexNew 失败: %v", err)
	}
	timestampIndex.AddFields("timestamp")
	err = table.CreateIndex(timestampIndex)
	if err != nil {
		t.Fatalf("CreateIndex timestamp 失败: %v", err)
	}

	baseTime := time.Now()

	testData := []map[string]any{
		{
			"id":         "reading001",
			"deviceName": "dev0001",
			"reading":    "temperature",
			"value":      25.5,
			"valueType":  "Float64",
			"baseType":   "Number",
			"timestamp":  baseTime.UnixNano(),
			"metadata":   `{"unit": "°C"}`,
		},
		{
			"id":         "reading002",
			"deviceName": "dev0001",
			"reading":    "temperature",
			"value":      26.2,
			"valueType":  "Float64",
			"baseType":   "Number",
			"timestamp":  baseTime.Add(1 * time.Minute).UnixNano(),
			"metadata":   `{"unit": "°C"}`,
		},
		{
			"id":         "reading003",
			"deviceName": "dev0002",
			"reading":    "count",
			"value":      1024.0,
			"valueType":  "Int64",
			"baseType":   "Number",
			"timestamp":  baseTime.UnixNano(),
			"metadata":   `{"unit": "count"}`,
		},
		{
			"id":         "reading004",
			"deviceName": "dev0003",
			"reading":    "status",
			"value":      1.0,
			"valueType":  "Bool",
			"baseType":   "Boolean",
			"timestamp":  baseTime.UnixNano(),
			"metadata":   `{}`,
		},
	}

	for i, item := range testData {
		deviceName := item["deviceName"].(string)
		if len(deviceName) != 7 {
			t.Fatalf("第 %d 条数据 deviceName 长度错误: 期望7, 实际%d", i+1, len(deviceName))
		}

		fields := table.GetAllFields()
		fields["id"] = item["id"]
		fields["deviceName"] = item["deviceName"]
		fields["reading"] = item["reading"]
		fields["value"] = item["value"]
		fields["valueType"] = item["valueType"]
		fields["baseType"] = item["baseType"]
		fields["timestamp"] = item["timestamp"]
		fields["metadata"] = item["metadata"]
		_, err := table.Insert(&fields)
		if err != nil {
			t.Fatalf("插入第 %d 条数据失败: %v", i+1, err)
		}
	}

	t.Run("PrimaryKeySearch", func(t *testing.T) {
		searchFields := map[string]any{
			"deviceName": "dev0001",
			"timestamp":  testData[0]["timestamp"],
		}
		dataIter, _ := table.Search(&searchFields)
		defer GlobalTableIterPool.Put(dataIter)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		records := dataIter.GetRecords(true)
		defer record.PutRecords(records)
		if len(records) != 1 {
			t.Errorf("期望找到1条记录，实际找到 %d 条", len(records))
		}
		if records[0]["value"] != testData[0]["value"] {
			t.Errorf("记录不匹配，期望: %v, 实际: %v", testData[0]["value"], records[0]["value"])
		}
	})

	t.Run("DeviceNameIndexSearch", func(t *testing.T) {
		searchFields := map[string]any{
			"deviceName": "dev0001",
		}
		dataIter, _ := table.Search(&searchFields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		records := dataIter.GetRecords(true)
		defer records.Release()
		if len(records) != 2 {
			t.Errorf("期望找到2条记录，实际找到 %d 条", len(records))
		}
		for _, rec := range records {
			if rec["deviceName"] != "dev0001" {
				t.Errorf("记录 deviceName 错误，期望: dev0001, 实际: %v", rec["deviceName"])
			}
		}
	})

	t.Run("TimestampIndexSearch", func(t *testing.T) {
		searchFields := map[string]any{
			"timestamp": testData[2]["timestamp"],
		}
		dataIter, _ := table.Search(&searchFields)
		defer GlobalTableIterPool.Put(dataIter)
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		records := dataIter.GetRecords(true)
		defer records.Release()
		if len(records) < 1 {
			t.Errorf("期望找到至少1条记录，实际找到 %d 条", len(records))
		}
		fmt.Printf("通过 timestamp 索引找到 %d 条记录\n", len(records))
	})

	t.Run("SearchAll", func(t *testing.T) {
		searchFields := map[string]any{
			"deviceName": nil,
		}
		dataIter, _ := table.Search(&searchFields)
		defer dataIter.Release()
		if dataIter.iter == nil {
			t.Fatalf("Search 失败")
		}
		records := dataIter.GetRecords(true)
		defer records.Release()
		if len(records) != len(testData) {
			t.Errorf("期望找到 %d 条记录，实际找到 %d 条", len(testData), len(records))
		}
		fmt.Printf("所有记录: %v\n", records)
	})

	fmt.Println("所有测试通过！")
}
