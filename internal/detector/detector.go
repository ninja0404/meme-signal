package detector

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"hash/crc32"

	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/pkg/logger"
)

const (
	// 时间窗口大小（5分钟）
	WindowSize = 5 * time.Minute
)

// Detector 信号检测器接口
type Detector interface {
	// Detect 检测信号
	Detect(stats *model.TokenStats, tx *model.Transaction, window *TokenWindow) []*model.Signal

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
		TxChan:       make(chan *model.Transaction, 100_000),
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
	//stats := window.GetStats()
	w.runDetectors(window, tx)
}

// runDetectors 运行所有检测器
func (w *Worker) runDetectors(window *TokenWindow, tx *model.Transaction) {
	stats := window.GetStats()
	for _, detector := range w.Detectors {
		signals := detector.Detect(stats, tx, window)
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

	now := time.Now()
	cutoff := now.Add(-WindowSize) // 保留5分钟内的数据，与时间窗口一致

	// 清理过期的代币窗口
	for addr, window := range w.TokenWindows {
		if window.GetLastUpdate().Before(cutoff) {
			delete(w.TokenWindows, addr)
		}
	}

	logger.Debug("🧹 Worker清理完成",
		logger.Int("worker_id", w.ID),
		logger.Int("active_tokens", len(w.TokenWindows)))
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

	// 启动统计监控协程
	go e.statsMonitor()

	return nil
}

// statsMonitor 统计监控协程，定期输出系统运行状态
func (e *Engine) statsMonitor() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟输出一次统计
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			// 获取所有Worker的统计信息
			totalTokens := 0
			workerStats := e.GetWorkerStats()

			for _, count := range workerStats {
				totalTokens += count
			}

			logger.Info("💹 检测引擎运行统计",
				logger.Int("total_workers", len(e.workers)),
				logger.Int("total_tokens_tracked", totalTokens),
				logger.String("window_size", WindowSize.String()))

			// 如果有代币在跟踪，输出最活跃的Worker统计
			if totalTokens > 0 {
				maxTokens := 0
				maxWorkerID := -1
				for workerID, count := range workerStats {
					if count > maxTokens {
						maxTokens = count
						maxWorkerID = workerID
					}
				}

				if maxWorkerID >= 0 {
					logger.Info("🔥 最活跃Worker",
						logger.Int("worker_id", maxWorkerID),
						logger.Int("tokens_tracked", maxTokens),
						logger.Float64("load_percentage", float64(maxTokens)/float64(totalTokens)*100))
				}
			}
		}
	}
}

// createDefaultDetectors 创建默认检测器
func (e *Engine) createDefaultDetectors() []Detector {
	// 使用新的配置化检测器系统
	registry := NewDetectorRegistry()

	// 注册默认的Meme信号检测器
	registry.Register("meme_signal", func() Detector {
		return registry.CreateMemeSignalDetector()
	})

	// 可以轻松添加更多检测器
	// registry.Register("volume_spike", func() Detector {
	//     return registry.CreateVolumeSpikeDetector()
	// })

	// 创建检测器实例
	memeDetector, err := registry.Create("meme_signal")
	if err != nil {
		logger.Error("❌ 创建检测器失败", logger.String("error", err.Error()))
		return []Detector{}
	}

	logger.Info("🔧 已加载配置化检测器",
		logger.String("detector", memeDetector.GetType()),
		logger.Any("registered", registry.GetRegisteredDetectors()))

	return []Detector{memeDetector}
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
	b := make([]byte, 8) // 8字节 = 16个十六进制字符
	_, err := rand.Read(b)
	if err != nil {
		// 降级到时间戳方案
		return time.Now().Format("20060102150405") + "_fallback"
	}
	return fmt.Sprintf("%x", b)
}

func randomString(length int) string {
	// 使用crypto/rand提供更好的随机性
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// 降级到时间戳方案
		return fmt.Sprintf("%d", time.Now().UnixNano())[:length]
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = chars[int(b[i])%len(chars)]
	}
	return string(result)
}
