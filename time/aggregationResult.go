package time

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/record"
)

// TimeAggregationResult 时间聚合结果
type TimeAggregationResult struct {
	TimeKey string  `json:"time_key"`
	Value   float64 `json:"value"`
}

// AggregateByTimeGranularity 按时间粒度聚合数据
func AggregateByTimeGranularity(records record.Records, timeField string, valueField string,
	granularity TimeGranularity, aggregationType string) ([]TimeAggregationResult, error) {

	// 1. 为每个记录添加时间粒度键
	for i := range records {
		if records[i] == nil {
			continue
		}

		timestamp, ok := records[i][timeField].(time.Time)
		if !ok {
			return nil, fmt.Errorf("field %s is not a time.Time", timeField)
		}

		// 添加时间粒度键
		timeKey := FormatTimeByGranularity(timestamp, granularity)
		records[i]["time_granularity_key"] = timeKey
	}

	// 2. 按时间粒度键分组
	groupOp := record.NewGroupVerticalOperation("time_granularity_key", "grouped_records")
	groupedResult := records.OperationVertical(groupOp)

	// 3. 提取分组结果
	if len(groupedResult) == 0 {
		return []TimeAggregationResult{}, nil
	}

	groupedData, ok := groupedResult[0]["grouped_records"].(map[any]record.Records)
	if !ok {
		return nil, fmt.Errorf("failed to get grouped records")
	}

	// 4. 对每个分组应用聚合函数
	var results []TimeAggregationResult
	for timeKey, groupRecords := range groupedData {
		var aggOp record.VerticalOperation

		// 根据聚合类型创建相应的聚合操作
		switch aggregationType {
		case "sum":
			aggOp = record.NewSumVerticalOperation(valueField, "aggregated_value")
		case "avg":
			aggOp = record.NewAvgVerticalOperation(valueField, "aggregated_value")
		case "count":
			aggOp = record.NewCountVerticalOperation(valueField, "aggregated_value")
		case "max":
			aggOp = record.NewMaxVerticalOperation(valueField, "aggregated_value")
		case "min":
			aggOp = record.NewMinVerticalOperation(valueField, "aggregated_value")
		default:
			return nil, fmt.Errorf("unsupported aggregation type: %s", aggregationType)
		}

		// 应用聚合操作
		aggResult := groupRecords.OperationVertical(aggOp)
		if len(aggResult) > 0 {
			aggregatedValue, ok := aggResult[0]["aggregated_value"].(float64)
			if ok {
				results = append(results, TimeAggregationResult{
					TimeKey: timeKey.(string),
					Value:   aggregatedValue,
				})
			}
		}
	}

	return results, nil
}
