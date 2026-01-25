package engine

import (
	"github.com/liaoran123/sfsDb/monitor"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

type batchContainer struct {
	indexs  *Indexs          // 索引集合
	values  map[uint8][]byte // 操作值集合
	kvStore storage.Store
	batch   storage.Batch
	tbid    uint8 //表ID
	len     int   // 操作数量

}

func NewBatchContainer(batch storage.Batch, indexs *Indexs, tbid uint8, kvStore storage.Store) *batchContainer {
	if batch == nil {
		batch = storage.KVDb.GetBatch()
	}
	return &batchContainer{
		indexs: indexs,
		batch:  batch,
		tbid:   tbid,
		//commitThreshold: 1000,
		kvStore: kvStore,
		//values3个nil值，key分别为0,1,2
		values: map[uint8][]byte{
			0: nil, //主键值
			1: nil, //普通索引值
			2: nil, //全文索引值
		},
	}
}
func (c *batchContainer) Add(key []byte, ValueMapKey uint8) {
	//默认规则主键值values[0]为nil，则是Delete；否则是Put
	if c.values[0] == nil {
		c.batch.Delete(key)
		//删除索引计数器
		monitor.AtomicMap(monitor.AtomicIntDec).Inc(key[0], key[2]) //key[0]为表ID，key[2]为索引ID
	} else {
		c.batch.Put(key, c.values[ValueMapKey])
		//添加索引计数器
		monitor.AtomicMap(monitor.AtomicInt).Inc(key[0], key[2]) //key[0]为表ID，key[2]为索引ID
	}
	c.len++
}

func (c *batchContainer) SetValue(key uint8, val []byte) error {
	c.values[key] = val
	return nil
}

// get value by key
func (c *batchContainer) GetValue(key uint8) []byte {
	return c.values[key]
}

// 添加/删除记录操作
// 添加时，key值已经存在的field值，value中会过滤掉，不重复添加。
func (c *batchContainer) Operation(fieldsBytes *map[string][]byte, existFields ...string) {
	// batch并发安全，防止多个goroutine同时操作
	pkValue := c.indexs.getPrimaryKey().JoinValue(fieldsBytes, c.tbid)
	c.Add(pkValue, 0) //添加主键记录key=pkValue,value=record

	//添加/删除普通索引key=indexValues,value=pkValue
	for _, Normal := range c.indexs.GetNormalIndexs() {
		indexValue := Normal.JoinValue(fieldsBytes, c.tbid, existFields...)
		if indexValue == nil {
			continue
		}
		c.Add(append([]byte{}, indexValue...), 1) //添加普通索引key=indexValues,value=pkValue
	}
	/*
		//添加/删除全文索引key=joinValue,value=t.primaryKey.ID()
		全文索引的特殊性，在正常情况下必须带上主键，否则后面的关键词都被覆盖，失去全文索引的意义。
		故而系统为了减少大量的字段值重复储存，全文索引的 value值设置 为 nil 。
		查询时，在key值里提取出主键值还原value值。
	*/
	for _, FullText := range c.indexs.GetFullTextIndexs() {
		joinValues := FullText.JoinFullValues(fieldsBytes, c.tbid, existFields...)
		defer util.PutBytesArray(joinValues)
		for _, joinValue := range joinValues {
			if joinValue == nil {
				continue
			}
			c.Add(append([]byte{}, joinValue...), 2) //添加全文索引key=joinValue,value=t.primaryKey.ID()
		}
	}

}
func (c *batchContainer) Len() int {
	return c.len
}
