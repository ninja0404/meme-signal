package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const TableNameBiTokenHolder = "bi_token_holders"

// BiTokenHolder 用户代币持仓表
type BiTokenHolder struct {
	ID                       int64           `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	TokenAddress             string          `gorm:"column:token_address;not null;comment:代币地址" json:"token_address"`                                           // 代币地址
	WalletAddress            string          `gorm:"column:wallet_address;not null;comment:钱包地址" json:"wallet_address"`                                         // 钱包地址
	Amount                   decimal.Decimal `gorm:"column:amount;not null;comment:链上代币数量(已处理精度)" json:"amount"`                                                // 链上代币数量(已处理精度)
	TotalBuyAmount           decimal.Decimal `gorm:"column:total_buy_amount;not null;comment:累计买入数量(已处理精度)" json:"total_buy_amount"`                            // 累计买入数量(已处理精度)
	TotalSellAmount          decimal.Decimal `gorm:"column:total_sell_amount;not null;comment:累计卖出数量(已处理精度)" json:"total_sell_amount"`                          // 累计卖出数量(已处理精度)
	TotalDepositAmount       decimal.Decimal `gorm:"column:total_deposit_amount;not null;comment:累计充值数量(已处理精度)" json:"total_deposit_amount"`                    // 累计充值数量(已处理精度)
	TotalWithdrawAmount      decimal.Decimal `gorm:"column:total_withdraw_amount;not null;comment:累计提现数量(已处理精度)" json:"total_withdraw_amount"`                  // 累计提现数量(已处理精度)
	TotalBuyUsd              decimal.Decimal `gorm:"column:total_buy_usd;not null;comment:累计买入花费金额" json:"total_buy_usd"`                                       // 累计买入花费金额
	TotalSellUsd             decimal.Decimal `gorm:"column:total_sell_usd;not null;comment:累计卖出收益金额" json:"total_sell_usd"`                                     // 累计卖出收益金额
	TotalPnlAmount           decimal.Decimal `gorm:"column:total_pnl_amount;not null;comment:累计已结算PNL的数量(已处理精度)" json:"total_pnl_amount"`                       // 累计已结算PNL的数量(已处理精度)
	TotalBuyInPnlUsd         decimal.Decimal `gorm:"column:total_buy_in_pnl_usd;not null;comment:累计已结算买单花费金额" json:"total_buy_in_pnl_usd"`                      // 累计已结算买单花费金额
	TotalBuyInRealizedPnlUsd decimal.Decimal `gorm:"column:total_buy_in_realized_pnl_usd;not null;comment:累计已实现盈亏的买单成本金额" json:"total_buy_in_realized_pnl_usd"` // 累计已实现盈亏的买单成本金额
	TotalRealizedPnlAmount   decimal.Decimal `gorm:"column:total_realized_pnl_amount;not null;comment:累计已实现盈亏的代币数量(已处理精度)" json:"total_realized_pnl_amount"`    // 累计已实现盈亏的代币数量(已处理精度)
	TotalRealizedPnlUsd      decimal.Decimal `gorm:"column:total_realized_pnl_usd;not null;comment:累计已盈亏金额" json:"total_realized_pnl_usd"`                      // 累计已盈亏金额
	CreatedAt                *time.Time      `gorm:"column:created_at;not null;autoCreateTime;comment:创建时间" json:"created_at"`                                  // 创建时间
	UpdatedAt                *time.Time      `gorm:"column:updated_at;not null;autoUpdateTime;comment:更新时间" json:"updated_at"`                                  // 更新时间
}

// TableName BiTokenHolder's table name
func (*BiTokenHolder) TableName() string {
	return TableNameBiTokenHolder
}
