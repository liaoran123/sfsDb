#include <stdio.h>
#include <stdlib.h>
#include "libsfsdb.h"

int main() {
    printf("Testing libsfsdb.a static library...\n");

    // 1. 初始化数据库
    SfsDbManager* manager = SfsDbInit("./test_db");
    if (manager == NULL) {
        printf("Failed to initialize database\n");
        return 1;
    }
    printf("Database initialized successfully\n");

    // 2. 创建表
    SfsTable* table = SfsDbCreateTable(manager, "test_table");
    if (table == NULL) {
        printf("Failed to create table\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Table created successfully\n");

    // 3. 设置表字段
    char* fieldNames[] = {"id", "name", "age"};
    char* fieldTypes[] = {"int", "string", "int"};
    int fieldCount = 3;
    if (SfsDbSetFields(table, (char**)fieldNames, (char**)fieldTypes, fieldCount) == 0) {
        printf("Failed to set fields\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Fields set successfully\n");

    // 4. 创建主键索引
    if (SfsDbCreatePrimaryKey(table, "id_index", "id") == 0) {
        printf("Failed to create primary key\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Primary key created successfully\n");

    // 5. 创建普通索引
    if (SfsDbCreateNormalIndex(table, "name_index", "name") == 0) {
        printf("Failed to create normal index\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Normal index created successfully\n");

    // 5. 插入数据
    char* insertFieldNames[] = {"id", "name", "age"};
    char* insertFieldValues1[] = {"1", "Alice", "25"};
    int64_t id1 = SfsDbInsert(table, (char**)insertFieldNames, (char**)insertFieldValues1, fieldCount);
    if (id1 == -1) {
        printf("Failed to insert data\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Inserted record with ID: %lld\n", id1);

    // 6. 搜索数据
    char* searchFieldNames[] = {"name"};
    char* searchFieldValues[] = {"Alice"};
    int searchFieldCount = 1;
    SfsIterator* iterator = SfsDbSearch(table, (char**)searchFieldNames, (char**)searchFieldValues, searchFieldCount);
    if (iterator == NULL) {
        printf("Failed to search data\n");
        SfsDbClose(manager);
        return 1;
    }
    printf("Search executed successfully\n");

    // 7. 获取记录
    SfsRecords* records = SfsDbGetRecords(iterator);
    if (records == NULL) {
        printf("Failed to get records\n");
        SfsDbReleaseIterator(iterator);
        SfsDbClose(manager);
        return 1;
    }
    printf("Retrieved %d records\n", SfsDbGetRecordsCount(records));

    // 8. 释放资源
    SfsDbReleaseRecords(records);
    SfsDbReleaseIterator(iterator);
    SfsDbClose(manager);

    printf("Test completed successfully\n");
    return 0;
}
