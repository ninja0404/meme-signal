package fragment

var _ IFragment = (*TradeFragment)(nil)
var _ IFragment = (*PoolFragment)(nil)
var _ IFragment = (*TransferFragment)(nil)
var _ IFragment = (*JupiterFragment)(nil)
var _ IFragment = (*TokenFragment)(nil)

type FragmentType int32

const (
	TradeType FragmentType = iota + 1
	PoolType
	TransferType
	JupiterType
	TokenType
)

type IFragment interface {
	Type() FragmentType
	IsSetInstNumber() bool
	SetInstNumber(instIdxInTx string)
}
