package condition

import (
	"time"

	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/shopspring/decimal"
)

// Condition 检测条件接口
type Condition interface {
	// Evaluate 评估条件是否满足
	Evaluate(context *EvaluationContext) bool

	// GetName 获取条件名称
	GetName() string

	// GetDescription 获取条件描述
	GetDescription() string
}

// EvaluationContext 评估上下文，包含所有需要的数据
type EvaluationContext struct {
	Stats5m     *model.TokenStats  // 5分钟统计数据
	Stats30s    *model.TokenStats  // 30秒统计数据
	Stats1m     *model.TokenStats  // 1分钟统计数据（可扩展）
	Transaction *model.Transaction // 当前交易
	TokenWindow TokenWindowReader  // 窗口数据读取器
}

// TokenWindowReader 窗口数据读取器接口
type TokenWindowReader interface {
	GetStats() *model.TokenStats
	GetLast30SecondStats() *model.TokenStats
	GetStatsForDuration(duration string) *model.TokenStats // 支持扩展任意时间段
	GetTransactionCount() int
	GetLastUpdate() time.Time
	GetLast30SecondBigTransactionStats(threshold decimal.Decimal) map[string]interface{} // 新增大额交易统计方法
	GetMaxSingleTransactionAmount() decimal.Decimal                                      // 获取5分钟内最大单笔交易金额
}

// LogicalOperator 逻辑操作符
type LogicalOperator string

const (
	AND LogicalOperator = "AND"
	OR  LogicalOperator = "OR"
	NOT LogicalOperator = "NOT"
)

// CompositeCondition 复合条件，支持AND/OR/NOT逻辑组合
type CompositeCondition struct {
	Name        string
	Description string
	Operator    LogicalOperator
	Conditions  []Condition
}

func (c *CompositeCondition) Evaluate(context *EvaluationContext) bool {
	switch c.Operator {
	case AND:
		for _, condition := range c.Conditions {
			if !condition.Evaluate(context) {
				return false
			}
		}
		return len(c.Conditions) > 0

	case OR:
		for _, condition := range c.Conditions {
			if condition.Evaluate(context) {
				return true
			}
		}
		return false

	case NOT:
		if len(c.Conditions) != 1 {
			return false // NOT操作符只能有一个条件
		}
		return !c.Conditions[0].Evaluate(context)

	default:
		return false
	}
}

func (c *CompositeCondition) GetName() string {
	return c.Name
}

func (c *CompositeCondition) GetDescription() string {
	return c.Description
}

// Builder 条件建造者，支持链式调用
type Builder struct {
	conditions []Condition
	operator   LogicalOperator
	name       string
	desc       string
}

func NewBuilder() *Builder {
	return &Builder{
		conditions: make([]Condition, 0),
		operator:   AND, // 默认AND操作
	}
}

func (b *Builder) Name(name string) *Builder {
	b.name = name
	return b
}

func (b *Builder) Description(desc string) *Builder {
	b.desc = desc
	return b
}

func (b *Builder) And(condition Condition) *Builder {
	b.operator = AND
	b.conditions = append(b.conditions, condition)
	return b
}

func (b *Builder) Or(condition Condition) *Builder {
	b.operator = OR
	b.conditions = append(b.conditions, condition)
	return b
}

func (b *Builder) Not(condition Condition) *Builder {
	b.operator = NOT
	b.conditions = []Condition{condition} // NOT只能有一个条件
	return b
}

func (b *Builder) Build() Condition {
	if len(b.conditions) == 1 && b.operator == AND {
		// 如果只有一个条件，直接返回该条件
		return b.conditions[0]
	}

	return &CompositeCondition{
		Name:        b.name,
		Description: b.desc,
		Operator:    b.operator,
		Conditions:  b.conditions,
	}
}

// WhaleTransaction 创建巨鲸交易条件的便捷方法
func (b *Builder) WhaleTransaction(name, description string, quietVolumeMax, quietMaxSingle, suddenThreshold float64, quietMaxTxCount int) *Builder {
	condition := NewWhaleTransactionCondition(name, description, quietVolumeMax, quietMaxSingle, suddenThreshold, quietMaxTxCount)
	b.conditions = append(b.conditions, condition)
	return b
}
