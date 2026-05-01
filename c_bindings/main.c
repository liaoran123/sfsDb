#include <stdio.h>
#include <stdlib.h>
#include "libsfsdb.h"

int main() {
    // 初始化数据库
    const char* dbPath = "./c_example_db";
    SfsDbManager* manager = SfsDbInit((char*)dbPath);
    if (manager == NULL) {
        printf("Failed to initialize database\n");
        return 1;
    }
    printf("Database initialized successfully\n");

    // 创建表
    const char* tableName = "users";
    SfsTable* table = SfsDbCreateTable(manager, (char*)tableName);
    if (table == NULL) {
        printf("Failed to create table\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Table created successfully\n");

    // 设置表字段
    const char* fieldNames[] = {"id", "name", "age", "email"};
    const char* fieldTypes[] = {"int", "string", "int", "string"};
    int fieldCount = 4;

    if (SfsDbSetFields(table, (char**)fieldNames, (char**)fieldTypes, fieldCount) == 0) {
        printf("Failed to set fields\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Fields set successfully\n");

    // 创建主键索引
    if (SfsDbCreatePrimaryKey(table, "id_index", "id") == 0) {
        printf("Failed to create primary key\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Primary key created successfully\n");

    // 创建普通索引
    if (SfsDbCreateNormalIndex(table, "name_index", "name") == 0) {
        printf("Failed to create normal index\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Normal index created successfully\n");

    // 插入数据
    const char* insertFieldNames[] = {"id", "name", "age", "email"};
    const char* insertFieldValues1[] = {"1", "John Doe", "30", "john@example.com"};
    const char* insertFieldValues2[] = {"2", "Jane Smith", "25", "jane@example.com"};

    int64_t id1 = SfsDbInsert(table, (char**)insertFieldNames, (char**)insertFieldValues1, fieldCount);
    if (id1 == -1) {
        printf("Failed to insert data\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Inserted data with ID: %lld\n", id1);

    int64_t id2 = SfsDbInsert(table, (char**)insertFieldNames, (char**)insertFieldValues2, fieldCount);
    if (id2 == -1) {
        printf("Failed to insert data\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Inserted data with ID: %lld\n", id2);

    // 搜索数据
    const char* searchFieldNames[] = {"name"};
    const char* searchFieldValues[] = {"John Doe"};
    int searchFieldCount = 1;

    SfsIterator* iterator = SfsDbSearch(table, (char**)searchFieldNames, (char**)searchFieldValues, searchFieldCount);
    if (iterator == NULL) {
        printf("Failed to search data\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Search executed successfully\n");

    // 获取记录
    SfsRecords* records = SfsDbGetRecords(iterator);
    if (records == NULL) {
        printf("Failed to get records\n");
        SfsDbReleaseIterator(iterator);
        SfsDbClose(manager);
        return 1;
    }

    int recordCount = SfsDbGetRecordsCount(records);
    printf("Found %d records\n", recordCount);

    // 释放资源
    SfsDbReleaseRecords(records);
    SfsDbReleaseIterator(iterator);
    SfsDbClose(manager);

    printf("All operations completed successfully\n");
    return 0;
}
