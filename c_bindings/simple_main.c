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

    // 关闭数据库
    SfsDbClose(manager);

    printf("All operations completed successfully\n");
    return 0;
}
