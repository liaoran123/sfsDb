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
}

func NewBatchContainer(batch storage.Batch, indexs *Indexs, tbid uint8, kvStore storage.Store) *batchContainer {
	// 确保batch不为nil
	if batch == nil {
		// 先检查kvStore是否为nil
		if kvStore != nil {
			batch = kvStore.GetBatch()
		}
		// 如果还是nil，创建一个新的batch
		if batch == nil {
			panic("failed to create batch")
		}
	}
	return &batchContainer{
		indexs:  indexs,
		batch:   batch,
		tbid:    tbid,
		kvStore: kvStore,
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
		//fmt.Println("Delete:", string(key))
	} else {
		c.batch.Put(key, c.values[ValueMapKey])
		//fmt.Println("Put:", string(key), string(c.values[ValueMapKey]))
	}
}

func (c *batchContainer) SetValue(key uint8, val []byte) error {
	c.values[key] = val
	return nil
}

// get value by key
func (c *batchContainer) GetValue(key uint8) []byte {
	return c.values[key]
}

// 添加/删除记录操作。修改操作，通过existFields，可以指定需要更新（包括加/删除）的字段。
// 这个是数据库的核心枢纽。通过索引为桥梁组织key值。
// 通过fieldsBytes *map[string][]byte 组织各个索引的key值。
// value值为record（主键值）和GetPrimaryKey().GetID()（其他索引指向的主键值），由外部预先传入batchContainer.values  map[uint8][]byte // 操作值集合。
func (c *batchContainer) Operation(fieldsBytes *map[string][]byte, existFields ...string) {
	var mapkey int
	var KeyFun monitor.Keyfun
	if c.values[0] == nil { //判断是Delete操作还是Put操作，记录每一个put,delete操作次数。
		KeyFun = monitor.KeyDec
	} else {
		KeyFun = monitor.KeyInc
	}
	for _, index := range c.indexs.GetAllIndexes() {
		switch index := index.(type) {
		case PrimaryKey:
			// batch并发安全，防止多个goroutine同时操作
			pkValue := c.indexs.getPrimaryKey().JoinValue(fieldsBytes, c.tbid)
			c.Add(pkValue, 0) //添加主键记录key=pkValue,value=record
			//添加主键索引计数器
			mapkey = monitor.GetIndexKey(c.tbid, index.GetId())
			KeyFun(mapkey, c.tbid, index.Name())
		case FullTextIndex:
			mapkey = monitor.GetIndexKey(c.tbid, index.GetId())
			joinValues := index.JoinFullValues(fieldsBytes, c.tbid, existFields...)
			defer util.PutBytesArray(joinValues)
			for _, joinValue := range joinValues {
				if joinValue == nil {
					continue
				}
				c.Add(append([]byte{}, joinValue...), 2) //添加全文索引key=joinValue,value=t.primaryKey.ID()
				KeyFun(mapkey, c.tbid, index.Name())
			}
		default:
			indexValue := index.JoinValue(fieldsBytes, c.tbid, existFields...)
			if indexValue == nil {
				continue
			}
			c.Add(append([]byte{}, indexValue...), 1) //添加普通索引key=indexValues,value=pkValue
			mapkey = monitor.GetIndexKey(c.tbid, index.GetId())
			KeyFun(mapkey, c.tbid, index.Name())
		}
	}
}

func (c *batchContainer) Len() int {
	return c.batch.Len()
}
