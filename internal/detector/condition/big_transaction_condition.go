package condition

import (
	"github.com/shopspring/decimal"
)

// BigTransactionCondition 大额交易条件
type BigTransactionCondition struct {
	Name              string
	Description       string
	AmountThreshold   decimal.Decimal // 大额交易金额阈值（USD）
	MinUsers          int             // 最少用户数
	BuyToSellRatioMin float64         // 买卖比例最小值（买/卖）
}

func (c *BigTransactionCondition) Evaluate(context *EvaluationContext) bool {
	if context.TokenWindow == nil {
		return false
	}

	// 获取最后30秒的大额交易统计
	stats := context.TokenWindow.GetLast30SecondBigTransactionStats(c.AmountThreshold)

	bigTxUsers := stats["big_tx_users"].(int)
	bigTxBuyCount := stats["big_tx_buy_count"].(int)
	bigTxSellCount := stats["big_tx_sell_count"].(int)

	// 检查用户数是否满足要求
	if bigTxUsers < c.MinUsers {
		return false
	}

	// 检查买卖比例
	// 如果没有卖单，只要有买单就满足条件
	if bigTxSellCount == 0 {
		return bigTxBuyCount > 0
	}

	// 计算买卖比例
	buyToSellRatio := float64(bigTxBuyCount) / float64(bigTxSellCount)

	return buyToSellRatio >= c.BuyToSellRatioMin
}

func (c *BigTransactionCondition) GetName() string {
	return c.Name
}

func (c *BigTransactionCondition) GetDescription() string {
	return c.Description
}
