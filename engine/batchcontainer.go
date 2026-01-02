package engine

import (
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

type BatchContainer struct {
	indexs *Indexs // 索引集合
	values map[uint8][]byte
	batch  storage.Batch
	tbname string
}

func NewBatchContainer(batch storage.Batch, indexs *Indexs, tbname string) *BatchContainer {
	if batch == nil {
		batch = storage.KVDb.GetBatch()
	}
	return &BatchContainer{
		indexs: indexs,
		batch:  batch,
		tbname: tbname,
		//values3个nil值，key分别为0,1,2
		values: map[uint8][]byte{
			0: nil, //主键值
			1: nil, //普通索引值
			2: nil, //全文索引值
		},
	}
}
func (c *BatchContainer) Add(key []byte, ValueMapKey uint8) {
	//默认规则主键值values[0]为nil，则是Delete；否则是Put
	if c.values[0] == nil {
		c.batch.Delete(key)
	} else {
		c.batch.Put(key, c.values[ValueMapKey])
	}
}
func (c *BatchContainer) SetValue(key uint8, val []byte) error {
	c.values[key] = val
	return nil
}

// get value by key
func (c *BatchContainer) GetValue(key uint8) []byte {
	return c.values[key]
}

// 添加/删除记录操作
// 添加时，key值已经存在的field值，value中会过滤掉，不重复添加。
func (c *BatchContainer) Operation(fieldsBytes map[string][]byte, existFields ...string) {

	pkValue := c.indexs.GetPrimaryKey().JoinValue(&fieldsBytes, c.tbname)
	c.Add(pkValue, 0) //添加主键记录key=pkValue,value=record

	//添加/删除普通索引key=indexValues,value=pkValue
	for _, Normal := range c.indexs.GetNormalIndexs() {
		indexValue := Normal.JoinValue(&fieldsBytes, c.tbname, existFields...)
		c.Add(append([]byte{}, indexValue...), 1) //添加普通索引key=indexValues,value=pkValue
	}
	/*
		//添加/删除全文索引key=joinValue,value=t.primaryKey.ID()
		全文索引的特殊性，在正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义。
		故而系统为了减少大量的字段值重复储存，全文索引的 value值设置 为 nil 。
		查询时，在key值里提取出主键值还原value值。
	*/
	for _, FullText := range c.indexs.GetFullTextIndexs() {
		joinValues := FullText.JoinFullValues(&fieldsBytes, c.tbname, existFields...)
		defer util.PutBytesArray(joinValues)
		for _, joinValue := range joinValues {
			c.Add(append([]byte{}, joinValue...), 2) //添加全文索引key=joinValue,value=t.primaryKey.ID()
		}
	}
}
