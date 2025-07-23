package fragment

import (
	"github.com/gagliardetto/solana-go"
	"github.com/shopspring/decimal"
)

type JupiterFragment struct {
	UserWallet  solana.PublicKey // 从指令里解析到的真实用户
	ProxyWallet solana.PublicKey // 代理钱包
	FeeWallet   solana.PublicKey // 手续费钱包
	InstIdxInTx string

	// swap 层面
	ProgramAddress   solana.PublicKey // dex的程序地址
	TypeId           uint64
	TokenInAddress   solana.PublicKey
	TokenInDecimals  int32
	TokenInAmount    decimal.Decimal
	TokenOutAddress  solana.PublicKey
	TokenOutDecimals int32
	TokenOutAmount   decimal.Decimal
	Dex              string
}

func (*JupiterFragment) Type() FragmentType {
	return JupiterType
}

func (t *JupiterFragment) IsSetInstNumber() bool {
	return t.InstIdxInTx != ""
}

func (t *JupiterFragment) SetInstNumber(instIdxInTx string) {
	t.InstIdxInTx = instIdxInTx
}
