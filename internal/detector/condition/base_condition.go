package condition

import (
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/shopspring/decimal"
)

// BaseCondition 基础条件，包含所有条件的通用字段和方法
type BaseCondition struct {
	Name        string
	Description string
	TimeWindow  string // "5m", "30s", "1m" 等
	Operator    string // ">=", ">", "<=", "<", "=="
}

func (c *BaseCondition) GetName() string {
	return c.Name
}

func (c *BaseCondition) GetDescription() string {
	return c.Description
}

// getStatsForTimeWindow 根据时间窗口获取对应的统计数据
func (c *BaseCondition) getStatsForTimeWindow(context *EvaluationContext) *model.TokenStats {
	switch c.TimeWindow {
	case "5m":
		return context.Stats5m
	case "30s":
		return context.Stats30s
	case "1m":
		return context.Stats1m
	default:
		return nil
	}
}

// CompareDecimal 统一的decimal比较方法
func (c *BaseCondition) CompareDecimal(value, threshold decimal.Decimal) bool {
	switch c.Operator {
	case ">=":
		return value.GreaterThanOrEqual(threshold)
	case ">":
		return value.GreaterThan(threshold)
	case "<=":
		return value.LessThanOrEqual(threshold)
	case "<":
		return value.LessThan(threshold)
	case "==":
		return value.Equal(threshold)
	default:
		return false
	}
}

// CompareInt 统一的int比较方法
func (c *BaseCondition) CompareInt(value, threshold int) bool {
	switch c.Operator {
	case ">=":
		return value >= threshold
	case ">":
		return value > threshold
	case "<=":
		return value <= threshold
	case "<":
		return value < threshold
	case "==":
		return value == threshold
	default:
		return false
	}
}

// CompareFloat64 统一的float64比较方法
func (c *BaseCondition) CompareFloat64(value, threshold float64) bool {
	switch c.Operator {
	case ">=":
		return value >= threshold
	case ">":
		return value > threshold
	case "<=":
		return value <= threshold
	case "<":
		return value < threshold
	case "==":
		return value == threshold
	default:
		return false
	}
}
