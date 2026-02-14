package time

import (
	"encoding/binary"
	"math"
	"time"
)

// CompressedTimeSeries 压缩后的时间序列数据
type CompressedTimeSeries struct {
	StartTime    time.Time
	Interval     time.Duration
	CompressedValues []byte
	CompressionType  string
}

// CompressTimeSeries 压缩时间序列数据
// points: 时间序列数据点
// compressionType: 压缩类型 ("delta" 或 "rle")
// interval: 时间间隔（如果为0，则自动计算）
func CompressTimeSeries(points []TimeSeriesPoint, compressionType string, interval time.Duration) (*CompressedTimeSeries, error) {
	if len(points) == 0 {
		return nil, nil
	}

	// 如果未指定时间间隔，自动计算
	if interval == 0 && len(points) > 1 {
		interval = points[1].Time.Sub(points[0].Time)
	}

	var compressedValues []byte

	switch compressionType {
	case "delta":
		compressedValues = compressDelta(points)
	case "rle":
		compressedValues = compressRLE(points)
	default:
		compressedValues = compressDelta(points) // 默认使用delta编码
	}

	return &CompressedTimeSeries{
		StartTime:       points[0].Time,
		Interval:        interval,
		CompressedValues: compressedValues,
		CompressionType:  compressionType,
	}, nil
}

// DecompressTimeSeries 解压缩时间序列数据
func DecompressTimeSeries(cts *CompressedTimeSeries, count int) ([]TimeSeriesPoint, error) {
	if cts == nil || len(cts.CompressedValues) == 0 {
		return []TimeSeriesPoint{}, nil
	}

	var points []TimeSeriesPoint

	switch cts.CompressionType {
	case "delta":
		points = decompressDelta(cts.CompressedValues, cts.StartTime, cts.Interval, count)
	case "rle":
		points = decompressRLE(cts.CompressedValues, cts.StartTime, cts.Interval, count)
	default:
		points = decompressDelta(cts.CompressedValues, cts.StartTime, cts.Interval, count) // 默认使用delta解码
	}

	return points, nil
}

// compressDelta 使用Delta编码压缩时间序列数据
func compressDelta(points []TimeSeriesPoint) []byte {
	if len(points) == 0 {
		return []byte{}
	}

	// 存储第一个值
	var compressed []byte
	firstValue := points[0].Value
	firstValueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(firstValueBytes, math.Float64bits(firstValue))
	compressed = append(compressed, firstValueBytes...)

	// 存储后续值的delta
	prevValue := firstValue
	for i := 1; i < len(points); i++ {
		delta := points[i].Value - prevValue
		deltaBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(deltaBytes, math.Float64bits(delta))
		compressed = append(compressed, deltaBytes...)
		prevValue = points[i].Value
	}

	return compressed
}

// decompressDelta 解压缩Delta编码的时间序列数据
func decompressDelta(compressed []byte, startTime time.Time, interval time.Duration, count int) []TimeSeriesPoint {
	if len(compressed) < 8 {
		return []TimeSeriesPoint{}
	}

	var points []TimeSeriesPoint

	// 读取第一个值
	firstValue := math.Float64frombits(binary.LittleEndian.Uint64(compressed[:8]))
	points = append(points, TimeSeriesPoint{
		Time:  startTime,
		Value: firstValue,
	})

	// 读取后续值的delta
	prevValue := firstValue
	for i := 1; i < count && (i*8+8) <= len(compressed); i++ {
		deltaBytes := compressed[i*8 : (i+1)*8]
		delta := math.Float64frombits(binary.LittleEndian.Uint64(deltaBytes))
		currentValue := prevValue + delta
		points = append(points, TimeSeriesPoint{
			Time:  startTime.Add(time.Duration(i) * interval),
			Value: currentValue,
		})
		prevValue = currentValue
	}

	return points
}

// compressRLE 使用Run-Length Encoding压缩时间序列数据
func compressRLE(points []TimeSeriesPoint) []byte {
	if len(points) == 0 {
		return []byte{}
	}

	var compressed []byte

	// 压缩算法：存储值和连续出现的次数
	currentValue := points[0].Value
	count := 1

	for i := 1; i < len(points); i++ {
		if points[i].Value == currentValue {
			count++
		} else {
			// 存储当前值和计数
			valueBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(valueBytes, math.Float64bits(currentValue))
			compressed = append(compressed, valueBytes...)

			countBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(countBytes, uint32(count))
			compressed = append(compressed, countBytes...)

			// 开始新的序列
			currentValue = points[i].Value
			count = 1
		}
	}

	// 存储最后一个序列
	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, math.Float64bits(currentValue))
	compressed = append(compressed, valueBytes...)

	countBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(countBytes, uint32(count))
	compressed = append(compressed, countBytes...)

	return compressed
}

// decompressRLE 解压缩Run-Length Encoding的时间序列数据
func decompressRLE(compressed []byte, startTime time.Time, interval time.Duration, count int) []TimeSeriesPoint {
	if len(compressed) < 12 {
		return []TimeSeriesPoint{}
	}

	var points []TimeSeriesPoint
	timeIndex := 0

	for i := 0; i < len(compressed) && len(points) < count; i += 12 {
		if i+12 > len(compressed) {
			break
		}

		// 读取值和计数
		value := math.Float64frombits(binary.LittleEndian.Uint64(compressed[i:i+8]))
		runCount := int(binary.LittleEndian.Uint32(compressed[i+8:i+12]))

		// 生成数据点
		for j := 0; j < runCount && len(points) < count; j++ {
			points = append(points, TimeSeriesPoint{
				Time:  startTime.Add(time.Duration(timeIndex) * interval),
				Value: value,
			})
			timeIndex++
		}
	}

	return points
}

// GetCompressionRatio 计算压缩率
func GetCompressionRatio(originalSize, compressedSize int) float64 {
	if originalSize == 0 {
		return 0
	}
	return float64(originalSize) / float64(compressedSize)
}
