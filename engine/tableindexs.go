package engine

import "slices"

func (t *Table) CreateIndex(index Index) error {
	if t.indexIDManager == nil {
		t.indexIDManager = NewIDManager(t.kvStore)
	}
	idxid, err := t.indexIDManager.GetOrCreateID(index.Name())
	if err != nil {
		return err
	}
	err = t.indexs.createIndex(index, idxid)
	if err != nil {
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

// 获取最佳匹配索引，返回匹配度最高的索引
func (t *Table) GetBestMatchIndex(fields ...string) Index {
	// 优先匹配主键
	pk := t.GetPrimaryKey()
	if pk.MatchFields(fields...) {
		return pk
	}

	// 然后匹配普通索引
	normalIndexes := t.indexs.GetNormalIndexs()
	for _, idx := range normalIndexes {
		if idx.MatchFields(fields...) {
			return idx
		}
	}

	// 最后匹配全文索引
	fullTextIndexes := t.indexs.GetFullTextIndexs()
	for _, idx := range fullTextIndexes {
		if idx.MatchFields(fields...) {
			return idx
		}
	}

	return nil
}

// 获取最佳匹配索引，并返回匹配结果
func (t *Table) GetMatchIndexResult(fields ...string) (Index, bool) {
	idx := t.GetBestMatchIndex(fields...)
	return idx, idx != nil
}

// 匹配索引
func (t *Table) MatchIndex(fields ...string) Index {
	return t.indexs.MatchIndex(fields...)
}
