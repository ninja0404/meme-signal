package fragment

import "github.com/gagliardetto/solana-go"

type TransferFragment struct {
	InstIdxInTx string

	WalletSrc  solana.PublicKey
	WalletDest solana.PublicKey
	AmountSrc  uint64
	AmountDest uint64
	Token      solana.PublicKey
	Decimals   int32
}

func (*TransferFragment) Type() FragmentType {
	return TransferType
}

func (t *TransferFragment) IsSetInstNumber() bool {
	return t.InstIdxInTx != ""
}

func (t *TransferFragment) SetInstNumber(instIdxInTx string) {
	t.InstIdxInTx = instIdxInTx
}
