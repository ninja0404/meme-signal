package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ninja0404/meme-signal/internal/common"
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/internal/repo"
	"github.com/ninja0404/meme-signal/pkg/logger"
)

// Source 数据库数据源实现
type Source struct {
	txChan     chan *model.Transaction
	errChan    chan error
	ctx        context.Context
	cancel     context.CancelFunc
	config     SourceConfig
	repo       repo.SwapTxRepo
	lastId     uint64 // 最后处理的ID
	isFirstRun bool   // 是否首次运行
}

// SourceConfig 数据库数据源配置
type SourceConfig struct {
	QueryInterval     time.Duration // 查询间隔
	InitWindowMinutes int           // 初始查询窗口（分钟）
	BatchSize         int           // 批量查询大小
}

// NewSource 创建数据库数据源
func NewSource(config SourceConfig, swapTxRepo repo.SwapTxRepo) *Source {
	ctx, cancel := context.WithCancel(context.Background())

	return &Source{
		txChan:     make(chan *model.Transaction, 10000),
		errChan:    make(chan error, 100),
		ctx:        ctx,
		cancel:     cancel,
		config:     config,
		repo:       swapTxRepo,
		lastId:     0,
		isFirstRun: true,
	}
}

// Start 启动数据库数据源
func (s *Source) Start(ctx context.Context) error {
	logger.Info("🗄️ 启动数据库数据源",
		logger.String("query_interval", s.config.QueryInterval.String()),
		logger.Int("init_window_minutes", s.config.InitWindowMinutes),
		logger.Int("batch_size", s.config.BatchSize))

	// 启动轮询协程
	go s.startPolling()

	logger.Info("✅ 数据库数据源已启动")
	return nil
}

// Stop 停止数据库数据源
func (s *Source) Stop() error {
	logger.Info("🛑 停止数据库数据源")
	s.cancel()

	close(s.txChan)
	close(s.errChan)

	return nil
}

// Subscribe 获取交易数据通道
func (s *Source) Subscribe() <-chan *model.Transaction {
	return s.txChan
}

// Errors 获取错误通道
func (s *Source) Errors() <-chan error {
	return s.errChan
}

// startPolling 启动轮询
func (s *Source) startPolling() {
	ticker := time.NewTicker(s.config.QueryInterval)
	defer ticker.Stop()

	// 首次查询
	s.performInitialQuery()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performIncrementalQuery()
		}
	}
}

// performInitialQuery 执行首次查询（5分钟内所有数据）
func (s *Source) performInitialQuery() {
	logger.Info("🔍 执行首次查询，获取5分钟内所有交易数据")

	// 计算5分钟前的时间
	since := time.Now().Add(-time.Duration(s.config.InitWindowMinutes) * time.Minute)

	// 分批查询
	offset := 0
	totalProcessed := 0

	for {
		// 查询数据
		swapTxs, err := s.repo.GetLatestTransactions(since, s.config.BatchSize)
		if err != nil {
			s.sendError(fmt.Errorf("首次查询失败: %w", err))
			return
		}

		// 如果没有更多数据，结束查询
		if len(swapTxs) == 0 {
			break
		}

		// 处理数据
		for _, swapTx := range swapTxs {
			tx := s.convertSwapTxToTransaction(swapTx)
			if tx != nil {
				s.sendTransaction(tx)
				s.lastId = swapTx.ID // 更新最后处理的ID
				totalProcessed++
			}
		}

		offset += len(swapTxs)

		// 如果返回的数据少于批量大小，说明已经查询完毕
		if len(swapTxs) < s.config.BatchSize {
			break
		}

		// 避免过快查询，稍微延迟
		time.Sleep(100 * time.Millisecond)
	}

	s.isFirstRun = false
	logger.Info("✅ 首次查询完成",
		logger.Int("total_processed", totalProcessed),
		logger.Uint64("last_id", s.lastId))
}

// performIncrementalQuery 执行增量查询（查询最新ID之后的数据）
func (s *Source) performIncrementalQuery() {
	// 查询最新数据
	swapTxs, err := s.repo.GetTransactionsAfterId(s.lastId, s.config.BatchSize)
	if err != nil {
		s.sendError(fmt.Errorf("增量查询失败: %w", err))
		return
	}

	if len(swapTxs) == 0 {
		return // 没有新数据
	}

	// 处理新数据
	processedCount := 0
	for _, swapTx := range swapTxs {
		tx := s.convertSwapTxToTransaction(swapTx)
		if tx != nil {
			s.sendTransaction(tx)
			s.lastId = swapTx.ID
			processedCount++
		}
	}

	if processedCount > 0 {
		logger.Debug("📈 处理新交易数据",
			logger.Int("count", processedCount),
			logger.Uint64("last_id", s.lastId))
	}
}

// convertSwapTxToTransaction 将SwapTx转换为Transaction
func (s *Source) convertSwapTxToTransaction(swapTx *model.SwapTx) *model.Transaction {
	// 过滤：只处理买卖交易
	if swapTx.Action != 1 && swapTx.Action != 2 {
		return nil
	}

	// 过滤：价格必须大于0
	if !swapTx.PriceUSD.IsPositive() {
		return nil
	}

	// 过滤：交易金额必须大于0
	if !swapTx.AmountUSD.IsPositive() {
		return nil
	}

	// 生成交易ID
	txID := fmt.Sprintf("%s_%s", swapTx.TxHash, swapTx.InstIdxInTx)

	// 转换Action
	var action common.Action
	switch swapTx.Action {
	case 1:
		action = common.BuyAction
	case 2:
		action = common.SellAction
	default:
		return nil
	}

	return &model.Transaction{
		ID:           txID,
		Signature:    swapTx.TxHash,
		Slot:         swapTx.BlockID,
		BlockTime:    swapTx.BlockTime,
		UserWallet:   swapTx.UserWallet,
		TokenAddress: swapTx.TokenAddress,
		AmountUSD:    swapTx.AmountUSD,
		PriceUSD:     swapTx.PriceUSD,
		Action:       action,
	}
}

// sendTransaction 发送交易数据
func (s *Source) sendTransaction(tx *model.Transaction) {
	select {
	case s.txChan <- tx:
	case <-s.ctx.Done():
	}
}

// sendError 发送错误
func (s *Source) sendError(err error) {
	select {
	case s.errChan <- err:
	case <-s.ctx.Done():
	}
}

// String 数据源名称
func (s *Source) String() string {
	return "database"
}

// GetStats 获取数据源统计信息
func (s *Source) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"last_id":          s.lastId,
		"is_first_run":     s.isFirstRun,
		"query_interval":   s.config.QueryInterval.String(),
		"batch_size":       s.config.BatchSize,
		"tx_channel_size":  len(s.txChan),
		"err_channel_size": len(s.errChan),
	}
}
