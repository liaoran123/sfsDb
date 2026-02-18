// sfsDb C语言集成示例
// 展示如何在C/C++项目中使用sfsDb静态库

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// 声明sfsDb的C接口函数
// 注意：这些函数需要在Go代码中使用//export指令导出

#ifdef __cplusplus
extern "C" {
#endif

// 初始化数据库
void InitDB(const char* dbPath);

// 关闭数据库
void CloseDB();

// 创建表
int CreateTable(const char* tableName);

// 设置表字段
int SetTableFields(const char* tableName, const char* fieldsJson);

// 创建索引
int CreateIndex(const char* tableName, const char* indexJson);

// 插入数据
int InsertData(const char* tableName, const char* dataJson);

// 更新数据
int UpdateData(const char* tableName, const char* dataJson);

// 删除数据
int DeleteData(const char* tableName, const char* dataJson);

// 搜索数据
char* SearchData(const char* tableName, const char* queryJson);

// 释放搜索结果
void FreeSearchResult(char* result);

#ifdef __cplusplus
}
#endif

int main() {
    printf("sfsDb C语言集成示例\n");
    printf("========================================\n");

    // 1. 初始化数据库
    printf("\n1. 初始化数据库\n");
    InitDB("./c_example_db");
    printf("数据库初始化成功\n");

    // 2. 创建表
    printf("\n2. 创建表\n");
    int tableId = CreateTable("users");
    if (tableId >= 0) {
        printf("表创建成功，ID: %d\n", tableId);
    } else {
        printf("表创建失败\n");
        goto cleanup;
    }

    // 3. 设置表字段
    printf("\n3. 设置表字段\n");
    const char* fieldsJson = "{\"id\": 0, \"name\": \"\", \"age\": 0, \"email\": \"\"}";
    int setFieldsResult = SetTableFields("users", fieldsJson);
    if (setFieldsResult == 0) {
        printf("字段设置成功\n");
    } else {
        printf("字段设置失败\n");
        goto cleanup;
    }

    // 4. 创建主键索引
    printf("\n4. 创建主键索引\n");
    const char* primaryKeyJson = "{\"name\": \"id\", \"fields\": [\"id\"], \"type\": \"primary\"}";
    int createIndexResult = CreateIndex("users", primaryKeyJson);
    if (createIndexResult == 0) {
        printf("主键索引创建成功\n");
    } else {
        printf("主键索引创建失败\n");
        goto cleanup;
    }

    // 5. 插入数据
    printf("\n5. 插入数据\n");
    const char* dataJson1 = "{\"id\": 1, \"name\": \"张三\", \"age\": 25, \"email\": \"zhangsan@example.com\"}";
    int insertResult1 = InsertData("users", dataJson1);
    if (insertResult1 > 0) {
        printf("数据插入成功，ID: %d\n", insertResult1);
    } else {
        printf("数据插入失败\n");
        goto cleanup;
    }

    const char* dataJson2 = "{\"id\": 2, \"name\": \"李四\", \"age\": 30, \"email\": \"lisi@example.com\"}";
    int insertResult2 = InsertData("users", dataJson2);
    if (insertResult2 > 0) {
        printf("数据插入成功，ID: %d\n", insertResult2);
    } else {
        printf("数据插入失败\n");
        goto cleanup;
    }

    // 6. 搜索数据
    printf("\n6. 搜索数据\n");
    const char* queryJson = "{\"id\": 1}";
    char* searchResult = SearchData("users", queryJson);
    if (searchResult != NULL) {
        printf("搜索结果: %s\n", searchResult);
        FreeSearchResult(searchResult);
    } else {
        printf("搜索失败\n");
        goto cleanup;
    }

    // 7. 更新数据
    printf("\n7. 更新数据\n");
    const char* updateJson = "{\"id\": 1, \"name\": \"张三(已更新)\", \"age\": 26}";
    int updateResult = UpdateData("users", updateJson);
    if (updateResult == 0) {
        printf("数据更新成功\n");
    } else {
        printf("数据更新失败\n");
        goto cleanup;
    }

    // 8. 再次搜索数据，验证更新
    printf("\n8. 验证更新结果\n");
    searchResult = SearchData("users", queryJson);
    if (searchResult != NULL) {
        printf("更新后结果: %s\n", searchResult);
        FreeSearchResult(searchResult);
    } else {
        printf("搜索失败\n");
        goto cleanup;
    }

    // 9. 删除数据
    printf("\n9. 删除数据\n");
    const char* deleteJson = "{\"id\": 2}";
    int deleteResult = DeleteData("users", deleteJson);
    if (deleteResult == 0) {
        printf("数据删除成功\n");
    } else {
        printf("数据删除失败\n");
        goto cleanup;
    }

    printf("\n========================================\n");
    printf("示例运行完成\n");
    printf("========================================\n");

cleanup:
    // 10. 关闭数据库
    printf("\n10. 关闭数据库\n");
    CloseDB();
    printf("数据库关闭成功\n");

    return 0;
}
