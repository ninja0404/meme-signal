package detector

import (
	"context"
	"hash/crc32"
	"sync"
	"time"

	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/pkg/logger"
	"github.com/shopspring/decimal"
)

const (
	// 时间窗口大小（5分钟）
	WindowSize = 5 * time.Minute
	// 时间桶大小（30秒为一个桶，5分钟=10个桶）
	BucketSize = 30 * time.Second
	// 桶数量
	BucketCount = int(WindowSize / BucketSize)
)

// Detector 信号检测器接口
type Detector interface {
	// Detect 检测信号
	Detect(stats *model.TokenStats, tx *model.Transaction) []*model.Signal

	// GetType 获取检测器类型
	GetType() string
}

// Worker 工作协程
type Worker struct {
	ID           int
	TokenWindows map[string]*TokenWindow
	TxChan       chan *model.Transaction
	SignalChan   chan *model.Signal
	Detectors    []Detector
	ctx          context.Context
	mutex        sync.RWMutex
}

// NewWorker 创建新的工作协程
func NewWorker(id int, ctx context.Context, signalChan chan *model.Signal) *Worker {
	return &Worker{
		ID:           id,
		TokenWindows: make(map[string]*TokenWindow),
		TxChan:       make(chan *model.Transaction, 100),
		SignalChan:   signalChan,
		Detectors:    make([]Detector, 0),
		ctx:          ctx,
	}
}

// AddDetector 添加检测器
func (w *Worker) AddDetector(detector Detector) {
	w.Detectors = append(w.Detectors, detector)
}

// Start 启动工作协程
func (w *Worker) Start() {
	go func() {
		// 定期清理过期数据
		cleanupTicker := time.NewTicker(time.Minute)
		defer cleanupTicker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case tx := <-w.TxChan:
				w.processTransaction(tx)
			case <-cleanupTicker.C:
				w.cleanup()
			}
		}
	}()
}

// processTransaction 处理交易
func (w *Worker) processTransaction(tx *model.Transaction) {
	tokenAddr := tx.TokenAddress
	if tokenAddr == "" {
		return
	}

	w.mutex.Lock()

	// 获取或创建代币窗口
	window, exists := w.TokenWindows[tokenAddr]
	if !exists {
		window = NewTokenWindow(tokenAddr)
		w.TokenWindows[tokenAddr] = window
	}

	w.mutex.Unlock()

	// 添加交易到窗口
	window.AddTransaction(tx)

	// 获取统计数据并运行检测器
	stats := window.GetStats()
	w.runDetectors(stats, tx)

	logger.Debug("📊 更新代币统计",
		logger.Int("worker_id", w.ID),
		logger.String("token", tokenAddr),
		logger.Int("tx_count", stats.TxCount5m),
		logger.Int("unique_wallets", stats.UniqueHolders))
}

// runDetectors 运行所有检测器
func (w *Worker) runDetectors(stats *model.TokenStats, tx *model.Transaction) {
	for _, detector := range w.Detectors {
		signals := detector.Detect(stats, tx)
		for _, signal := range signals {
			select {
			case w.SignalChan <- signal:
				logger.Info("🚨 Worker检测到信号",
					logger.Int("worker_id", w.ID),
					logger.String("type", string(signal.Type)),
					logger.String("token", signal.TokenAddress))
			case <-w.ctx.Done():
				return
			}
		}
	}
}

// cleanup 清理过期数据
func (w *Worker) cleanup() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	cutoff := time.Now().Add(-time.Hour) // 保留1小时内的数据
	for addr, window := range w.TokenWindows {
		if window.GetLastUpdate().Before(cutoff) {
			delete(w.TokenWindows, addr)
		}
	}
}

// Engine 信号检测引擎
type Engine struct {
	workers    []*Worker
	signalChan chan *model.Signal
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewEngine 创建信号检测引擎
func NewEngine() *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	engine := &Engine{
		workers:    make([]*Worker, WorkerCount),
		signalChan: make(chan *model.Signal, 1000),
		ctx:        ctx,
		cancel:     cancel,
	}

	// 创建工作协程
	for i := 0; i < WorkerCount; i++ {
		engine.workers[i] = NewWorker(i, ctx, engine.signalChan)
	}

	return engine
}

// Start 启动检测引擎
func (e *Engine) Start() error {
	// 使用复合检测器
	detectorList := e.createDefaultDetectors()

	for _, worker := range e.workers {
		for _, detector := range detectorList {
			worker.AddDetector(detector)
		}
		worker.Start()
	}

	logger.Info("🎯 信号检测引擎已启动",
		logger.Int("worker_count", WorkerCount),
		logger.String("window_size", WindowSize.String()),
		logger.Int("detector_count", len(detectorList)))

	return nil
}

// createDefaultDetectors 创建默认检测器
func (e *Engine) createDefaultDetectors() []Detector {
	// 直接实例化复合检测器，避免循环依赖
	return []Detector{
		&compositeSignalDetector{
			priceChangeThreshold: decimal.NewFromFloat(15.0), // 30%涨幅
			volumeThreshold:      decimal.NewFromInt(30000),  // 30k USD交易量
		},
	}
}

