package monitor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	GIndexStatsMap = NewIndexStatsMap()
}

var GIndexStatsMap *IndexStatsMap

type IndexStatsMap struct {
	Data  map[int]*IndexStats `json:"data"`
	mutex sync.RWMutex        // 添加互斥锁确保并发安全
}

func NewIndexStatsMap() *IndexStatsMap {
	return &IndexStatsMap{
		Data: make(map[int]*IndexStats),
	}
}

func (i *IndexStatsMap) Settime(indexKey int, duration time.Duration, tblName string, indxName string) {
	i.mutex.Lock()

	if _, ok := i.Data[indexKey]; !ok {
		i.Data[indexKey] = NewIndexStats(tblName, indxName)
	}

	// 先获取IndexStats实例，然后释放锁，因为RecordTime方法内部已经是并发安全的
	stats := i.Data[indexKey]
	i.mutex.Unlock()

	stats.Inc()
	stats.RecordTime(duration)
}

//var IndexStatsMap = make(map[int]*IndexStats)

type IndexStats struct {
	TblName   string        `json:"tblName"`
	IndxName  string        `json:"indxName"`
	Count     atomic.Int64  `json:"count"`     //搜索次数
	AvgTime   time.Duration `json:"avgTime"`   //平均搜索耗时
	TotalTime atomic.Int64  `json:"totalTime"` //总搜索耗时（纳秒）
	MaxTime   atomic.Int64  `json:"maxTime"`   //最大搜索耗时（纳秒）
	MinTime   atomic.Int64  `json:"minTime"`   //最小搜索耗时（纳秒）
}

func NewIndexStats(tblName string, indxName string) *IndexStats {
	stats := &IndexStats{
		TblName:  tblName,
		IndxName: indxName,
	}
	// 初始化为较大的值
	stats.MinTime.Store(int64(time.Hour * 24 * 365))
	return stats
}
func (i *IndexStats) Inc() {
	i.Count.Add(1)
}

// 搜索用时
func (i *IndexStats) RecordTime(duration time.Duration) {
	// 原子更新总时间
	oldTotal := i.TotalTime.Load()
	newTotal := oldTotal + int64(duration)
	i.TotalTime.Store(newTotal)

	// 计算平均时间
	count := i.Count.Load()
	if count > 0 {
		i.AvgTime = time.Duration(newTotal / count)
	}

	// 原子更新最大时间
	oldMax := i.MaxTime.Load()
	if int64(duration) > oldMax {
		i.MaxTime.Store(int64(duration))
	}

	// 原子更新最小时间
	oldMin := i.MinTime.Load()
	if int64(duration) < oldMin {
		i.MinTime.Store(int64(duration))
	}
}
func (i *IndexStats) GetCount() (int64, time.Duration, time.Duration, time.Duration) {
	return i.Count.Load(), time.Duration(i.TotalTime.Load()), time.Duration(i.MaxTime.Load()), time.Duration(i.MinTime.Load())
}

// MarshalJSON 自定义JSON序列化方法
func (i *IndexStats) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"tblName":   i.TblName,
		"indxName":  i.IndxName,
		"count":     i.Count.Load(),
		"avgTime":   i.AvgTime,
		"totalTime": time.Duration(i.TotalTime.Load()),
		"maxTime":   time.Duration(i.MaxTime.Load()),
		"minTime":   time.Duration(i.MinTime.Load()),
	})
}

// 将2个uint8合并为字符串，然后将字符串转为int
func GetIndexKey(tbid, idxid uint8) int {
	// 将tbid和idxid转换为字符串并合并
	keyStr := fmt.Sprintf("%d%d", tbid, idxid)
	// 转换为整数
	key, err := strconv.Atoi(keyStr)
	if err != nil {
		// 如果转换失败，返回默认值
		return 0
	}
	return key
}
