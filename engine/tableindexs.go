package engine

import "slices"

func (t *Table) CreateIndex(index Index) error {
	if t.indexIDManager == nil {
		t.indexIDManager = NewIDManager(t.kvStore)
	}
	fkey := t.indexIDManager.GenerateIndexKey(t.id, index.Name())
	idxid, err := t.indexIDManager.GetOrCreateID(fkey)
	if err != nil {
		return err
	}
	err = t.indexs.createIndex(index, idxid)
	if err != nil {
		// 回退ID
		_, err = t.indexIDManager.GetPreviousID(fkey)
		if err != nil {
			return err
		}
		return err
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

func (t *Table) GetPrimaryKey() PrimaryKey {
	pk := t.indexs.getPrimaryKey()
	if pk == nil {
		//没有主键，需要创建一个默认主键
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
