package engine

import (
	"errors"
	"slices"
)

// 索引管理结构
type Indexs struct {
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

// 创建索引，检查字段必须存在表字段，索引名不能重复
func (i *Indexs) CreateIndex(index Index) error {
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
	for _, idx := range i.indexs {
		// 检查索引名是否重复
		if idx.Name() == indexName {
			return errors.New("index name \"" + indexName + "\" already exists")
		}

		// 检查主键是否重复
		if isNewPrimary {
			if _, ok := idx.(PrimaryKey); ok {
				return errors.New("primary key index already exists")
			}
		}

		// 检查字段是否完全相同
		if slices.Equal(fields, idx.GetFields()) {
			return errors.New("index with identical fields already exists")
		}
	}

	// 添加索引
	i.indexs = append(i.indexs, index)
	return nil
}

// 返回PrimaryKey索引
func (i *Indexs) GetPrimaryKey() PrimaryKey {
	for _, index := range i.indexs {
		if _, ok := index.(PrimaryKey); ok {
			return index.(PrimaryKey)
		}
	}
	// 如果没有找到PrimaryKey索引，系统默认创建一个字段为id的PrimaryKey索引
	primaryKey, err := DefaultPrimaryKeyNew("pk")
	if err != nil {
		return nil
	}
	primaryKey.AddFields("id")
	if err := i.CreateIndex(primaryKey); err != nil {
		return nil
	}
	if primaryKey != nil {
		i.indexs = append(i.indexs, primaryKey)
	}
	return primaryKey
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

/*

// 返回普通索引
func (i *Indexs) GetNormalIndexs() []NormalIndex {
	normalIndexs := make([]NormalIndex, 0)
	//普通索引是基类，全部匹配，所以需要判断不是PrimaryKey和FullTextIndex才符合普通索引
	for _, index := range i.indexs {
		if _, ok := index.(NormalIndex); ok {
			if _, ok := index.(PrimaryKey); !ok {
				if _, ok := index.(FullTextIndex); !ok {
					normalIndexs = append(normalIndexs, index.(NormalIndex))
				}
			}
		}
	}
	return normalIndexs
}
*/

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
