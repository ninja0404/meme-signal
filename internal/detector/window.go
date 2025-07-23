package detector

import (
	"sync"
	"time"

	"github.com/ninja0404/meme-signal/internal/common"
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/shopspring/decimal"
)

// WindowSize 必须在外部定义，例如：const WindowSize = 5 * time.Minute
// InitialCapacity 为每个 token 窗口切片的预估初始容量
const InitialCapacity = 1000

// TokenWindow 高性能代币滑动窗口（增量统计）
type TokenWindow struct {
	TokenAddress string

	transactions []*model.Transaction // 当前窗口内的交易

	// 增量统计
	totalVolume decimal.Decimal
	buyCount    int
	sellCount   int
	walletCnt   map[string]uint32 // wallet → 交易次数

	firstPrice decimal.Decimal
	lastPrice  decimal.Decimal
	lastUpdate time.Time

	mutex sync.RWMutex
}

// NewTokenWindow 创建窗口实例
func NewTokenWindow(tokenAddress string) *TokenWindow {
	return &TokenWindow{
		TokenAddress: tokenAddress,
		transactions: make([]*model.Transaction, 0, InitialCapacity),
		walletCnt:    make(map[string]uint32, InitialCapacity/10),
		totalVolume:  decimal.Zero,
		firstPrice:   decimal.Zero,
		lastPrice:    decimal.Zero,
	}
}

// AddTransaction 在 O(1) 时间内新增一笔交易并更新统计
func (tw *TokenWindow) AddTransaction(tx *model.Transaction) {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	// 1. 先移除过期交易
	tw.removeExpiredTransactions(tx.BlockTime)

	// 2. 追加新交易（按时间递增假设）
	if len(tw.transactions) == cap(tw.transactions) {
		// 容量不足，按 1.5× 扩容
		newCap := cap(tw.transactions) * 3 / 2
		if newCap == 0 {
			newCap = 1
		}
		newSlice := make([]*model.Transaction, len(tw.transactions), newCap)
		copy(newSlice, tw.transactions)
		tw.transactions = newSlice
	}
	tw.transactions = append(tw.transactions, tx)
	tw.lastUpdate = tx.BlockTime

	// 3. 增量更新统计
	tw.addTransactionStats(tx)
}

// removeExpiredTransactions 会调整切片并同步减去统计量
func (tw *TokenWindow) removeExpiredTransactions(currentTime time.Time) {
	cutoff := currentTime.Add(-WindowSize)
	expired := 0

	// 统计过期条数
	for expired < len(tw.transactions) && tw.transactions[expired].BlockTime.Before(cutoff) {
		expired++
	}

	if expired == 0 {
		return
	}

	// 批量减去过期交易的统计
	for i := 0; i < expired; i++ {
		tw.removeTransactionStats(tw.transactions[i])
	}

	// 把剩余交易前移
	copy(tw.transactions, tw.transactions[expired:])
	// 截断
	for i := len(tw.transactions) - expired; i < len(tw.transactions); i++ {
		tw.transactions[i] = nil // 帮助 GC
	}
	tw.transactions = tw.transactions[:len(tw.transactions)-expired]

	// 刷新首价
	if len(tw.transactions) > 0 {
		tw.firstPrice = tw.transactions[0].PriceUSD
	} else {
		tw.firstPrice = decimal.Zero
		tw.lastPrice = decimal.Zero
	}
}

// addTransactionStats 增量+1
func (tw *TokenWindow) addTransactionStats(tx *model.Transaction) {
	tw.totalVolume = tw.totalVolume.Add(tx.AmountUSD)

	switch tx.Action {
	case common.BuyAction:
		tw.buyCount++
	case common.SellAction:
		tw.sellCount++
	}

	// 更新钱包出现次数
	tw.walletCnt[tx.UserWallet]++

	// 更新价格
	if tw.firstPrice.IsZero() {
		tw.firstPrice = tx.PriceUSD
	}
	tw.lastPrice = tx.PriceUSD
}

// removeTransactionStats 增量-1
func (tw *TokenWindow) removeTransactionStats(tx *model.Transaction) {
	tw.totalVolume = tw.totalVolume.Sub(tx.AmountUSD)

	switch tx.Action {
	case common.BuyAction:
		if tw.buyCount > 0 {
			tw.buyCount--
		}
	case common.SellAction:
		if tw.sellCount > 0 {
			tw.sellCount--
		}
	}

	// 钱包计数减少
	if c := tw.walletCnt[tx.UserWallet]; c <= 1 {
		delete(tw.walletCnt, tx.UserWallet)
	} else {
		tw.walletCnt[tx.UserWallet] = c - 1
	}
}

// GetStats O(1) 读取窗口统计
func (tw *TokenWindow) GetStats() *model.TokenStats {
	tw.mutex.RLock()
	defer tw.mutex.RUnlock()

	stats := &model.TokenStats{
		Address:    tw.TokenAddress,
		LastUpdate: tw.lastUpdate,
	}

	txCount := len(tw.transactions)
	if txCount == 0 {
		return stats
	}

	stats.TxCount5m = txCount
	stats.Volume5m = tw.totalVolume
	stats.UniqueHolders = len(tw.walletCnt)
	stats.StartPrice = tw.firstPrice
	stats.CurrentPrice = tw.lastPrice
	if !stats.StartPrice.IsZero() {
		stats.PriceChangePercent = stats.CurrentPrice.
			Sub(stats.StartPrice).
			Div(stats.StartPrice).
			Mul(decimal.NewFromInt(100))
	}
	return stats
}

// GetTransactionCount 返回当前窗口内交易条数
func (tw *TokenWindow) GetTransactionCount() int {
	tw.mutex.RLock()
	defer tw.mutex.RUnlock()
	return len(tw.transactions)
}

// GetLastUpdate 返回最后写入时间
func (tw *TokenWindow) GetLastUpdate() time.Time {
	tw.mutex.RLock()
	defer tw.mutex.RUnlock()
	return tw.lastUpdate
}

// GetWindowInfo 仅用于调试打印
func (tw *TokenWindow) GetWindowInfo() map[string]interface{} {
	tw.mutex.RLock()
	defer tw.mutex.RUnlock()

	info := map[string]interface{}{
		"token_address":     tw.TokenAddress,
		"transaction_count": len(tw.transactions),
		"capacity":          cap(tw.transactions),
	}
	if len(tw.transactions) == 0 {
		info["status"] = "empty"
		return info
	}

	firstTx := tw.transactions[0]
	lastTx := tw.transactions[len(tw.transactions)-1]
	windowStart := lastTx.BlockTime.Add(-WindowSize)

	info["window_start"] = windowStart.Format("15:04:05")
	info["window_end"] = lastTx.BlockTime.Format("15:04:05")
	info["first_tx_time"] = firstTx.BlockTime.Format("15:04:05")
	info["last_tx_time"] = lastTx.BlockTime.Format("15:04:05")
	info["first_price"] = tw.firstPrice.String()
	info["last_price"] = tw.lastPrice.String()
	info["unique_wallets"] = len(tw.walletCnt)
	info["total_volume"] = tw.totalVolume.String()
	info["buy_count"] = tw.buyCount
	info["sell_count"] = tw.sellCount

	return info
}

// ForceCleanup 仅单元测试用，强制依当前最后一笔的时间做滑窗清理
func (tw *TokenWindow) ForceCleanup() {
	tw.mutex.Lock()
	defer tw.mutex.Unlock()

	if len(tw.transactions) > 0 {
		lastTime := tw.transactions[len(tw.transactions)-1].BlockTime
		tw.removeExpiredTransactions(lastTime)
	}
}
