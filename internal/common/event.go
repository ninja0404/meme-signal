package common

import (
	"fmt"
	"time"

	"github.com/ninja0404/meme-signal/internal/common/fragment"
	"github.com/shopspring/decimal"
)

type EventType int32

const (
	TradeEventType EventType = iota + 1
	TransferEventType
	PoolEventType
	TokenEventType
)

type Action int32

const (
	UnknownAction         Action = 0
	BuyAction             Action = 1
	SellAction            Action = 2
	TransferAction        Action = 3
	DepositWithdrawAction Action = 6
	ArbitrageAction       Action = 7
)

func (a Action) Enum() int32 {
	return int32(a)
}

func (e EventType) Enum() int32 {
	return int32(e)
}

type Event struct {
	Type       EventType  `json:"type"`
	InnerEvent InnerEvent `json:"inner_event"`
}

type InnerEvent interface {
	GetKey() string
	GetInstIdxInTx() string
}

type TradeEvent struct {
	ID           uint64    `json:"id"`
	Signature    string    `json:"signature"`
	Fee          uint64    `json:"fee"`
	Slot         uint64    `json:"slot"`
	BlockTime    time.Time `json:"block_time"`
	TxIdxInBlock int32     `json:"tx_idx_in_block"`
	InstIdxInTx  string    `json:"inst_idx_in_tx"`

	UserWallet    string   `json:"user_wallet"`
	TokenAddress  string   `json:"token_address"`
	MainAddress   string   `json:"main_address"`
	MarketAddress []string `json:"market_address"`

	TokenInAddress   string `json:"token_in_address"`
	TokenOutAddress  string `json:"token_out_address"`
	TokenInDecimals  int32  `json:"token_in_decimals"`
	TokenOutDecimals int32  `json:"token_out_decimals"`

	TokenInAmount  decimal.Decimal `json:"token_in_amount"`
	TokenOutAmount decimal.Decimal `json:"token_out_amount"`

	AmountUSD decimal.Decimal `json:"amount_usd"`
	PriceUSD  decimal.Decimal `json:"price_usd"`
	Action    Action          `json:"action"`
	Dex       string          `json:"dex"`

	PoolInAcc  string `json:"pool_in_acc"`
	PoolOutAcc string `json:"pool_out_acc"`

	// 带上token change信息
	TokenBalanceChange *TokenBalanceChange `json:"token_balance_change"`

	PoolInLPVault  string `json:"pool_in_lp_vault"`
	PoolOutLPVault string `json:"pool_out_lp_vault"`
}

type TokenBalanceChange struct {
	TokenAddress     string          `json:"token_address"`
	TokenDecimals    int32           `json:"token_decimals"`
	ChangeType       int32           `json:"change_type"` // 1-sol余额变动; 2-token余额变动
	PreTokenBalance  decimal.Decimal `json:"pre_token_balance"`
	PostTokenBalance decimal.Decimal `json:"post_token_balance"`
}

func (t *TradeEvent) GetInstIdxInTx() string {
	return t.InstIdxInTx
}

func (t *TradeEvent) GetKey() string {
	return fmt.Sprintf("%s-%s", t.Signature, t.InstIdxInTx)
}

type TransferEvent struct {
	Signature    string    `json:"signature"`
	Fee          uint64    `json:"fee"`
	Slot         uint64    `json:"slot"`
	BlockTime    time.Time `json:"block_time"`
	TxIdxInBlock int32     `json:"tx_idx_in_block"`
	InstIdxInTx  string    `json:"inst_idx_in_tx"`

	UserWallet     string          `json:"user_wallet"`
	UserWallet2    string          `json:"user_wallet2"`
	TokenAddress   string          `json:"token_address"`
	Decimals       int32           `json:"decimals"`
	TokenInAmount  decimal.Decimal `json:"token_in_amount"`
	TokenOutAmount decimal.Decimal `json:"token_out_amount"`

	Action Action `json:"action"`
}

func (t *TransferEvent) GetInstIdxInTx() string {
	return t.InstIdxInTx
}

func (t *TransferEvent) GetKey() string {
	return fmt.Sprintf("%s-%s", t.Signature, t.InstIdxInTx)
}

type PoolEvent struct {
	Signature    string    `json:"signature"`
	Slot         uint64    `json:"slot"`
	BlockTime    time.Time `json:"block_time"`
	TxIdxInBlock int32     `json:"tx_idx_in_block"`
	InstIdxInTx  string    `json:"inst_idx_in_tx"`

	EventType     fragment.LiquidityDirection `json:"event_type"`
	UserWallet    string                      `json:"user_wallet"`
	MarketAddress string                      `json:"market_address"`
	TokenAddress  string                      `json:"token_address"`
	MainAddress   string                      `json:"main_address"`

	TokenXAddress  string          `json:"token_x_address"`
	TokenYAddress  string          `json:"token_y_address"`
	PoolX          string          `json:"pool_x"`
	PoolY          string          `json:"pool_y"`
	TokenXAmount   decimal.Decimal `json:"token_x_amount"`
	TokenYAmount   decimal.Decimal `json:"token_y_amount"`
	TokenXDecimals int32           `json:"token_x_decimals"`
	TokenYDecimals int32           `json:"token_y_decimals"`

	LpMintToken string `json:"lp_mint_token"`
	Dex         string `json:"dex"`
}

func (t *PoolEvent) GetInstIdxInTx() string {
	return t.InstIdxInTx
}

func (t *PoolEvent) GetKey() string {
	return fmt.Sprintf("%s-%s", t.Signature, t.InstIdxInTx)
}

type TokenEvent struct {
	// block info
	Signature    string    `json:"signature"`
	Slot         uint64    `json:"slot"`
	BlockTime    time.Time `json:"block_time"`
	TxIdxInBlock int32     `json:"tx_idx_in_block"`
	InstIdxInTx  string    `json:"inst_idx_in_tx"`

	// token info
	TokenAddress    string          `json:"token_address"`
	MetadataAddress string          `json:"metadata_address"`
	Name            string          `json:"name"`
	Symbol          string          `json:"symbol"`
	Decimals        int32           `json:"decimals"`
	LogoURI         string          `json:"logo_uri"`
	FeeRate         decimal.Decimal `json:"fee_rate"`
	Creator         string          `json:"creator"`
	Supply          decimal.Decimal `json:"supply"`
	IsMutable       *bool           `json:"is_mutable"`
	Option          int32           `json:"option"` // 预留-0，新增-1，更新mint账户-2，更新meta账户-3
}

func (t *TokenEvent) GetInstIdxInTx() string {
	return t.InstIdxInTx
}

func (t *TokenEvent) GetKey() string {
	return fmt.Sprintf("%s-%s", t.Signature, t.InstIdxInTx)
}
