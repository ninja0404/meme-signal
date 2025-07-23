package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const TableNameTokensInfo = "tokens_info"

// TokensInfo 代币信息表
type TokensInfo struct {
	ID              int64           `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Name            string          `gorm:"column:name;not null;comment:代币名称" json:"name"`                                                // 代币名称
	Symbol          string          `gorm:"column:symbol;not null;comment:代币符号" json:"symbol"`                                            // 代币符号
	TokenAddress    string          `gorm:"column:token_address;not null;comment:代币地址" json:"token_address"`                              // 代币地址
	Decimals        int32           `gorm:"column:decimals;not null;comment:代币精度" json:"decimals"`                                        // 代币精度
	URI             string          `gorm:"column:uri;not null;comment:代币URI" json:"uri"`                                                 // 代币URI
	OssURI          string          `gorm:"column:oss_uri;not null;comment:oss存储的代币URI" json:"oss_uri"`                                   // oss存储的代币URI
	Metadata        string          `gorm:"column:metadata;not null;comment:代币元数据" json:"metadata"`                                       // 代币元数据
	CreateSignature string          `gorm:"column:create_signature;not null;comment:创建代币交易hash" json:"create_signature"`                  // 创建代币交易hash
	Supply          decimal.Decimal `gorm:"column:supply;not null;default:0.000000000000000000;comment:代币总量" json:"supply"`               // 代币总量
	MarketCap       decimal.Decimal `gorm:"column:market_cap;not null;default:0.000000000000000000;comment:代币市值" json:"market_cap"`       // 代币市值
	Creator         string          `gorm:"column:creator;not null;comment:创建者地址" json:"creator"`                                         // 创建者地址
	CurrentPrice    decimal.Decimal `gorm:"column:current_price;not null;default:0.000000000000000000;comment:当前价格" json:"current_price"` // 当前价格
	BlockTime       time.Time       `gorm:"column:block_time;not null" json:"block_time"`
	CreatedAt       *time.Time      `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt       *time.Time      `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
	Fee             decimal.Decimal `gorm:"column:fee;not null;default:0.000000;comment:代币的手续费" json:"fee"`                 // 代币的手续费
	IsVerified      bool            `gorm:"column:is_verified;not null;comment:平台是否背书" json:"is_verified"`                  // 平台是否背书
	IsMutable       bool            `gorm:"column:is_mutable;not null;comment:meta信息是否可修改" json:"is_mutable"`               // meta信息是否可修改
	MetadataAddress string          `gorm:"column:metadata_address;not null;comment:代币meta信息的链上地址" json:"metadata_address"` // 代币meta信息的链上地址
	PriceUpdateTime *time.Time      `gorm:"column:price_update_time;comment:价格更新时间" json:"price_update_time"`               // 价格更新时间
	JsonURI         string          `gorm:"column:json_uri;not null;comment:代币信息json链接" json:"json_uri"`                    // 代币信息json链接
}

// TableName TokensInfo's table name
func (*TokensInfo) TableName() string {
	return TableNameTokensInfo
}
