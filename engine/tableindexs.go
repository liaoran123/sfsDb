package engine

import (
	"slices"

	"github.com/liaoran123/sfsDb/util"
)

func (t *Table) CreateIndex(index Index) error {
	if t.indexIDManager == nil {
		t.indexIDManager = NewIDManager(t.kvStore)
	}
	fkey := t.indexIDManager.GenerateIndexKey(t.id, index.Name())
	idxID, isNew, err := t.indexIDManager.GetOrCreateID(fkey)
	if err != nil {
		return err
	}
	err = t.indexs.createIndex(index, idxID)
	if err != nil {
		// 回退ID
		_, err = t.indexIDManager.GetPreviousID(fkey)
		if err != nil {
			return err
		}
		return err
	}
	//对table现有数据创建对应的新索引数据
	if isNew {
		t.createIndexData(index)
	}
	return nil
}

// 创建普通复合索引
func (t *Table) CreateCompositeIndex(name string, fields ...string) error {
	idx, err := DefaultNormalIndexNew(name)
	if err != nil {
		return err
	}
	idx.AddFields(fields...)
	return t.CreateIndex(idx)
}

// 创建主键复合索引
func (t *Table) CreateCompositePrimaryKey(name string, fields ...string) error {
	idx, err := DefaultPrimaryKeyNew(name)
	if err != nil {
		return err
	}
	idx.AddFields(fields...)
	return t.CreateIndex(idx)
}

// 创建主键索引（支持单个或多个字段）
func (t *Table) CreatePrimaryKey(fields ...string) error {
	return t.CreateCompositePrimaryKey("primary", fields...)
}

// 创建普通索引（简化版，直接指定名称和字段）
func (t *Table) CreateSimpleIndex(name string, fields ...string) error {
	return t.CreateCompositeIndex(name, fields...)
}

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

// 获取所有索引
func (t *Table) GetAllIndexes() []Index {
	return t.indexs.GetAllIndexes()
}

// 根据名称获取索引
func (t *Table) GetIndexByName(name string) Index {
	return t.indexs.GetIndex(name)
}

// 根据字段名获取包含该字段的所有索引
func (t *Table) GetIndexesByField(field string) []Index {
	var result []Index
	for _, idx := range t.indexs.GetAllIndexes() {
		if slices.Contains(idx.GetFields(), field) {
			result = append(result, idx)
		}
	}
	return result
}

// 删除指定名称的索引
// 删除索引不会删除现存数据，不会对数据产生影响，只是存在冗余。
func (t *Table) DropIndex(name string) error {
	return t.indexs.DeleteIndex(name)
}

// 删除主键索引
func (t *Table) DropPrimaryKey() error {
	pk := t.GetPrimaryKey()
	if pk != nil {
		return t.indexs.DeleteIndex(pk.Name())
	}
	return nil
}

// 匹配索引
func (t *Table) MatchIndex(fields ...string) Index {
	return t.indexs.MatchIndex(fields...)
}

// 对table现有数据创建对应的新索引数据
func (t *Table) createIndexData(index Index) error {
	//排除主键
	if index == t.GetPrimaryKey() {
		return nil
	}
	//根据主键前缀遍历所有数据，创建索引数据
	//获取主键前缀
	pk := t.GetPrimaryKey()
	pkPrefix := pk.Prefix(t.id)
	slice := util.NewRangeHelper(pkPrefix).FromComparison(util.Like, pkPrefix)
	//pkPrefix创建迭代器
	iter := t.kvStore.Iterator(slice.Start, slice.Limit)
	defer iter.Release()
	//遍历所有数据
	//var fieldsBytes *map[string][]byte
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
				t.kvStore.Put(joinValue, indexValue)
			}
		default:
			return nil
		}

	}
	return nil
}
