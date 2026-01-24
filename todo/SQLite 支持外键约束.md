SQLite 支持外键约束 ，但有一些重要的前提和特性需要注意：

### 1. 外键支持的基本情况
- 支持版本 ：从 SQLite 3.6.19 版本开始引入外键约束支持
- 默认状态 ： 默认禁用 ，需要手动启用
- 核心特性 ：支持参照完整性检查、级联操作（CASCADE）等
### 2. 如何启用外键约束
需要通过 PRAGMA 语句显式启用：

```
-- 启用外键约束（当前连接有效）
PRAGMA foreign_keys = ON;

-- 检查外键是否启用（返回1表示启用，0表示禁用）
PRAGMA foreign_keys;
```
### 3. 外键语法示例
创建表时定义外键约束：

```
-- 主表（父表）
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

-- 从表（子表），带外键约束
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    product TEXT NOT NULL,
    -- 定义外键约束，引用users表的id列
    FOREIGN KEY (user_id) REFERENCES users(id) 
        ON DELETE CASCADE  -- 级联删除：当users表中的记录
        被删除时，关联的orders记录也被删除
        ON UPDATE CASCADE  -- 级联更新：当users表的id更新
        时，关联的orders记录的user_id也更新
);
```
### 4. 支持的外键操作
- CASCADE ：级联操作，父表操作时自动更新/删除子表关联记录
- SET NULL ：父表操作时，子表关联字段设为NULL
- SET DEFAULT ：父表操作时，子表关联字段设为默认值
- RESTRICT ：禁止父表操作（如果存在关联的子表记录）
- NO ACTION ：默认行为，与RESTRICT类似（但检查时机不同）
### 5. 外键注意事项
- 索引要求 ：SQLite不会自动为外键创建索引， 建议手动创建 （否则可能导致全表扫描，影响性能）
- 数据类型 ：外键字段的数据类型必须与被引用字段完全匹配
- 主键要求 ：被引用的字段必须是父表的主键或具有唯一索引
- 临时表限制 ：临时表不支持外键约束
- 视图限制 ：视图不支持外键约束
- 触发器冲突 ：如果定义了与外键操作冲突的触发器，触发器优先级更高
### 6. 查看外键信息
```
-- 查看表的外键约束列表
PRAGMA foreign_key_list(orders);

-- 查看数据库中所有外键相关的PRAGMA设置
PRAGMA foreign_keys;
```
### 总结
SQLite 完整支持外键约束 ，但需要手动启用，并且有一些语法和性能方面的注意事项。合理使用外键可以确保数据的参照完整性，适合需要严格数据关系的应用场景。

对比sfsDb ：根据之前的功能分析，sfsDb目前 不支持外键约束 ，这是sfsDb与SQLite在功能上的一个差异点。