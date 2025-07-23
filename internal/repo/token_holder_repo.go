package repo

import (
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TokenHolderRepo interface {
	// GetHolderCount 根据代币地址获取持仓人数(Amount > 0的用户数量)
	GetHolderCount(tokenAddress string) (int64, error)

	// GetTokenHolders 根据代币地址获取所有持仓用户
	GetTokenHolders(tokenAddress string) ([]*model.BiTokenHolder, error)
}

type tokenHolderRepoImpl struct {
	db *gorm.DB
}

func NewTokenHolderRepo(db *gorm.DB) TokenHolderRepo {
	return &tokenHolderRepoImpl{
		db: db,
	}
}

// GetHolderCount 根据代币地址获取持仓人数(Amount > 0的用户数量)
func (r *tokenHolderRepoImpl) GetHolderCount(tokenAddress string) (int64, error) {
	var count int64

	err := r.db.Model(&model.BiTokenHolder{}).
		Where("token_address = ? AND amount > ?", tokenAddress, decimal.Zero).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetTokenHolders 根据代币地址获取所有持仓用户
func (r *tokenHolderRepoImpl) GetTokenHolders(tokenAddress string) ([]*model.BiTokenHolder, error) {
	var holders []*model.BiTokenHolder

	err := r.db.
		Where("token_address = ? AND amount > ?", tokenAddress, decimal.Zero).
		Order("amount DESC").
		Find(&holders).Error

	if err != nil {
		return nil, err
	}

	return holders, nil
}
