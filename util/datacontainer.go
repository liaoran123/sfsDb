package util

import (
	"github.com/liaoran123/sfsDb/storage"
)

/*
//###数据集合容器接口
用于集中存储不同出处的相同数据进行集中处理
特别是数据库这类存储系统，需要批量集中处理数据，特别是用户手动事务的情况。
*/
type DataContainer interface {
	Add(key uint8, val []byte)
	SetValue(key uint8, val []byte)
}
type BytesContainer struct {
	values [][]byte
}

func NewBytesContainer(bs [][]byte) *BytesContainer {
	if bs == nil {
		bs = make([][]byte, 0)
	}
	return &BytesContainer{
		values: bs,
	}
}

/*
//###字节数组容器
用于存储字节数组类型的数据
*/
func (c *BytesContainer) Add(val []byte) {
	c.values = append(c.values, val)
}
func (c *BytesContainer) SetValue(val []byte) {
	c.values = [][]byte{val}
}

//---------------------------------------

type PutBatchContainer struct {
	Values map[uint8][]byte
	batch  storage.Batch
}

func NewPutBatchContainer(batch storage.Batch) *PutBatchContainer {
	if batch == nil {
		batch = storage.KVDb.GetBatch()
	}
	return &PutBatchContainer{
		//Values3个nil值，key分别为0,1,2
		Values: map[uint8][]byte{
			0: nil, //主键值
			1: nil, //普通索引值
			2: nil, //全文索引值
		},
		batch: batch,
	}
}
func (c *PutBatchContainer) Add(key uint8, val []byte) {
	c.batch.Put(val, c.Values[key])
}
func (c *PutBatchContainer) SetValue(key uint8, val []byte) {
	c.Values[key] = val
}

// get value by key
func (c *PutBatchContainer) GetValue(key uint8) []byte {
	return c.Values[key]
}

//---------------------------------------

type DelteBatchContainer struct {
	Values map[uint8][]byte
	batch  storage.Batch
}

func NewDelteBatchContainer(batch storage.Batch) *DelteBatchContainer {
	if batch == nil {
		batch = storage.KVDb.GetBatch()
	}
	return &DelteBatchContainer{
		Values: map[uint8][]byte{
			0: nil, //主键值
			1: nil, //普通索引值
			2: nil, //全文索引值
		},
		batch: batch,
	}
}
func (c *DelteBatchContainer) Add(key uint8, val []byte) {
	c.batch.Delete(val)
}
func (c *DelteBatchContainer) SetValue(key uint8, val []byte) {
	//c.Value = val
}

// get value by key
func (c *DelteBatchContainer) GetValue(key uint8) []byte {
	return c.Values[key]
}
