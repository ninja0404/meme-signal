package fragment

import (
	"github.com/gagliardetto/solana-go"
	"github.com/shopspring/decimal"
)

type LiquidityDirection int32

const (
	AddLiquidity    LiquidityDirection = 1
	RemoveLiquidity LiquidityDirection = 2
	InitPool        LiquidityDirection = 3
)

type PoolFragment struct {
	InstIdxInTx string

	MarketAddress      solana.PublicKey
	UserWallet         solana.PublicKey
	LiquidityDirection LiquidityDirection // 流动性改变方向，加池子和减池子
	TokenAddress       solana.PublicKey
	MainAddress        solana.PublicKey

	TokenXMint     solana.PublicKey // 代币对地址X
	TokenXAmount   decimal.Decimal
	TokenXDecimals int32

	TokenYMint     solana.PublicKey // 代币对地址Y
	TokenYAmount   decimal.Decimal
	TokenYDecimals int32

	PoolXAddress solana.PublicKey // 池子代币X下的账户地址
	PoolYAddress solana.PublicKey // 池子代币Y下的账户地址
	LpMintToken  solana.PublicKey
	Dex          string
}

func (*PoolFragment) Type() FragmentType {
	return PoolType
}

func (i *PoolFragment) IsSetInstNumber() bool {
	return i.InstIdxInTx != ""
}

func (i *PoolFragment) SetInstNumber(instIdxInTx string) {
	i.InstIdxInTx = instIdxInTx
}
