package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type SwapTx struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement;comment:主键"`
	TxHash       string `gorm:"column:tx_hash;type:varchar(128);uniqueIndex:idx_tx_inst_seq;not null;default:'';comment:交易hash"`
	BlockID      uint64 `gorm:"column:block_id;not null;comment:区块 Slot"`
	IndexInBlock int32  `gorm:"column:index_in_block;not null;default:0;comment:交易在区块中的索引"`
	InstIdxInTx  string `gorm:"column:inst_idx_in_tx;uniqueIndex:idx_tx_inst_seq;not null;default:'';comment:指令在交易中的索引位置"`
	SequenceID   int32  `gorm:"column:sequence_id;uniqueIndex:idx_tx_inst_seq;not null;default:0;comment:相同TxHash和InstIdxInTx的记录序号"`

	BlockTime   time.Time `gorm:"column:block_time;not null;comment:链上时间"`
	UserWallet  string    `gorm:"column:user_wallet;type:varchar(128);not null;comment:用户钱包地址"`
	UserWallet2 string    `gorm:"column:user_wallet2;type:varchar(128);not null;comment:用户钱包地址2"`

	TokenAddress     string `gorm:"column:token_address;type:varchar(128);not null;comment:token 地址"`
	MainAddress      string `gorm:"column:main_address;type:varchar(128);not null;comment:主代币地址"`
	MarketAddress    string
	TokenInAddress   string `gorm:"column:token_in_address;type:varchar(128);not null;comment:Token in地址"`
	TokenOutAddress  string `gorm:"column:token_out_address;type:varchar(128);not null;comment:Token out地址"`
	TokenInDecimals  int32  `gorm:"column:token_in_decimals;not null;comment:TokenIn 精度"`
	TokenOutDecimals int32  `gorm:"column:token_out_decimals;not null;comment:TokenOut 精度"`

	TokenInAmount  decimal.Decimal `gorm:"column:token_in_amount;type:decimal(65,0);not null;comment:TokenIn 数量"`
	TokenOutAmount decimal.Decimal `gorm:"column:token_out_amount;type:decimal(65,0);not null;comment:TokenOut 数量"`
	AmountUSD      decimal.Decimal `gorm:"column:amount_usd;type:decimal(32,18);default:0;comment:按 USD 计价的总token价值"`
	PriceUSD       decimal.Decimal `gorm:"column:price_usd;type:decimal(32,18);default:0;comment:代币单价"`
	Fee            decimal.Decimal `gorm:"column:fee;type:decimal(65,0);not null;comment:手续费"`

	Action int32  `gorm:"column:action;not null;comment:0=未知, 1=买Token, 2=卖Token, 3-代币转账"`
	Dex    string `gorm:"column:dex;type:varchar(32);default:'';comment:聚合后主 DEX Program (可为空)"`

	IsLossTx  int32
	IsBundled int32

	CreatedAt *time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt *time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (*SwapTx) TableName() string {
	return "swap_tx"
}
