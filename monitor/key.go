package monitor

import (
	"encoding/json"
	"sync"
	"sync/atomic"
)

var GlobalKeysMap = &KeysMap{}

type KeysMap struct {
	Data sync.Map // 使用 sync.Map 替代 map + 互斥锁
}

func (m *KeysMap) Inc(key int, tbId uint8, indxName string) {
	value, ok := m.Data.Load(key)
	if !ok {
		// 使用 LoadOrStore 避免竞态条件
		newValue := NewKeys(tbId, indxName)
		value, ok = m.Data.LoadOrStore(key, newValue)
		if !ok {
			value = newValue
		}
	}
	k := value.(*Keys)
	k.Inc()
}
func (m *KeysMap) Dec(key int, tbId uint8, indxName string) {
	value, ok := m.Data.Load(key)
	if !ok {
		return
	}
	k := value.(*Keys)
	k.Dec()
}

func (m *KeysMap) GetAll() map[int]*Keys {
	result := make(map[int]*Keys)
	m.Data.Range(func(key, value interface{}) bool {
		result[key.(int)] = value.(*Keys)
		return true
	})
	return result
}

type Keyfun func(key int, tbId uint8, indxName string)

var KeyInc Keyfun = func(key int, tbId uint8, indxName string) {
	GlobalKeysMap.Inc(key, tbId, indxName)
}
var KeyDec Keyfun = func(key int, tbId uint8, indxName string) {
	GlobalKeysMap.Dec(key, tbId, indxName)
}

// Keys 索引计数器,如果WriteBatch不成功或回滚，都会进行计算，不能准确统计添加/删除次数
type Keys struct {
	TbId        uint8        `json:"tbId"`        //表ID
	IndxName    string       `json:"indxName"`    //索引名称
	PutCount    atomic.Int64 `json:"putCount"`    //添加次数
	DeleteCount atomic.Int64 `json:"deleteCount"` //删除次数
}

func NewKeys(tbId uint8, indxName string) *Keys {
	return &Keys{
		TbId:     tbId,
		IndxName: indxName,
	}
}
func (k *Keys) Inc() {
	k.PutCount.Add(1)
}
func (k *Keys) Dec() {
	k.DeleteCount.Add(1)
}

// MarshalJSON 自定义JSON序列化方法
func (k *Keys) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"TbId":        k.TbId,
		"IndxName":    k.IndxName,
		"PutCount":    k.PutCount.Load(),
		"DeleteCount": k.DeleteCount.Load(),
	})
}
