package main

/*
#include <stdint.h>
#include <stdlib.h>

// 定义 C 接口结构体
typedef struct {
    void* dbManager;
} SfsDbManager;

typedef struct {
    void* table;
} SfsTable;

typedef struct {
    void* iterator;
} SfsIterator;

typedef struct {
    void* records;
    int count;
} SfsRecords;
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

// 导出函数：初始化数据库管理器
//
//export SfsDbInit
func SfsDbInit(dbPath *C.char) *C.SfsDbManager {
	dbPathStr := C.GoString(dbPath)
	dbManager := storage.GetDBManager()
	_, err := dbManager.OpenDB(dbPathStr)
	if err != nil {
		fmt.Printf("OpenDB failed: %v\n", err)
		return nil
	}

	// 分配 C 内存来存储 SfsDbManager
	manager := (*C.SfsDbManager)(C.malloc(C.sizeof_SfsDbManager))
	if manager == nil {
		fmt.Println("malloc failed")
		return nil
	}

	// 存储 dbManager 指针
	manager.dbManager = unsafe.Pointer(dbManager)

	return manager
}

// 导出函数：关闭数据库管理器
//
//export SfsDbClose
func SfsDbClose(manager *C.SfsDbManager) {
	if manager != nil && manager.dbManager != nil {
		dbManager := (*storage.DBManager)(manager.dbManager)
		dbManager.CloseDB()
		// 释放 C 内存
		C.free(unsafe.Pointer(manager))
	}
}

// 导出函数：创建或打开表
//
//export SfsDbCreateTable
func SfsDbCreateTable(manager *C.SfsDbManager, tableName *C.char) *C.SfsTable {
	if manager == nil || manager.dbManager == nil {
		return nil
	}

	tableNameStr := C.GoString(tableName)
	table, err := engine.TableNew(tableNameStr)
	if err != nil {
		fmt.Printf("TableNew failed: %v\n", err)
		return nil
	}

	// 分配 C 内存来存储 SfsTable
	tbl := (*C.SfsTable)(C.malloc(C.sizeof_SfsTable))
	if tbl == nil {
		fmt.Println("malloc failed")
		return nil
	}

	// 存储 table 指针
	tbl.table = unsafe.Pointer(table)

	return tbl
}

// 导出函数：设置表字段
//
//export SfsDbSetFields
func SfsDbSetFields(table *C.SfsTable, fieldNames **C.char, fieldTypes **C.char, fieldCount C.int) C.int {
	if table == nil || table.table == nil {
		return 0
	}

	tbl := (*engine.Table)(table.table)
	fields := make(map[string]any)

	for i := 0; i < int(fieldCount); i++ {
		// 获取字段名
		fieldNamePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldNames)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldName := C.GoString(*fieldNamePtr)

		// 获取字段类型
		fieldTypePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldTypes)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldType := C.GoString(*fieldTypePtr)

		// 根据字段类型设置默认值
		switch fieldType {
		case "int":
			fields[fieldName] = 0
		case "string":
			fields[fieldName] = ""
		case "float":
			fields[fieldName] = 0.0
		case "bool":
			fields[fieldName] = false
		default:
			fields[fieldName] = ""
		}
	}

	err := tbl.SetFields(fields)
	if err != nil {
		fmt.Printf("SetFields failed: %v\n", err)
		return 0
	}

	return 1
}

// 导出函数：创建主键索引
//
//export SfsDbCreatePrimaryKey
func SfsDbCreatePrimaryKey(table *C.SfsTable, indexName *C.char, fieldName *C.char) C.int {
	if table == nil || table.table == nil {
		return 0
	}

	tbl := (*engine.Table)(table.table)
	indexNameStr := C.GoString(indexName)
	fieldNameStr := C.GoString(fieldName)

	primaryKey, err := engine.NewDefaultPrimaryKey(indexNameStr)
	if err != nil {
		fmt.Printf("NewDefaultPrimaryKey failed: %v\n", err)
		return 0
	}

	primaryKey.AddFields(fieldNameStr)
	err = tbl.CreateIndex(primaryKey)
	if err != nil {
		fmt.Printf("CreateIndex failed: %v\n", err)
		return 0
	}

	return 1
}

// 导出函数：创建普通索引
//
//export SfsDbCreateNormalIndex
func SfsDbCreateNormalIndex(table *C.SfsTable, indexName *C.char, fieldName *C.char) C.int {
	if table == nil || table.table == nil {
		return 0
	}

	tbl := (*engine.Table)(table.table)
	indexNameStr := C.GoString(indexName)
	fieldNameStr := C.GoString(fieldName)

	normalIndex, err := engine.NewDefaultNormalIndex(indexNameStr)
	if err != nil {
		fmt.Printf("NewDefaultNormalIndex failed: %v\n", err)
		return 0
	}

	normalIndex.AddFields(fieldNameStr)
	err = tbl.CreateIndex(normalIndex)
	if err != nil {
		fmt.Printf("CreateIndex failed: %v\n", err)
		return 0
	}

	return 1
}

// 导出函数：插入数据
//
//export SfsDbInsert
func SfsDbInsert(table *C.SfsTable, fieldNames **C.char, fieldValues **C.char, fieldCount C.int) C.int64_t {
	if table == nil || table.table == nil {
		return -1
	}

	tbl := (*engine.Table)(table.table)
	data := make(map[string]any)

	for i := 0; i < int(fieldCount); i++ {
		// 获取字段名
		fieldNamePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldNames)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldName := C.GoString(*fieldNamePtr)

		// 获取字段值
		fieldValuePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldValues)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldValue := C.GoString(*fieldValuePtr)

		data[fieldName] = fieldValue
	}

	currentID, err := tbl.Insert(&data)
	if err != nil {
		fmt.Printf("Insert failed: %v\n", err)
		return -1
	}

	return C.int64_t(currentID)
}

// 导出函数：搜索数据
//
//export SfsDbSearch
func SfsDbSearch(table *C.SfsTable, fieldNames **C.char, fieldValues **C.char, fieldCount C.int) *C.SfsIterator {
	if table == nil || table.table == nil {
		return nil
	}

	tbl := (*engine.Table)(table.table)
	searchData := make(map[string]any)

	for i := 0; i < int(fieldCount); i++ {
		// 获取字段名
		fieldNamePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldNames)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldName := C.GoString(*fieldNamePtr)

		// 获取字段值
		fieldValuePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldValues)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldValue := C.GoString(*fieldValuePtr)

		searchData[fieldName] = fieldValue
	}

	iter, err := tbl.Search(&searchData)
	if err != nil {
		fmt.Printf("Search failed: %v\n", err)
		return nil
	}

	// 分配 C 内存来存储 SfsIterator
	iterator := (*C.SfsIterator)(C.malloc(C.sizeof_SfsIterator))
	if iterator == nil {
		fmt.Println("malloc failed")
		return nil
	}

	// 存储 iterator 指针
	iterator.iterator = unsafe.Pointer(iter)

	return iterator
}

// 导出函数：获取记录
//
//export SfsDbGetRecords
func SfsDbGetRecords(iterator *C.SfsIterator) *C.SfsRecords {
	if iterator == nil || iterator.iterator == nil {
		return nil
	}

	iter := (*engine.TableIter)(iterator.iterator)
	records := iter.GetRecords(true)
	if records == nil || len(records) == 0 {
		return nil
	}

	// 分配 C 内存来存储 SfsRecords
	recs := (*C.SfsRecords)(C.malloc(C.sizeof_SfsRecords))
	if recs == nil {
		fmt.Println("malloc failed")
		return nil
	}

	recs.records = unsafe.Pointer(&records[0])
	recs.count = C.int(len(records))

	return recs
}

// 导出函数：获取记录数量
//
//export SfsDbGetRecordsCount
func SfsDbGetRecordsCount(records *C.SfsRecords) C.int {
	if records == nil {
		return 0
	}
	return records.count
}

// 导出函数：释放迭代器
//
//export SfsDbReleaseIterator
func SfsDbReleaseIterator(iterator *C.SfsIterator) {
	if iterator != nil && iterator.iterator != nil {
		iter := (*engine.TableIter)(iterator.iterator)
		iter.Release()
		// 释放 C 内存
		C.free(unsafe.Pointer(iterator))
	}
}

// 导出函数：释放记录
//
//export SfsDbReleaseRecords
func SfsDbReleaseRecords(records *C.SfsRecords) {
	if records != nil {
		// 释放 C 内存
		C.free(unsafe.Pointer(records))
	}
}

// 导出函数：批量插入数据（带自动增值）
//
//export SfsDbBatchInsertInc
func SfsDbBatchInsertInc(table *C.SfsTable, records **C.char, recordCount C.int) *C.int64_t {
	if table == nil || table.table == nil {
		return nil
	}

	tbl := (*engine.Table)(table.table)
	var recordList []*map[string]any

	for i := 0; i < int(recordCount); i++ {
		// 获取记录数据
		recordPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(records)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		_ = C.GoString(*recordPtr)

		// 解析记录数据（这里需要根据实际格式解析，示例为简化版）
		data := make(map[string]any)
		// 这里需要根据实际的记录格式解析，示例为简化版
		recordList = append(recordList, &data)
	}

	ids, err := tbl.BatchInsertInc(recordList)
	if err != nil {
		fmt.Printf("BatchInsertInc failed: %v\n", err)
		return nil
	}

	// 分配 C 内存存储 ID 列表
	idArray := (*C.int64_t)(C.malloc(C.size_t(len(ids)) * C.sizeof_int64_t))
	if idArray == nil {
		fmt.Println("malloc failed")
		return nil
	}

	// 复制 ID 到 C 内存
	for i, id := range ids {
		(*[1 << 30]C.int64_t)(unsafe.Pointer(idArray))[i] = C.int64_t(id)
	}

	return idArray
}

// 导出函数：批量插入数据（无自动增值）
//
//export SfsDbBatchInsertNoInc
func SfsDbBatchInsertNoInc(table *C.SfsTable, records **C.char, recordCount C.int) *C.int64_t {
	if table == nil || table.table == nil {
		return nil
	}

	tbl := (*engine.Table)(table.table)
	var recordList []*map[string]any

	for i := 0; i < int(recordCount); i++ {
		// 获取记录数据
		recordPtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(records)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		_ = C.GoString(*recordPtr)

		// 解析记录数据（这里需要根据实际格式解析，示例为简化版）
		data := make(map[string]any)
		// 这里需要根据实际的记录格式解析，示例为简化版
		recordList = append(recordList, &data)
	}

	ids, err := tbl.BatchInsertNoInc(recordList)
	if err != nil {
		fmt.Printf("BatchInsertNoInc failed: %v\n", err)
		return nil
	}

	// 分配 C 内存存储 ID 列表
	idArray := (*C.int64_t)(C.malloc(C.size_t(len(ids)) * C.sizeof_int64_t))
	if idArray == nil {
		fmt.Println("malloc failed")
		return nil
	}

	// 复制 ID 到 C 内存
	for i, id := range ids {
		(*[1 << 30]C.int64_t)(unsafe.Pointer(idArray))[i] = C.int64_t(id)
	}

	return idArray
}

// 导出函数：更新数据
//
//export SfsDbUpdate
func SfsDbUpdate(table *C.SfsTable, fieldNames **C.char, fieldValues **C.char, fieldCount C.int) C.int {
	if table == nil || table.table == nil {
		return 0
	}

	tbl := (*engine.Table)(table.table)
	data := make(map[string]any)

	for i := 0; i < int(fieldCount); i++ {
		// 获取字段名
		fieldNamePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldNames)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldName := C.GoString(*fieldNamePtr)

		// 获取字段值
		fieldValuePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldValues)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldValue := C.GoString(*fieldValuePtr)

		data[fieldName] = fieldValue
	}

	err := tbl.Update(&data)
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
		return 0
	}

	return 1
}

// 导出函数：删除数据
//
//export SfsDbDelete
func SfsDbDelete(table *C.SfsTable, fieldNames **C.char, fieldValues **C.char, fieldCount C.int) C.int {
	if table == nil || table.table == nil {
		return 0
	}

	tbl := (*engine.Table)(table.table)
	data := make(map[string]any)

	for i := 0; i < int(fieldCount); i++ {
		// 获取字段名
		fieldNamePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldNames)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldName := C.GoString(*fieldNamePtr)

		// 获取字段值
		fieldValuePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldValues)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldValue := C.GoString(*fieldValuePtr)

		data[fieldName] = fieldValue
	}

	err := tbl.Delete(&data)
	if err != nil {
		fmt.Printf("Delete failed: %v\n", err)
		return 0
	}

	return 1
}

// 导出函数：范围搜索
//
//export SfsDbSearchRange
func SfsDbSearchRange(table *C.SfsTable, startFieldNames **C.char, startFieldValues **C.char, limitFieldNames **C.char, limitFieldValues **C.char, fieldCount C.int) *C.SfsIterator {
	if table == nil || table.table == nil {
		return nil
	}

	tbl := (*engine.Table)(table.table)
	startData := make(map[string]any)
	limitData := make(map[string]any)

	// 解析起始条件
	for i := 0; i < int(fieldCount); i++ {
		// 获取字段名
		fieldNamePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(startFieldNames)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldName := C.GoString(*fieldNamePtr)

		// 获取字段值
		fieldValuePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(startFieldValues)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldValue := C.GoString(*fieldValuePtr)

		startData[fieldName] = fieldValue
	}

	// 解析结束条件
	for i := 0; i < int(fieldCount); i++ {
		// 获取字段名
		fieldNamePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(limitFieldNames)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldName := C.GoString(*fieldNamePtr)

		// 获取字段值
		fieldValuePtr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(limitFieldValues)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		fieldValue := C.GoString(*fieldValuePtr)

		limitData[fieldName] = fieldValue
	}

	iter, err := tbl.SearchRange(nil, &startData, &limitData)
	if err != nil {
		fmt.Printf("SearchRange failed: %v\n", err)
		return nil
	}

	// 分配 C 内存来存储 SfsIterator
	iterator := (*C.SfsIterator)(C.malloc(C.sizeof_SfsIterator))
	if iterator == nil {
		fmt.Println("malloc failed")
		return nil
	}

	// 存储 iterator 指针
	iterator.iterator = unsafe.Pointer(iter)

	return iterator
}

// 主函数，用于编译成静态库
func main() {}
