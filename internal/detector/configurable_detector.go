package detector

import (
	"fmt"
	"time"

	"github.com/ninja0404/meme-signal/internal/detector/condition"
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/pkg/logger"
	"github.com/shopspring/decimal"
)

// ConfigurableDetector 可配置的检测器
type ConfigurableDetector struct {
	Name        string
	Description string
	Type        string
	Condition   condition.Condition
	SignalType  model.SignalType
	Severity    int
	Confidence  float64
}

func (d *ConfigurableDetector) GetType() string {
	return d.Type
}

func (d *ConfigurableDetector) Detect(stats *model.TokenStats, tx *model.Transaction, window *TokenWindow) []*model.Signal {
	// 准备评估上下文
	context := &condition.EvaluationContext{
		Stats5m:     stats,
		Stats30s:    window.GetLast30SecondStats(),
		Transaction: tx,
		TokenWindow: &TokenWindowAdapter{window: window},
	}

	// 评估条件
	conditionMet := d.Condition.Evaluate(context)

	if conditionMet {
		signal := &model.Signal{
			ID:           generateSignalID(),
			Type:         d.SignalType,
			TokenAddress: stats.Address,
			TokenSymbol:  "",
			Severity:     d.Severity,
			Confidence:   d.Confidence,
			Message:      fmt.Sprintf("检测器[%s]触发: %s", d.Name, d.Description),
			Data: map[string]interface{}{
				"detector_name":   d.Name,
				"detector_type":   d.Type,
				"condition_name":  d.Condition.GetName(),
				"condition_desc":  d.Condition.GetDescription(),
				"trigger_time":    time.Now().Format("2006-01-02 15:04:05"),
				"trigger_tx_sig":  tx.Signature,
				"current_price":   stats.CurrentPrice.Truncate(8).String(),
				"price_change_5m": stats.PriceChangePercent.Truncate(2).String(),
				"unique_wallets":  stats.UniqueHolders,
				"tx_count_5m":     stats.TxCount5m,
				"volume_5m":       stats.Volume5m.Truncate(2).String(),
			},
			Timestamp: time.Now(),
			SourceTx:  tx,
		}

		logger.Info("🚨 检测器条件满足",
			logger.String("detector", d.Name),
			logger.String("condition", d.Condition.GetName()),
			logger.String("token", stats.Address),
			logger.String("message", signal.Message))

		return []*model.Signal{signal}
	}

	return nil
}

// TokenWindowAdapter 适配器，让TokenWindow实现TokenWindowReader接口
type TokenWindowAdapter struct {
	window *TokenWindow
}

func (a *TokenWindowAdapter) GetStats() *model.TokenStats {
	return a.window.GetStats()
}

func (a *TokenWindowAdapter) GetLast30SecondStats() *model.TokenStats {
	return a.window.GetLast30SecondStats()
}

func (a *TokenWindowAdapter) GetStatsForDuration(duration string) *model.TokenStats {
	// 可以扩展支持任意时间段的统计
	switch duration {
	case "5m":
		return a.window.GetStats()
	case "30s":
		return a.window.GetLast30SecondStats()
	default:
		return nil
	}
}

func (a *TokenWindowAdapter) GetTransactionCount() int {
	return a.window.GetTransactionCount()
}

func (a *TokenWindowAdapter) GetLastUpdate() time.Time {
	return a.window.GetLastUpdate()
}

// GetLast30SecondBigTransactionStats 获取最后30秒大额交易统计
func (a *TokenWindowAdapter) GetLast30SecondBigTransactionStats(threshold decimal.Decimal) map[string]interface{} {
	return a.window.GetLast30SecondBigTransactionStats(threshold)
}

// GetMaxSingleTransactionAmount 获取5分钟内最大单笔交易金额
func (a *TokenWindowAdapter) GetMaxSingleTransactionAmount() decimal.Decimal {
	return a.window.GetMaxSingleTransactionAmount()
}

// DetectorBuilder 检测器建造者
type DetectorBuilder struct {
	name         string
	description  string
	detectorType string
	signalType   model.SignalType
	severity     int
	confidence   float64
	condition    condition.Condition
}

func NewDetectorBuilder() *DetectorBuilder {
	return &DetectorBuilder{
		severity:   5,
		confidence: 0.8,
		signalType: model.SignalTypeLargeTransaction,
	}
}

