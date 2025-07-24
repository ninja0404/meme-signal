package model

import (
	"time"

	"github.com/ninja0404/meme-signal/internal/common"
	"github.com/shopspring/decimal"
)

// DEX名称

// Transaction DEX交易数据
type Transaction struct {
	ID           string          `json:"id"`
	Signature    string          `json:"signature"`
	Slot         uint64          `json:"slot"`
	BlockTime    time.Time       `json:"block_time"`
	UserWallet   string          `json:"user_wallet"`
	TokenAddress string          `json:"token_address"`
	AmountUSD    decimal.Decimal `json:"amount_usd"` // USD价值
	PriceUSD     decimal.Decimal `json:"price_usd"`  // 代币价格
	Action       common.Action   `json:"action"`
}

// TokenInfo 代币信息
type TokenInfo struct {
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
	Symbol   string `json:"symbol,omitempty"`
	Name     string `json:"name,omitempty"`
}

// SignalType 信号类型
type SignalType string

const (
	SignalTypePriceSpike       SignalType = "price_spike"       // 价格异动
	SignalTypeVolumeSpike      SignalType = "volume_spike"      // 交易量突增
	SignalTypeLargeTransaction SignalType = "large_transaction" // 大额交易
	SignalTypeNewToken         SignalType = "new_token"         // 新代币上线
	SignalTypeWhaleActivity    SignalType = "whale_activity"    // 巨鲸活动
	SignalTypeCompositeSignal  SignalType = "composite_signal"  // 混合信号
)

// Signal 检测到的信号
type Signal struct {
	ID           string                 `json:"id"`
	Type         SignalType             `json:"type"`
	TokenAddress string                 `json:"token_address"`
	TokenSymbol  string                 `json:"token_symbol,omitempty"`
	Severity     int                    `json:"severity"`   // 1-10 信号强度
	Confidence   float64                `json:"confidence"` // 0-1 置信度
	Message      string                 `json:"message"`    // 信号描述
	Data         map[string]interface{} `json:"data"`       // 额外数据
	Timestamp    time.Time              `json:"timestamp"`
	SourceTx     *Transaction           `json:"source_tx,omitempty"` // 触发信号的交易
}

// TokenStats 代币统计信息
type TokenStats struct {
	Address            string                 `json:"address"`
	Volume5m           decimal.Decimal        `json:"volume_5m"`
	VolumeChange       decimal.Decimal        `json:"volume_change"`  // 5m交易量变化百分比
	PriceChange        decimal.Decimal        `json:"price_change"`   // 5m价格变化百分比
	TxCount5m          int                    `json:"tx_count_5m"`    // 5m交易次数
	UniqueHolders      int                    `json:"unique_holders"` // 5m独立持有者数量
	LastUpdate         time.Time              `json:"last_update"`
	StartPrice         decimal.Decimal        `json:"start_price"`          // 5分钟窗口开始价格
	CurrentPrice       decimal.Decimal        `json:"current_price"`        // 当前价格
	PriceChangePercent decimal.Decimal        `json:"price_change_percent"` // 5分钟内价格变化百分比
	Data               map[string]interface{} `json:"data,omitempty"`       // 额外的统计数据（如大额交易统计）
}

// MarketEvent 市场事件
type MarketEvent struct {
	Type      string                 `json:"type"`
	Token     string                 `json:"token"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}
