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

// 索引评分结构体，用于排序匹配的索引
// 支持复杂SQL查询的索引优先级评估
type IndexScore struct {
	Index         Index   // 匹配的索引
	Score         int     // 评分，越高优先级越高
	IndexType     int     // 索引类型: 0=PrimaryKey, 1=NormalIndex, 2=FullTextIndex
	MatchCount    int     // 匹配的字段数量
	IsPrefixMatch bool    // 是否为前缀匹配（仅针对联合索引）
	FieldCoverage float64 // 字段覆盖率
}

// MatchIndexes 匹配所有适用的索引，返回按照优先级排序的索引列表
// 支持复杂SQL查询，可匹配多个索引
// 参数：
//
//	fields: 查询涉及的字段列表
//
// 返回：
//
//	按照优先级排序的索引列表，优先级从高到低
func (i *Indexs) MatchIndexes(fields ...string) []Index {
	if len(fields) == 0 {
		return nil
	}

	// 收集所有匹配的索引并评分
	var matchedIndexes []IndexScore

	for _, index := range i.indexs {
		indexFields := index.GetFields()
		matchCount := 0
		isPrefixMatch := false

		// 检查索引字段与查询字段的匹配情况
		for idx, idxField := range indexFields {
			if slices.Contains(fields, idxField) {
				matchCount++
			} else {
				// 对于联合索引，检查是否为前缀匹配
				if idx == 0 {
					// 第一个字段不匹配，该索引完全不适用
					matchCount = 0
				}
				break
			}
		}

		// 如果有匹配的字段
		if matchCount > 0 {
			// 判断是否为前缀匹配（匹配所有索引字段或前缀）
			isPrefixMatch = matchCount > 0

			// 计算索引类型分数
			indexTypeScore := 0
			indexType := 2 // 默认FullTextIndex

			// 使用类型断言检查索引类型
			if _, ok := index.(PrimaryKey); ok {
				indexTypeScore = 300 // 主键索引优先级最高
				indexType = 0
			} else if _, ok := index.(NormalIndex); ok {
				indexTypeScore = 200 // 普通索引次之
				indexType = 1
			} else if _, ok := index.(FullTextIndex); ok {
				indexTypeScore = 100 // 全文索引优先级最低
				indexType = 2
			}

			// 计算匹配字段分数
			matchScore := matchCount * 10

			// 计算字段覆盖率
			fieldCoverage := float64(matchCount) / float64(len(indexFields))

			// 前缀匹配加分
			prefixScore := 0
			if matchCount == len(indexFields) {
				prefixScore = 5 // 完全匹配加分
			}

			// 总分数 = 索引类型分数 + 匹配字段分数 + 前缀匹配分数
			totalScore := indexTypeScore + matchScore + prefixScore

			matchedIndexes = append(matchedIndexes, IndexScore{
				Index:         index,
				Score:         totalScore,
				IndexType:     indexType,
				MatchCount:    matchCount,
				IsPrefixMatch: isPrefixMatch,
				FieldCoverage: fieldCoverage,
			})
		}
	}

	// 按照分数降序排序
	slices.SortFunc(matchedIndexes, func(a, b IndexScore) int {
		if a.Score != b.Score {
			return b.Score - a.Score // 分数高的排在前面
		}
		// 分数相同，索引类型优先级：PrimaryKey > NormalIndex > FullTextIndex
		if a.IndexType != b.IndexType {
			return a.IndexType - b.IndexType
		}
		// 索引类型相同，匹配字段数量多的优先
		if a.MatchCount != b.MatchCount {
			return b.MatchCount - a.MatchCount
		}
		// 匹配字段数量相同，字段覆盖率高的优先
		if a.FieldCoverage != b.FieldCoverage {
			if a.FieldCoverage > b.FieldCoverage {
				return -1
			}
			return 1
		}
		return 0
	})

	// 提取排序后的索引
	result := make([]Index, len(matchedIndexes))
	for i, idxScore := range matchedIndexes {
		result[i] = idxScore.Index
	}

	return result
}

// MatchBestIndex 匹配最优的单个索引
// 支持复杂SQL查询，选择最适合的单个索引
func (i *Indexs) MatchBestIndex(fields ...string) Index {
	matchedIndexes := i.MatchIndexes(fields...)
	if len(matchedIndexes) > 0 {
		return matchedIndexes[0] // 返回分数最高的索引
	}
	return nil
}

// IndexCombination 索引组合结构体，用于复杂查询的索引组合
type IndexCombination struct {
	Indexes     []Index
	TotalScore  int
	MatchFields []string
	Coverage    float64
}

// MatchIndexCombination 匹配最优的索引组合，用于复杂查询
// 支持多条件查询的索引组合优化
func (i *Indexs) MatchIndexCombination(fields ...string) *IndexCombination {
	if len(fields) == 0 {
		return nil
	}

	// 获取所有匹配的索引
	matchedIndexes := i.MatchIndexes(fields...)
	if len(matchedIndexes) == 0 {
		return nil
	}

	// 对于复杂查询，选择分数最高的索引作为主索引
	// 可以根据实际需求扩展为更复杂的索引组合算法
	mainIndex := matchedIndexes[0]
	mainIndexFields := mainIndex.GetFields()

	// 计算匹配的字段
	var matchedFields []string
	for _, field := range fields {
		if slices.Contains(mainIndexFields, field) {
			matchedFields = append(matchedFields, field)
		}
	}

	// 计算覆盖率
	coverage := float64(len(matchedFields)) / float64(len(fields))

	return &IndexCombination{
		Indexes:     []Index{mainIndex},
		TotalScore:  100, // 简化评分，实际可根据组合效果计算
		MatchFields: matchedFields,
		Coverage:    coverage,
	}
}

// GetIndexByType 根据索引类型获取索引列表
func (i *Indexs) GetIndexByType(indexType any) []Index {
	var result []Index
	for _, index := range i.indexs {
		switch indexType.(type) {
		case PrimaryKey:
			if _, ok := index.(PrimaryKey); ok {
				result = append(result, index)
			}
		case NormalIndex:
			if _, ok := index.(NormalIndex); ok {
				result = append(result, index)
			}
		case FullTextIndex:
			if _, ok := index.(FullTextIndex); ok {
				result = append(result, index)
			}
		}
	}
	return result
}