func (b *DetectorBuilder) Name(name string) *DetectorBuilder {
	b.name = name
	return b
}

func (b *DetectorBuilder) Description(desc string) *DetectorBuilder {
	b.description = desc
	return b
}

func (b *DetectorBuilder) Type(detectorType string) *DetectorBuilder {
	b.detectorType = detectorType
	return b
}

func (b *DetectorBuilder) SignalType(signalType model.SignalType) *DetectorBuilder {
	b.signalType = signalType
	return b
}

func (b *DetectorBuilder) Severity(severity int) *DetectorBuilder {
	b.severity = severity
	return b
}

func (b *DetectorBuilder) Confidence(confidence float64) *DetectorBuilder {
	b.confidence = confidence
	return b
}

func (b *DetectorBuilder) WithCondition(cond condition.Condition) *DetectorBuilder {
	b.condition = cond
	return b
}

func (b *DetectorBuilder) Build() Detector {
	return &ConfigurableDetector{
		Name:        b.name,
		Description: b.description,
		Type:        b.detectorType,
		Condition:   b.condition,
		SignalType:  b.signalType,
		Severity:    b.severity,
		Confidence:  b.confidence,
	}
}

// DetectorRegistry 检测器注册表
type DetectorRegistry struct {
	detectors map[string]func() Detector
	factory   *condition.ConditionFactory
}

func NewDetectorRegistry() *DetectorRegistry {
	return &DetectorRegistry{
		detectors: make(map[string]func() Detector),
		factory:   condition.NewConditionFactory(),
	}
}

func (r *DetectorRegistry) Register(name string, factory func() Detector) {
	r.detectors[name] = factory
}

func (r *DetectorRegistry) Create(name string) (Detector, error) {
	factory, exists := r.detectors[name]
	if !exists {
		return nil, fmt.Errorf("检测器 '%s' 未注册", name)
	}
	return factory(), nil
}

func (r *DetectorRegistry) CreateMemeSignalDetector() Detector {
	// 使用条件建造者创建当前的meme信号检测逻辑
	priceChange5m := r.factory.CreatePriceChangeCondition(
		"price_change_5m_20pct",
		"5分钟内涨幅≥20%",
		"5m",
		">=",
		25.0,
	)

	priceChange30s := r.factory.CreatePriceChangeCondition(
		"price_change_30s_15pct",
		"30秒内涨幅≥15%",
		"30s",
		">=",
		15.0,
	)

	txCount := r.factory.CreateTransactionCountCondition(
		"tx_count_5m_300plus",
		"5分钟内交易次数>300笔",
		"5m",
		">",
		300,
	)

	// 新增条件：独立钱包数量超过50个
	uniqueWallets := r.factory.CreateUniqueWalletsCondition(
		"unique_wallets_5m_50plus",
		"5分钟内独立钱包数量>100个",
		"5m",
		">",
		100,
	)

	// 新增条件：大额交易条件
	bigTransactionCondition := r.factory.CreateBigTransactionCondition(
		"big_tx_30s_analysis",
		"30秒内大额交易(>1000U): 用户数≥5个, 买卖比≥2:1",
		600.0, // 大额交易阈值：1000 USD
		8,     // 最少用户数：8个
		1.5,   // 买卖比例：买单数量 ≥ 卖单数量 * 2
	)

	// 组合条件：所有条件都必须满足 (AND)
	combinedCondition := condition.NewBuilder().
		Name("meme_signal_composite").
		Description("Meme代币信号检测组合条件").
		And(priceChange5m).
		And(priceChange30s).
		And(txCount).
		And(uniqueWallets).
		And(bigTransactionCondition).
		Build()

	return NewDetectorBuilder().
		Name("meme_signal_detector").
		Description("Meme代币信号检测器").
		Type("meme_signal").
		SignalType(model.SignalTypeCompositeSignal).
		Severity(8).
		Confidence(0.95).
		WithCondition(combinedCondition).
		Build()
}

// GetRegisteredDetectors 获取已注册的检测器列表
func (r *DetectorRegistry) GetRegisteredDetectors() []string {
	names := make([]string, 0, len(r.detectors))
	for name := range r.detectors {
		names = append(names, name)
	}
	return names
}
