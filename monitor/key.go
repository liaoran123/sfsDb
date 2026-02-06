package monitor

import (
	"encoding/json"
	"sync"
	"sync/atomic"
)

var GlobalKeysMap *KeysMap

func init() {
	GlobalKeysMap = NewKeysMap()
}

type KeysMap struct {
	Data  map[int]*Keys `json:"data"`
	mutex sync.RWMutex  // 添加互斥锁确保并发安全
}

func NewKeysMap() *KeysMap {
	return &KeysMap{
		Data: make(map[int]*Keys),
	}
}
func (m *KeysMap) Inc(key int, tbId uint8, indxName string) {
	m.mutex.Lock()
	if _, ok := m.Data[key]; !ok {
		m.Data[key] = NewKeys(tbId, indxName)
	}
	k := m.Data[key]
	m.mutex.Unlock()
	k.Inc()
}
func (m *KeysMap) Dec(key int, tbId uint8, indxName string) {
	m.mutex.Lock()
	if _, ok := m.Data[key]; !ok {
		return
		//m.Data[key] = NewKeys(tbId, indxName)
	}
	k := m.Data[key]
	m.mutex.Unlock()
	k.Dec()
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
		"tbId":        k.TbId,
		"indxName":    k.IndxName,
		"putCount":    k.PutCount.Load(),
		"deleteCount": k.DeleteCount.Load(),
	})
}

// GetAllCounters 获取所有计数器数据
// 返回:
//
//	map[string]int64: put操作的计数器数据
