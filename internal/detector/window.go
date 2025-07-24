package detector

import (
	"sync"
	"time"

	"github.com/ninja0404/meme-signal/internal/common"
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/shopspring/decimal"
)

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

	// 计算价格变化百分比
	var priceChangePercent decimal.Decimal
	if !tw.firstPrice.IsZero() {
		priceChangePercent = tw.lastPrice.Sub(tw.firstPrice).
			Div(tw.firstPrice).
			Mul(decimal.NewFromInt(100))
	}

	return &model.TokenStats{
		Address:            tw.TokenAddress,
		TxCount5m:          len(tw.transactions),
		UniqueHolders:      len(tw.walletCnt),
		Volume5m:           tw.totalVolume,
		PriceChangePercent: priceChangePercent,
		CurrentPrice:       tw.lastPrice,
		StartPrice:         tw.firstPrice,      // 恢复: 5分钟窗口开始价格
		LastUpdate:         tw.lastUpdate,      // 恢复: 最后更新时间
		VolumeChange:       decimal.Zero,       // 恢复: 交易量变化百分比（暂时设为0）
		PriceChange:        priceChangePercent, // 恢复: 价格变化百分比
		Data:               nil,                // 恢复: 额外统计数据
	}
}

// GetMaxSingleTransactionAmount 获取5分钟内最大单笔交易金额
func (tw *TokenWindow) GetMaxSingleTransactionAmount() decimal.Decimal {
	tw.mutex.RLock()
	defer tw.mutex.RUnlock()

	maxAmount := decimal.Zero
	for _, tx := range tw.transactions {
		if tx.AmountUSD.GreaterThan(maxAmount) {
			maxAmount = tx.AmountUSD
		}
	}
	return maxAmount
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

// GetLast30SecondStats 获取最后30秒的统计数据（优化版：一次遍历获取所有统计）
func (tw *TokenWindow) GetLast30SecondStats() *model.TokenStats {
	return tw.getLast30SecondStatsWithBigTx(decimal.Zero)
}

// GetLast30SecondBigTransactionStats 获取最后30秒大额交易统计（复用优化后的方法）
func (tw *TokenWindow) GetLast30SecondBigTransactionStats(bigTransactionThreshold decimal.Decimal) map[string]interface{} {
	// 使用优化后的方法，避免重复遍历
	stats := tw.getLast30SecondStatsWithBigTx(bigTransactionThreshold)

	// 从stats.Data中提取大额交易统计
	if bigTxData, exists := stats.Data["big_tx_stats"]; exists {
		return bigTxData.(map[string]interface{})
	}

	// 如果没有大额交易数据，返回空统计
	return map[string]interface{}{
		"big_tx_users":      0,
		"big_tx_buy_count":  0,
		"big_tx_sell_count": 0,
		"total_big_tx":      0,
	}
}

// getLast30SecondStatsWithBigTx 优化后的内部方法：一次遍历获取所有30秒统计
func (tw *TokenWindow) getLast30SecondStatsWithBigTx(bigTransactionThreshold decimal.Decimal) *model.TokenStats {
	tw.mutex.RLock()
	defer tw.mutex.RUnlock()

	stats := &model.TokenStats{
		Address:    tw.TokenAddress,
		LastUpdate: tw.lastUpdate,
		Data:       make(map[string]interface{}),
	}

	if len(tw.transactions) == 0 {
		return stats
	}

	// 计算30秒前的时间点
	last30SecondsCutoff := tw.lastUpdate.Add(-30 * time.Second)

	// 普通统计变量
	var txCount30s int
	var volume30s decimal.Decimal = decimal.Zero
	var price30sAgo decimal.Decimal = decimal.Zero
	var buyCount30s, sellCount30s int
	wallets30s := make(map[string]bool)

	// 大额交易统计变量
	var needBigTxStats = !bigTransactionThreshold.IsZero()
	bigTxUsers := make(map[string]bool)
	bigTxBuyCount := 0
	bigTxSellCount := 0
	totalBigTx := 0

	// 一次遍历获取所有统计数据
	for i := len(tw.transactions) - 1; i >= 0; i-- {
		tx := tw.transactions[i]
		if tx.BlockTime.Before(last30SecondsCutoff) {
			// 如果这笔交易早于30秒前，记录30秒前的价格
			if price30sAgo.IsZero() {
				price30sAgo = tx.PriceUSD
			}
			break
		}

		// 这笔交易在30秒内 - 更新普通统计
		txCount30s++
		volume30s = volume30s.Add(tx.AmountUSD)
		wallets30s[tx.UserWallet] = true

		switch tx.Action {
		case common.BuyAction:
			buyCount30s++
		case common.SellAction:
			sellCount30s++
		}

		// 如果需要大额交易统计，同时更新大额交易统计
		if needBigTxStats && tx.AmountUSD.GreaterThanOrEqual(bigTransactionThreshold) {
			totalBigTx++
			bigTxUsers[tx.UserWallet] = true

			switch tx.Action {
			case common.BuyAction:
				bigTxBuyCount++
			case common.SellAction:
				bigTxSellCount++
			}
		}
	}

	// 如果没有30秒前的价格，使用最早的交易价格
	if price30sAgo.IsZero() && len(tw.transactions) > 0 {
		price30sAgo = tw.transactions[0].PriceUSD
	}

	// 计算30秒内的价格变化
	currentPrice := tw.lastPrice
	var priceChangePercent30s decimal.Decimal
	if !price30sAgo.IsZero() {
		priceChangePercent30s = currentPrice.
			Sub(price30sAgo).
			Div(price30sAgo).
			Mul(decimal.NewFromInt(100))
	}

	// 设置普通统计数据
	stats.TxCount5m = txCount30s // 复用字段表示30秒内交易数
	stats.Volume5m = volume30s   // 复用字段表示30秒内交易量
	stats.UniqueHolders = len(wallets30s)
	stats.StartPrice = price30sAgo
	stats.CurrentPrice = currentPrice
	stats.PriceChangePercent = priceChangePercent30s

	// 如果需要，设置大额交易统计数据
	if needBigTxStats {
		stats.Data["big_tx_stats"] = map[string]interface{}{
			"big_tx_users":      len(bigTxUsers),
			"big_tx_buy_count":  bigTxBuyCount,
			"big_tx_sell_count": bigTxSellCount,
			"total_big_tx":      totalBigTx,
		}
	}

	return stats
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
