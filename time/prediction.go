package time

import (
	"math"
	"time"
)

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	Time  time.Time
	Value float64
}

// MovingAveragePrediction 移动平均预测结果
type MovingAveragePrediction struct {
	PredictedPoints []TimeSeriesPoint
	WindowSize      int
}

// LinearRegressionPrediction 线性回归预测结果
type LinearRegressionPrediction struct {
	PredictedPoints []TimeSeriesPoint
	Slope           float64
	Intercept       float64
	R2              float64 // 决定系数，衡量模型拟合度
}

// NewMovingAveragePrediction 创建移动平均预测
// points: 历史数据点
// windowSize: 移动窗口大小
// predictCount: 预测点数量
// interval: 预测点时间间隔
func NewMovingAveragePrediction(points []TimeSeriesPoint, windowSize, predictCount int, interval time.Duration) *MovingAveragePrediction {
	if len(points) < windowSize {
		return &MovingAveragePrediction{
			PredictedPoints: []TimeSeriesPoint{},
			WindowSize:      windowSize,
		}
	}

	// 计算移动平均值
	var movingAverages []float64
	for i := windowSize - 1; i < len(points); i++ {
		sum := 0.0
		for j := i - windowSize + 1; j <= i; j++ {
			sum += points[j].Value
		}
		avg := sum / float64(windowSize)
		movingAverages = append(movingAverages, avg)
	}

	// 生成预测点
	var predictedPoints []TimeSeriesPoint
	lastPoint := points[len(points)-1]
	lastMA := movingAverages[len(movingAverages)-1]

	for i := 1; i <= predictCount; i++ {
		predictedTime := lastPoint.Time.Add(time.Duration(i) * interval)
		// 简单预测：使用最后一个移动平均值
		predictedPoints = append(predictedPoints, TimeSeriesPoint{
			Time:  predictedTime,
			Value: lastMA,
		})
	}

	return &MovingAveragePrediction{
		PredictedPoints: predictedPoints,
		WindowSize:      windowSize,
	}
}

// LinearRegressionResult 线性回归结果
type LinearRegressionResult struct {
	Slope     float64
	Intercept float64
	R2        float64
}

// calculateLinearRegression 计算线性回归
func calculateLinearRegression(points []TimeSeriesPoint) LinearRegressionResult {
	n := float64(len(points))
	if n < 2 {
		return LinearRegressionResult{0, 0, 0}
	}

	// 计算x和y的平均值
	var sumX, sumY, sumXY, sumXX, sumYY float64
	for i, point := range points {
		x := float64(i) // 使用索引作为x值
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
		sumYY += y * y
	}

	// 计算斜率和截距
	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// 计算决定系数R²
	var ssRes, ssTot float64
	meanY := sumY / n
	for i, point := range points {
		x := float64(i)
		yPred := slope*x + intercept
		ssRes += math.Pow(point.Value-yPred, 2)
		ssTot += math.Pow(point.Value-meanY, 2)
	}

	r2 := 1 - (ssRes / ssTot)
	if math.IsNaN(r2) {
		r2 = 0
	}

	return LinearRegressionResult{
		Slope:     slope,
		Intercept: intercept,
		R2:        r2,
	}
}

// NewLinearRegressionPrediction 创建线性回归预测
// points: 历史数据点
// predictCount: 预测点数量
// interval: 预测点时间间隔
func NewLinearRegressionPrediction(points []TimeSeriesPoint, predictCount int, interval time.Duration) *LinearRegressionPrediction {
	if len(points) < 2 {
		return &LinearRegressionPrediction{
			PredictedPoints: []TimeSeriesPoint{},
			Slope:           0,
			Intercept:       0,
			R2:              0,
		}
	}

	// 计算线性回归
	lrResult := calculateLinearRegression(points)

	// 生成预测点
	var predictedPoints []TimeSeriesPoint
	lastPoint := points[len(points)-1]
	lastIndex := float64(len(points) - 1)

	for i := 1; i <= predictCount; i++ {
		predictedTime := lastPoint.Time.Add(time.Duration(i) * interval)
		predictedIndex := lastIndex + float64(i)
		predictedValue := lrResult.Slope*predictedIndex + lrResult.Intercept
		predictedPoints = append(predictedPoints, TimeSeriesPoint{
			Time:  predictedTime,
			Value: predictedValue,
		})
	}

	return &LinearRegressionPrediction{
		PredictedPoints: predictedPoints,
		Slope:           lrResult.Slope,
		Intercept:       lrResult.Intercept,
		R2:              lrResult.R2,
	}
}

// PredictTimeSeries 预测时间序列数据
// points: 历史数据点
// method: 预测方法 ("moving_average" 或 "linear_regression")
// params: 预测参数
//   - 对于 moving_average: {"window_size": int, "predict_count": int, "interval": time.Duration}
//   - 对于 linear_regression: {"predict_count": int, "interval": time.Duration}
func PredictTimeSeries(points []TimeSeriesPoint, method string, params map[string]any) (any, error) {
	switch method {
	case "moving_average":
		windowSize, ok1 := params["window_size"].(int)
		predictCount, ok2 := params["predict_count"].(int)
		interval, ok3 := params["interval"].(time.Duration)
		if !ok1 || !ok2 || !ok3 {
			return nil, nil
		}
		return NewMovingAveragePrediction(points, windowSize, predictCount, interval), nil

	case "linear_regression":
		predictCount, ok1 := params["predict_count"].(int)
		interval, ok2 := params["interval"].(time.Duration)
		if !ok1 || !ok2 {
			return nil, nil
		}
		return NewLinearRegressionPrediction(points, predictCount, interval), nil

	default:
		return nil, nil
	}
}