// compositeSignalDetector 复合信号检测器实现
type compositeSignalDetector struct {
	priceChangeThreshold decimal.Decimal // 价格涨幅阈值（百分比）
	volumeThreshold      decimal.Decimal // 交易量阈值（USD）
}

func (d *compositeSignalDetector) GetType() string {
	return "composite_signal"
}

func (d *compositeSignalDetector) Detect(stats *model.TokenStats, tx *model.Transaction) []*model.Signal {
	// 检查条件1：代币涨幅超过30%
	priceChangeOK := stats.PriceChangePercent.GreaterThanOrEqual(d.priceChangeThreshold)

	// 检查条件2：交易量超过250k USD
	//volumeOK := stats.Volume5m.GreaterThanOrEqual(d.volumeThreshold)

	// 只有当所有条件都满足时才发送信号
	if priceChangeOK {
		signal := &model.Signal{
			ID:           generateSignalID(),
			Type:         model.SignalTypeLargeTransaction,
			TokenAddress: stats.Address,
			TokenSymbol:  "",
			Severity:     d.calculateCompositeSeverity(stats.PriceChangePercent, stats.Volume5m),
			Confidence:   0.95, // 复合条件，置信度更高
			Message:      "检测到符合条件的代币信号",
			Data: map[string]interface{}{
				"price_change_percent":   stats.PriceChangePercent.String(),
				"price_change_threshold": d.priceChangeThreshold.String(),
				"volume_5min":            stats.Volume5m.String(),
				"volume_threshold":       d.volumeThreshold.String(),
				"start_price":            stats.StartPrice.String(),
				"current_price":          stats.CurrentPrice.String(),
				"tx_count_5min":          stats.TxCount5m,
				"unique_wallets":         stats.UniqueHolders,
				"conditions_met":         []string{"price_increase", "high_volume"},
			},
			Timestamp: time.Now(),
			SourceTx:  tx,
		}
		return []*model.Signal{signal}
	}

	return nil
}

// calculateCompositeSeverity 根据涨幅和交易量计算复合严重性等级
func (d *compositeSignalDetector) calculateCompositeSeverity(priceChangePercent, volume decimal.Decimal) int {
	// 基础分数
	baseScore := 5

	// 根据涨幅调整分数
	hundred := decimal.NewFromInt(100)
	sixty := decimal.NewFromInt(60)
	thirty := decimal.NewFromInt(30)

	if priceChangePercent.GreaterThanOrEqual(hundred) { // 涨幅超过100%
		baseScore += 3
	} else if priceChangePercent.GreaterThanOrEqual(sixty) { // 涨幅超过60%
		baseScore += 2
	} else if priceChangePercent.GreaterThanOrEqual(thirty) { // 涨幅超过30%
		baseScore += 1
	}

	// 根据交易量调整分数
	oneMillion := decimal.NewFromInt(1000000)
	fiveHundredK := decimal.NewFromInt(500000)

	if volume.GreaterThanOrEqual(oneMillion) { // 交易量超过1M
		baseScore += 2
	} else if volume.GreaterThanOrEqual(fiveHundredK) { // 交易量超过500k
		baseScore += 1
	}

	// 确保分数在1-10范围内
	if baseScore > 10 {
		baseScore = 10
	}
	if baseScore < 1 {
		baseScore = 1
	}

	return baseScore
}

// AddDetectors 添加外部检测器到所有worker
func (e *Engine) AddDetectors(detectors []Detector) {
	for _, worker := range e.workers {
		for _, detector := range detectors {
			worker.AddDetector(detector)
		}
	}
}

// ProcessTransaction 处理交易
func (e *Engine) ProcessTransaction(tx *model.Transaction) {
	tokenAddr := tx.TokenAddress
	if tokenAddr == "" {
		return
	}

	// 根据token地址hash分配到对应的worker
	hash := crc32.ChecksumIEEE([]byte(tokenAddr))
	workerIndex := int(hash) % WorkerCount

	select {
	case e.workers[workerIndex].TxChan <- tx:
	case <-e.ctx.Done():
	default:
		// 如果通道满了，丢弃消息并记录警告
		logger.Warn("⚠️ Worker通道已满，丢弃交易",
			logger.Int("worker_id", workerIndex),
			logger.String("token", tokenAddr))
	}
}

// Signals 获取信号通道
func (e *Engine) Signals() <-chan *model.Signal {
	return e.signalChan
}

// Stop 停止检测引擎
func (e *Engine) Stop() {
	logger.Info("🛑 停止信号检测引擎")
	e.cancel()
	close(e.signalChan)
}

// GetWorkerStats 获取worker统计信息
func (e *Engine) GetWorkerStats() map[int]int {
	stats := make(map[int]int)
	for i, worker := range e.workers {
		worker.mutex.RLock()
		stats[i] = len(worker.TokenWindows)
		worker.mutex.RUnlock()
	}
	return stats
}

func generateSignalID() string {
	return time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}
