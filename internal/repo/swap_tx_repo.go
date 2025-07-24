package repo

import (
	"time"

	"github.com/ninja0404/meme-signal/internal/model"
	"gorm.io/gorm"
)

type SwapTxRepo interface {
	// GetLatestTransactions 获取指定时间范围内的最新交易
	GetLatestTransactions(since time.Time, limit int) ([]*model.SwapTx, error)

	// GetLatestTransactionsWithOffset 获取指定时间范围内的最新交易（支持分页）
	GetLatestTransactionsWithOffset(since time.Time, limit int, offset int) ([]*model.SwapTx, error)

	// GetTransactionsAfterId 获取指定ID之后的交易
	GetTransactionsAfterId(lastId uint64, limit int) ([]*model.SwapTx, error)

	// GetMaxId 获取当前最大ID
	GetMaxId() (uint64, error)

	// GetMinIdSince 获取指定时间之后的最小ID
	GetMinIdSince(since time.Time) (uint64, error)

	// GetTokenBundleRatio 获取指定代币的捆绑交易占比
	GetTokenBundleRatio(tokenAddress string) (float64, error)
}

type swapTxRepoImpl struct {
	db *gorm.DB
}

func NewSwapTxRepo(db *gorm.DB) SwapTxRepo {
	return &swapTxRepoImpl{
		db: db,
	}
}

// GetLatestTransactions 获取指定时间范围内的最新交易（用于首次初始化）
func (r *swapTxRepoImpl) GetLatestTransactions(since time.Time, limit int) ([]*model.SwapTx, error) {
	var transactions []*model.SwapTx

	err := r.db.
		Where("block_time >= ? AND action IN (?, ?)", since, 1, 2). // 只获取买卖交易
		Where("is_loss_tx = 0").
		Order("id ASC").
		Limit(limit).
		Find(&transactions).Error

	return transactions, err
}

// GetTransactionsAfterId 获取指定ID之后的交易（用于增量查询）
func (r *swapTxRepoImpl) GetTransactionsAfterId(lastId uint64, limit int) ([]*model.SwapTx, error) {
	var transactions []*model.SwapTx

	err := r.db.
		Where("id > ? AND action IN (?, ?)", lastId, 1, 2). // 只获取买卖交易
		Where("is_loss_tx = 0").
		Order("id ASC").
		Limit(limit).
		Find(&transactions).Error

	return transactions, err
}

// GetMaxId 获取当前最大ID
func (r *swapTxRepoImpl) GetMaxId() (uint64, error) {
	var maxId uint64

	err := r.db.Model(&model.SwapTx{}).
		Select("COALESCE(MAX(id), 0)").
		Scan(&maxId).Error

	return maxId, err
}

// GetLatestTransactionsWithOffset 获取指定时间范围内的最新交易（支持分页）
func (r *swapTxRepoImpl) GetLatestTransactionsWithOffset(since time.Time, limit int, offset int) ([]*model.SwapTx, error) {
	var transactions []*model.SwapTx

	err := r.db.
		Where("block_time >= ? AND action IN (?, ?)", since, 1, 2). // 只获取买卖交易
		Where("is_loss_tx = 0").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, err
}

// GetMinIdSince 获取指定时间之后的最小ID
func (r *swapTxRepoImpl) GetMinIdSince(since time.Time) (uint64, error) {
	var minId uint64

	err := r.db.Model(&model.SwapTx{}).
		Where("block_time >= ? AND action IN (?, ?)", since, 1, 2).
		Where("is_loss_tx = 0").
		Select("COALESCE(MIN(id), 0)").
		Scan(&minId).Error

	return minId, err
}

// GetTokenBundleRatio 获取指定代币的捆绑交易占比
func (r *swapTxRepoImpl) GetTokenBundleRatio(tokenAddress string) (float64, error) {
	var totalCount int64
	var bundledCount int64

	// 查询该代币的总交易数量（买卖交易）
	err := r.db.Model(&model.SwapTx{}).
		Where("token_address = ? AND action IN (?, ?)", tokenAddress, 1, 2).
		Count(&totalCount).Error
	if err != nil {
		return 0, err
	}

	// 如果没有交易，返回0
	if totalCount == 0 {
		return 0, nil
	}

	// 查询该代币的捆绑交易数量
	err = r.db.Model(&model.SwapTx{}).
		Where("token_address = ? AND action IN (?, ?) AND is_bundled = 1", tokenAddress, 1, 2).
		Count(&bundledCount).Error
	if err != nil {
		return 0, err
	}

	// 计算占比
	ratio := float64(bundledCount) / float64(totalCount)
	return ratio, nil
}
