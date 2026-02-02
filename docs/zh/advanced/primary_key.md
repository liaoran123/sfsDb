# 主键管理

## 3.1 单主键

```go
// 创建单主键表
userTable, err := engine.TableNew("user_table")
if err != nil {
    panic(err)
}

// 设置字段
userFields := map[string]any{
    "id":   0,     // 主键字段
    "name": "",    // 普通字段
    "age":  0,     // 普通字段
}
err = userTable.SetFields(userFields)
if err != nil {
    panic(err)
}

// 创建单主键索引
pk, err := engine.DefaultPrimaryKeyNew("pk_id")
err = pk.AddFields("id")
err = userTable.CreateIndex(pk)
if err != nil {
    panic(err)
}
```

## 3.2 复合主键

```go
// 创建复合主键表
orderTable, err := engine.TableNew("order_table")
if err != nil {
    panic(err)
}

// 设置复合主键字段
orderFields := map[string]any{
    "order_id": 0,     // 复合主键字段1
    "user_id":  0,     // 复合主键字段2
    "product":  "",    // 普通字段
    "quantity": 0,     // 普通字段
}
err = orderTable.SetFields(orderFields)
if err != nil {
    panic(err)
}

// 创建复合主键索引
compositePk, err := engine.DefaultPrimaryKeyNew("pk_order_user")
if err != nil {
    panic(err)
}
// 添加多个主键字段
compositePk.AddFields("order_id", "user_id")
err = orderTable.CreateIndex(compositePk)
if err != nil {
    panic(err)
}

// 插入复合主键记录
order := map[string]any{
    "order_id": 1,
    "user_id":  100,
    "product":  "商品A",
    "quantity": 2,
}
_, err = orderTable.Insert(&order)
if err != nil {
    panic(err)
}
```

## 3.3 主键自动创建机制

当表中没有显式创建主键时，sfsDb 会在首次需要主键时自动创建一个默认主键。这个机制由 `GetPrimaryKey()` 方法实现：

```go
// 获取主键索引
func (t *Table) GetPrimaryKey() PrimaryKey {
    pk := t.indexs.getPrimaryKey()
    if pk == nil {
        //没有主键，需要创建一个默认主键。
        //所以主键必须在表未有数据前创建。
        pk, _ = DefaultPrimaryKeyNew("id")
        pk.AddFields("id")
        t.CreateIndex(pk)
    }
    return pk
}
```

**工作原理**：
1. 当调用 `GetPrimaryKey()` 方法时，系统会首先检查表是否已有主键
2. 如果没有主键，系统会自动创建一个名为 "id" 的默认主键
3. 默认主键会添加 "id" 字段作为主键字段
4. 然后将这个默认主键索引添加到表中
5. 最后返回创建的主键索引

**注意事项**：
- 系统注释明确指出："所以主键必须在表未有数据前创建。"
- 这意味着如果表中已经有数据，再调用需要主键的操作时，可能会导致数据不一致
- 因此，建议在表创建后、插入数据前，显式设置主键

**使用场景**：
- 快速原型开发，暂时不需要自定义主键
- 简单应用，使用默认的自增 ID 作为主键即可满足需求

**索引创建顺序的重要性**：
在添加索引时，系统会执行 `createIndexData` 函数来为现有数据创建索引数据，该函数会调用 `GetPrimaryKey()` 方法：

```go
// 对table现有数据创建对应的新索引数据
func (t *Table) createIndexData(index Index) error {
    //排除主键
    if index == t.GetPrimaryKey() {
        return nil
    }
    //根据主键前缀遍历所有数据，创建索引数据
    //获取主键前缀
    pk := t.GetPrimaryKey() // 这里会触发自动创建主键
    pkPrefix := pk.Prefix(t.id)
    slice := util.NewRangeHelper(pkPrefix).FromComparison(util.Like, pkPrefix)
    //pkPrefix创建迭代器
    iter := t.kvStore.Iterator(slice.Start, slice.Limit)
    defer GlobalTableIterPool.Put( iter)
    //遍历所有数据
    var value []byte
    for iter.Next() {
        _, value = iter.Key(), iter.Value()
        rval, err := pk.Parse(t.fieldsid, value)
        if err != nil {
            return err
        }
        if rval == nil {
            continue
        }
        indexValue := pk.GetID(rval)
        switch index := index.(type) {
        case NormalIndex:
            idxvalueofkey := index.Join(rval)
            idxkey := index.JoinPrefix(t.id, idxvalueofkey)
            t.kvStore.Put(idxkey, indexValue)
        case FullTextIndex:
            //func (c *batchContainer) Operation 函数基本一样
            fieldsBytes, err := pk.Parse(t.fieldsid, value)
            if err != nil {
                return err
            }
            if fieldsBytes == nil {
                continue
            }
            joinValues := index.JoinFullValues(fieldsBytes, t.id)
            defer util.PutBytesArray(joinValues)
            for _, joinValue := range joinValues {
                if joinValue == nil {
                    continue
                }
                // 为全文索引创建索引数据
                t.kvStore.Put(joinValue, indexValue)
            }
        }
    }
    return nil
}
```

**重要建议**：
1. **先添加主键，后添加其他索引**：如果在添加其他索引之前没有显式创建主键，系统会在执行 `createIndexData` 时自动创建一个默认主键
2. **避免自动创建主键**：自动创建的主键可能不符合你的业务需求，建议始终显式创建主键
3. **在数据插入前设置主键**：确保在插入任何数据之前设置好主键，以避免数据不一致