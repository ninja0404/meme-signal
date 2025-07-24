package condition

import (
	"github.com/shopspring/decimal"
)

// WhaleTransactionCondition 突然性巨鲸活动条件
// 检测：平静期 + 突然出现大额交易
type WhaleTransactionCondition struct {
	BaseCondition
	QuietPeriodVolumeMax   decimal.Decimal // 平静期最大交易量阈值 ($40,000)
	QuietPeriodMaxSingleTx decimal.Decimal // 平静期最大单笔交易阈值 ($5,000)
	QuietPeriodMaxTxCount  int             // 平静期最大交易数阈值 (80笔)
	SuddenTxThreshold      decimal.Decimal // 突然大额交易阈值 ($10,000)
}

// NewWhaleTransactionCondition 创建突然性巨鲸活动条件
func NewWhaleTransactionCondition(name, description string, quietVolumeMax, quietMaxSingle, suddenThreshold float64, quietMaxTxCount int) *WhaleTransactionCondition {
	return &WhaleTransactionCondition{
		BaseCondition: BaseCondition{
			Name:        name,
			Description: description,
		},
		QuietPeriodVolumeMax:   decimal.NewFromFloat(quietVolumeMax),
		QuietPeriodMaxSingleTx: decimal.NewFromFloat(quietMaxSingle),
		QuietPeriodMaxTxCount:  quietMaxTxCount,
		SuddenTxThreshold:      decimal.NewFromFloat(suddenThreshold),
	}
}

// Evaluate 评估突然性巨鲸活动条件
func (c *WhaleTransactionCondition) Evaluate(ctx *EvaluationContext) bool {
	// 检查当前交易是否存在
	if ctx.Transaction == nil {
		return false
	}

	// 检查5分钟统计数据是否存在
	if ctx.Stats5m == nil {
		return false
	}

	// 1. 检查平静期条件：5分钟总交易量 < $40,000
	if ctx.Stats5m.Volume5m.GreaterThanOrEqual(c.QuietPeriodVolumeMax) {
		return false
	}

	// 2. 检查平静期条件：5分钟内交易数 < 80笔
	if ctx.Stats5m.TxCount5m >= c.QuietPeriodMaxTxCount {
		return false
	}

	// 3. 检查平静期条件：5分钟内最大单笔交易 < $5,000
	if ctx.TokenWindow != nil {
		maxSingleTx := ctx.TokenWindow.GetMaxSingleTransactionAmount()
		if maxSingleTx.GreaterThanOrEqual(c.QuietPeriodMaxSingleTx) {
			return false
		}
	}

	// 4. 检查突破条件：当前交易 > $10,000
	if ctx.Transaction.AmountUSD.LessThan(c.SuddenTxThreshold) {
		return false
	}

	// 所有条件都满足，触发突然性巨鲸活动信号
	return true
}

// GetName 获取条件名称
func (c *WhaleTransactionCondition) GetName() string {
	return c.Name
}

// GetDescription 获取条件描述
func (c *WhaleTransactionCondition) GetDescription() string {
	return c.Description
}
