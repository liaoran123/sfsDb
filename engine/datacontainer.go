package engine

import (
	"github.com/liaoran123/sfsDb/storage"
)

/*
//###数据集合容器接口
用于集中存储不同出处的相同数据进行集中处理
特别是数据库这类存储系统，需要批量集中处理数据，特别是用户手动事务的情况。
*/
type DataContainer interface {
	Add(val []byte)
	SetValue(val []byte)
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
	Value []byte
	batch storage.Batch
}

func NewPutBatchContainer(batch storage.Batch) *PutBatchContainer {
	if batch == nil {
		batch = storage.KVDb.GetBatch()
	}
	return &PutBatchContainer{
		batch: batch,
	}
}
func (c *PutBatchContainer) Add(val []byte) {
	c.batch.Put(val, c.Value)
}
func (c *PutBatchContainer) SetValue(val []byte) {
	c.Value = val
}

//---------------------------------------

type DelteBatchContainer struct {
	batch storage.Batch
}

func NewDelteBatchContainer(batch storage.Batch) *DelteBatchContainer {
	if batch == nil {
		batch = storage.KVDb.GetBatch()
	}
	return &DelteBatchContainer{
		batch: batch,
	}
}
func (c *DelteBatchContainer) Add(val []byte) {
	c.batch.Delete(val)
}
func (c *DelteBatchContainer) SetValue(val []byte) {
	//c.Value = val
}
