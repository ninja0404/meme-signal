package fragment

import (
	"github.com/gagliardetto/solana-go"
	"github.com/shopspring/decimal"
)

type TradeFragment struct {
	// tx 层面
	//Signature  string
	//Slot       uint64
	//Fee        uint64
	//BlockTime  time.Time
	UserWallet  solana.PublicKey // 从指令里解析到的真实用户
	InstIdxInTx string

	// swap 层面
	ProgramAddress   solana.PublicKey
	TypeId           uint64
	MarketAddress    solana.PublicKey
	TokenInAddress   solana.PublicKey
	TokenInDecimals  int32
	TokenInAmount    decimal.Decimal
	TokenOutAddress  solana.PublicKey
	TokenOutDecimals int32
	TokenOutAmount   decimal.Decimal
	Dex              string

	PoolIn  solana.PublicKey
	PoolOut solana.PublicKey

	PoolInLPMint   solana.PublicKey
	PoolOutLPMint  solana.PublicKey
	PoolInLPVault  solana.PublicKey
	PoolOutLPVault solana.PublicKey
}

func (*TradeFragment) Type() FragmentType {
	return TradeType
}

func (t *TradeFragment) IsSetInstNumber() bool {
	return t.InstIdxInTx != ""
}

func (t *TradeFragment) SetInstNumber(instIdxInTx string) {
	t.InstIdxInTx = instIdxInTx
}
