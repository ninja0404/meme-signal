package condition

import (
	"github.com/shopspring/decimal"
)

// WhaleTransactionCondition 巨鲸交易条件
// 检测单笔交易金额是否超过指定阈值
type WhaleTransactionCondition struct {
	BaseCondition
	AmountThresholdUSD decimal.Decimal // 交易金额阈值（USD）
}

// NewWhaleTransactionCondition 创建巨鲸交易条件
func NewWhaleTransactionCondition(name, description string, thresholdUSD decimal.Decimal) *WhaleTransactionCondition {
	return &WhaleTransactionCondition{
		BaseCondition: BaseCondition{
			Name:        name,
			Description: description,
		},
		AmountThresholdUSD: thresholdUSD,
	}
}

// Evaluate 评估巨鲸交易条件
func (c *WhaleTransactionCondition) Evaluate(ctx *EvaluationContext) bool {
	// 检查当前交易的USD金额
	if ctx.Transaction == nil {
		return false
	}

	// 当前交易的USD金额
	currentTxAmountUSD := ctx.Transaction.AmountUSD

	// 检查是否超过阈值
	return currentTxAmountUSD.GreaterThan(c.AmountThresholdUSD)
}

// GetName 获取条件名称
func (c *WhaleTransactionCondition) GetName() string {
	return c.Name
}

// GetDescription 获取条件描述
func (c *WhaleTransactionCondition) GetDescription() string {
	return c.Description
}
