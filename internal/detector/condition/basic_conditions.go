package condition

import (
	"github.com/shopspring/decimal"
)

// PriceChangeCondition 价格变化条件
type PriceChangeCondition struct {
	BaseCondition
	Threshold decimal.Decimal // 阈值百分比
}

func (c *PriceChangeCondition) Evaluate(context *EvaluationContext) bool {
	stats := c.getStatsForTimeWindow(context)
	if stats == nil {
		return false
	}
	return c.CompareDecimal(stats.PriceChangePercent, c.Threshold)
}

// TransactionCountCondition 交易数量条件
type TransactionCountCondition struct {
	BaseCondition
	Threshold int
}

func (c *TransactionCountCondition) Evaluate(context *EvaluationContext) bool {
	stats := c.getStatsForTimeWindow(context)
	if stats == nil {
		return false
	}
	return c.CompareInt(stats.TxCount5m, c.Threshold)
}

// VolumeCondition 交易量条件
type VolumeCondition struct {
	BaseCondition
	Threshold decimal.Decimal
}

func (c *VolumeCondition) Evaluate(context *EvaluationContext) bool {
	stats := c.getStatsForTimeWindow(context)
	if stats == nil {
		return false
	}
	return c.CompareDecimal(stats.Volume5m, c.Threshold)
}

// UniqueWalletsCondition 独立钱包数条件
type UniqueWalletsCondition struct {
	BaseCondition
	Threshold int
}

func (c *UniqueWalletsCondition) Evaluate(context *EvaluationContext) bool {
	stats := c.getStatsForTimeWindow(context)
	if stats == nil {
		return false
	}
	return c.CompareInt(stats.UniqueHolders, c.Threshold)
}

// ConditionFactory 条件工厂，支持从配置创建条件
type ConditionFactory struct{}

func NewConditionFactory() *ConditionFactory {
	return &ConditionFactory{}
}

func (f *ConditionFactory) CreatePriceChangeCondition(name, desc, timeWindow, operator string, threshold float64) Condition {
	return &PriceChangeCondition{
		BaseCondition: BaseCondition{
			Name:        name,
			Description: desc,
			TimeWindow:  timeWindow,
			Operator:    operator,
		},
		Threshold: decimal.NewFromFloat(threshold),
	}
}

func (f *ConditionFactory) CreateTransactionCountCondition(name, desc, timeWindow, operator string, threshold int) Condition {
	return &TransactionCountCondition{
		BaseCondition: BaseCondition{
			Name:        name,
			Description: desc,
			TimeWindow:  timeWindow,
			Operator:    operator,
		},
		Threshold: threshold,
	}
}

func (f *ConditionFactory) CreateVolumeCondition(name, desc, timeWindow, operator string, threshold float64) Condition {
	return &VolumeCondition{
		BaseCondition: BaseCondition{
			Name:        name,
			Description: desc,
			TimeWindow:  timeWindow,
			Operator:    operator,
		},
		Threshold: decimal.NewFromFloat(threshold),
	}
}

func (f *ConditionFactory) CreateUniqueWalletsCondition(name, desc, timeWindow, operator string, threshold int) Condition {
	return &UniqueWalletsCondition{
		BaseCondition: BaseCondition{
			Name:        name,
			Description: desc,
			TimeWindow:  timeWindow,
			Operator:    operator,
		},
		Threshold: threshold,
	}
}

// CreateBigTransactionCondition 创建大额交易条件
func (f *ConditionFactory) CreateBigTransactionCondition(name, desc string, amountThreshold float64, minUsers int, buyToSellRatioMin float64) Condition {
	return &BigTransactionCondition{
		Name:              name,
		Description:       desc,
		AmountThreshold:   decimal.NewFromFloat(amountThreshold),
		MinUsers:          minUsers,
		BuyToSellRatioMin: buyToSellRatioMin,
	}
}

// CreateWhaleTransactionCondition 创建巨鲸交易条件
func (f *ConditionFactory) CreateWhaleTransactionCondition(name, desc string, thresholdUSD float64) Condition {
	return NewWhaleTransactionCondition(name, desc, decimal.NewFromFloat(thresholdUSD))
}
