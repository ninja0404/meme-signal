package repo

import (
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TokenInfoRepo interface {
	// GetTokenSymbol 根据代币地址获取代币符号
	GetTokenSymbol(tokenAddress string) (string, error)

	// GetTokenMarketData 根据代币地址获取代币市值计算所需数据
	GetTokenMarketData(tokenAddress string) (symbol string, currentPrice, supply decimal.Decimal, err error)

	// GetTokenInfo 根据代币地址获取代币信息
	GetTokenInfo(tokenAddress string) (*model.TokensInfo, error)
}

type tokenInfoRepoImpl struct {
	db *gorm.DB
}

func NewTokenInfoRepo(db *gorm.DB) TokenInfoRepo {
	return &tokenInfoRepoImpl{
		db: db,
	}
}

// GetTokenSymbol 根据代币地址获取代币符号
func (r *tokenInfoRepoImpl) GetTokenSymbol(tokenAddress string) (string, error) {
	var tokenInfo model.TokensInfo

	err := r.db.
		Select("symbol").
		Where("token_address = ?", tokenAddress).
		First(&tokenInfo).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "UNKNOWN", nil // 如果没找到，返回默认值
		}
		return "", err
	}

	return tokenInfo.Symbol, nil
}

// GetTokenMarketData 根据代币地址获取代币市值计算所需数据
func (r *tokenInfoRepoImpl) GetTokenMarketData(tokenAddress string) (symbol string, currentPrice, supply decimal.Decimal, err error) {
	var tokenInfo model.TokensInfo

	err = r.db.
		Select("symbol, current_price, supply").
		Where("token_address = ?", tokenAddress).
		First(&tokenInfo).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "UNKNOWN", decimal.Zero, decimal.Zero, nil // 如果没找到，返回默认值
		}
		return "", decimal.Zero, decimal.Zero, err
	}

	return tokenInfo.Symbol, tokenInfo.CurrentPrice, tokenInfo.Supply, nil
}

// GetTokenInfo 根据代币地址获取代币信息
func (r *tokenInfoRepoImpl) GetTokenInfo(tokenAddress string) (*model.TokensInfo, error) {
	var tokenInfo model.TokensInfo

	err := r.db.
		Where("token_address = ?", tokenAddress).
		First(&tokenInfo).Error

	if err != nil {
		return nil, err
	}

	return &tokenInfo, nil
}
