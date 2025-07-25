package repo

import (
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TokenHolderRepo 代币持有者数据访问接口
type TokenHolderRepo interface {
	// GetHolderCount 获取持仓人数（Amount > 0）
	GetHolderCount(tokenAddress string) (int64, error)

	// GetTokenHolders 获取所有持仓人
	GetTokenHolders(tokenAddress string) ([]*model.BiTokenHolder, error)

	// GetTop10HoldersRatio 获取top10持仓人的总持仓比例
	GetTop10HoldersRatio(tokenAddress string, supply decimal.Decimal) (float64, error)
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

// GetTop10HoldersRatio 获取top10持仓人的总持仓比例
func (r *tokenHolderRepoImpl) GetTop10HoldersRatio(tokenAddress string, supply decimal.Decimal) (float64, error) {
	if supply.IsZero() {
		return 0, nil
	}

	var holders []*model.BiTokenHolder

	// 查询所有持仓人，按持仓量降序排序，取前10个
	result := r.db.Where("token_address = ? AND amount > 0", tokenAddress).
		Order("amount DESC").
		Limit(10).
		Find(&holders)

	if result.Error != nil {
		return 0, result.Error
	}

	if len(holders) == 0 {
		return 0, nil
	}

	// 计算top10的总持仓量
	top10Total := decimal.Zero
	for _, holder := range holders {
		top10Total = top10Total.Add(holder.Amount)
	}

	// 计算比例（百分比）：top10持仓量 / 总供应量 * 100
	ratio := top10Total.Div(supply).InexactFloat64() * 100
	return ratio, nil
}
