package fragment

import "github.com/shopspring/decimal"

const (
	TokenOptionUnknown int32 = iota
	TokenOptionCreate
	TokenOptionUpdateMint
	TokenOptionUpdateMetadata
)

type TokenFragment struct {
	TokenAddress    string          `json:"token_address"`
	MetadataAddress string          `json:"metadata_address"`
	Name            string          `json:"name"`
	Symbol          string          `json:"symbol"`
	Decimals        int32           `json:"decimals"`
	LogoURI         string          `json:"logo_uri"`
	FeeRate         decimal.Decimal `json:"fee_rate"`
	Creator         string          `json:"creator"`
	Supply          decimal.Decimal `json:"supply"`
	IsMutable       *bool           `json:"is_mutable"` // could be empty
	Option          int32           `json:"option"`

	InstIdxInTx string `json:"instIdxInTx,omitempty"`
}

func (t *TokenFragment) Type() FragmentType {
	return TokenType
}

func (t *TokenFragment) IsSetInstNumber() bool {
	return t.InstIdxInTx != ""
}

func (t *TokenFragment) SetInstNumber(instIdxInTx string) {
	t.InstIdxInTx = instIdxInTx
}
