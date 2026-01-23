package engine

import (
	"errors"
	"slices"
)

// 索引管理结构
type Indexs struct {
	id     uint8
	indexs []Index
	fields *map[string]any //表字段
}

/*
创建索引，检查字段必须存在表字段，索引名不能重复
删除索引
匹配索引，优先匹配唯一索引PrimaryKey，再匹配普通索引NormalIndex，FullTextIndex
*/

// 创建索引管理实例
func NewIndexs(fields *map[string]any) *Indexs {
	return &Indexs{
		indexs: make([]Index, 0),
		fields: fields,
	}
}

// 创建索引
// 索引规则，1，索引名称唯一；2，索引字段不能完全相同；3，只能有一个PrimaryKey索引；4，字段必须存在表字段
func (i *Indexs) createIndex(index Index, idxid uint8) error {
	fields := index.GetFields()

	// 检查索引字段是否存在于表字段中
	for _, field := range fields {
		if _, ok := (*i.fields)[field]; !ok {
			return errors.New("field \"" + field + "\" does not exist in table")
		}
	}

	// 检查索引名、主键类型和字段重复 - 合并为单次循环
	isNewPrimary := false
	if _, ok := index.(PrimaryKey); ok {
		isNewPrimary = true
	}

	indexName := index.Name()
	//性能优化，一次遍历，执行3个功能 ：1，索引名称唯一；2，索引字段不能完全相同；3，只能有一个PrimaryKey索引；
	for _, idx := range i.indexs {
		// 检查索引名是否重复。1，索引名称唯一；
		if idx.Name() == indexName {
			return errors.New("index name \"" + indexName + "\" already exists")
		}
		// 检查主键是否重复。3，只能有一个PrimaryKey索引；
		if isNewPrimary {
			if _, ok := idx.(PrimaryKey); ok {
				return errors.New("primary key index already exists")
			}
		}
		// 检查字段是否完全相同。2，索引字段不能完全相同；
		if slices.Equal(fields, idx.GetFields()) {
			return errors.New("index with identical fields already exists")
		}
	}

	// 添加索引
	index.setId(idxid)
	i.indexs = append(i.indexs, index)
	return nil
}

// 返回PrimaryKey索引
func (i *Indexs) getPrimaryKey() PrimaryKey {
	for _, index := range i.indexs {
		if _, ok := index.(PrimaryKey); ok {
			return index.(PrimaryKey)
		}
	}
	return nil
}

// 返回普通索引
func (i *Indexs) GetNormalIndexs() []NormalIndex {
	normalIndexs := make([]NormalIndex, 0)
	//普通索引是基类，全部匹配，所以需要判断不是PrimaryKey和FullTextIndex才符合普通索引
	for _, index := range i.indexs {
		if _, ok := index.(NormalIndex); ok {
			normalIndexs = append(normalIndexs, index.(NormalIndex))
		}
	}
	return normalIndexs
}

// 返回全文索引
func (i *Indexs) GetFullTextIndexs() []FullTextIndex {
	fullTextIndexs := make([]FullTextIndex, 0)
	for _, index := range i.indexs {
		if _, ok := index.(FullTextIndex); ok {
			fullTextIndexs = append(fullTextIndexs, index.(FullTextIndex))
		}
	}
	return fullTextIndexs
}

// 删除索引
func (i *Indexs) DeleteIndex(name string) error {
	for idx, index := range i.indexs {
		if index.Name() == name {
			// 移除索引
			i.indexs = append(i.indexs[:idx], i.indexs[idx+1:]...)
			return nil
		}
	}
	return errors.New("index \"" + name + "\" does not exist")
}

// 根据名称获取索引
func (i *Indexs) GetIndex(name string) Index {
	for _, index := range i.indexs {
		if index.Name() == name {
			return index
		}
	}
	return nil
}

// 获取所有索引
func (i *Indexs) GetAllIndexes() []Index {
	return i.indexs
}

// 获取索引数量
func (i *Indexs) Len() int {
	return len(i.indexs)
}

// 匹配索引，优先匹配唯一索引PrimaryKey，再匹配普通索引NormalIndex，FullTextIndex
func (i *Indexs) MatchIndex(fields ...string) Index {
	// 1. 优先匹配唯一索引PrimaryKey
	for _, index := range i.indexs {
		if _, ok := index.(PrimaryKey); ok {
			if index.MatchFields(fields...) {
				return index
			}
		}
	}

	// 2. 再匹配普通索引NormalIndex
	for _, index := range i.indexs {
		if _, ok := index.(NormalIndex); ok {
			if index.MatchFields(fields...) {
				return index
			}
		}
	}

	// 3. 最后匹配全文索引FullTextIndex
	for _, index := range i.indexs {
		if _, ok := index.(FullTextIndex); ok {
			if index.MatchFields(fields...) {
				return index
			}
		}
	}

	return nil
}

// 修改所有包含字段名称的索引
func (i *Indexs) UpdateFields(oldfields string, newfields string) {
	for _, index := range i.indexs {
		index.UpdateFields(oldfields, newfields)
	}
}

// 删除所有包含字段的索引
func (i *Indexs) DeleteFields(field ...string) {
	for _, index := range i.indexs {
		index.DeleteFields(field...)
	}
}

// 获取所有索引的名称和id映射
func (i *Indexs) GetAllIndexNameIdMap() map[string]uint8 {
	indexNameIdMap := make(map[string]uint8)
	for _, index := range i.indexs {
		indexNameIdMap[index.Name()] = index.GetId()
	}
	return indexNameIdMap
}
