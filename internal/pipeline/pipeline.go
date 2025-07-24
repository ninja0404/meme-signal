package pipeline

import (
	"context"

	"github.com/ninja0404/meme-signal/internal/detector"
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/internal/publisher"
	"github.com/ninja0404/meme-signal/internal/source"
	"github.com/ninja0404/meme-signal/pkg/logger"
)

// Pipeline 数据处理管道
type Pipeline struct {
	sourceManager     *source.Manager
	detectorEngine    *detector.Engine
	publisherManager  *publisher.Manager
	ctx               context.Context
	cancel            context.CancelFunc
	initialDataLoaded bool // 初始数据是否已加载完成
}

// NewPipeline 创建数据处理管道
func NewPipeline() *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pipeline{
		sourceManager:  source.NewManager(),
		detectorEngine: detector.NewEngine(),
		// publisherManager 延迟创建，等待配置设置
		ctx:               ctx,
		cancel:            cancel,
		initialDataLoaded: false, // 初始状态：数据未加载完成
	}
}

// SetPublisherConfig 设置发布器配置并创建发布管理器
func (p *Pipeline) SetPublisherConfig(config publisher.PublisherConfig) {
	p.publisherManager = publisher.NewManager(config)
}

// GetSourceManager 获取数据源管理器
func (p *Pipeline) GetSourceManager() *source.Manager {
	return p.sourceManager
}

// GetDetectorEngine 获取检测引擎
func (p *Pipeline) GetDetectorEngine() *detector.Engine {
	return p.detectorEngine
}

// GetPublisherManager 获取发布管理器
func (p *Pipeline) GetPublisherManager() *publisher.Manager {
	return p.publisherManager
}

// IsInitialDataLoaded 获取初始数据加载状态
func (p *Pipeline) IsInitialDataLoaded() bool {
	return p.initialDataLoaded
}

// Start 启动数据处理管道
func (p *Pipeline) Start() error {
	logger.Info("启动数据处理管道")

	// 启动检测引擎
	if err := p.detectorEngine.Start(); err != nil {
		return err
	}

	// 启动发布管理器
	if err := p.publisherManager.Start(); err != nil {
		return err
	}

	// 启动数据源管理器
	if err := p.sourceManager.Start(); err != nil {
		return err
	}

	// 启动数据处理协程
	go p.processTransactions()
	go p.processSignals()
	go p.processErrors()

	logger.Info("数据处理管道已启动")
	return nil
}

// Stop 停止数据处理管道
func (p *Pipeline) Stop() error {
	logger.Info("停止数据处理管道")

	// 取消上下文
	p.cancel()

	// 停止各个组件
	if err := p.sourceManager.Stop(); err != nil {
		logger.Error("停止数据源管理器失败", logger.FieldErr(err))
	}

	p.detectorEngine.Stop()

	if err := p.publisherManager.Stop(); err != nil {
		logger.Error("停止发布管理器失败", logger.FieldErr(err))
	}

	logger.Info("数据处理管道已停止")
	return nil
}

// processTransactions 处理交易数据
func (p *Pipeline) processTransactions() {
	txChan := p.sourceManager.Transactions()

	for {
		select {
		case <-p.ctx.Done():
			return
		case tx, ok := <-txChan:
			if !ok {
				return
			}

			// 检查初始数据是否已加载完成
			if !p.initialDataLoaded {
				if p.sourceManager.IsInitialDataLoaded() {
					p.initialDataLoaded = true
					// 通知检测引擎初始数据已加载完成
					p.detectorEngine.SetInitialDataLoaded(true)
					logger.Info("🎯 初始数据加载完成，开始正常信号检测")
				}
			}

			// 处理交易数据
			p.handleTransaction(tx)
		}
	}
}

// processSignals 处理检测到的信号
func (p *Pipeline) processSignals() {
	signalChan := p.detectorEngine.Signals()

	for {
		select {
		case <-p.ctx.Done():
			return
		case signal, ok := <-signalChan:
			if !ok {
				return
			}

			// 只有在初始数据加载完成后才发布信号
			if p.initialDataLoaded {
				// 发布信号
				p.publisherManager.PublishSignal(signal)
			} else {
				logger.Debug("⏳ 初始数据加载中，跳过信号发送",
					logger.String("signal_type", string(signal.Type)),
					logger.String("token", signal.TokenAddress))
			}
		}
	}
}

// processErrors 处理错误
func (p *Pipeline) processErrors() {
	errorChan := p.sourceManager.Errors()

	for {
		select {
		case <-p.ctx.Done():
			return
		case err, ok := <-errorChan:
			if !ok {
				return
			}

			// 记录错误
			logger.Error("数据源错误", logger.FieldErr(err))
		}
	}
}

// handleTransaction 处理单个交易
func (p *Pipeline) handleTransaction(tx *model.Transaction) {
	// 使用检测引擎处理交易
	p.detectorEngine.ProcessTransaction(tx)
}

// Stats 获取管道统计信息
type Stats struct {
	TransactionsProcessed int64 `json:"transactions_processed"`
	SignalsDetected       int64 `json:"signals_detected"`
	ErrorsCount           int64 `json:"errors_count"`
}

// GetStats 获取管道统计信息
func (p *Pipeline) GetStats() *Stats {
	// TODO: 实现统计信息收集
	return &Stats{
		TransactionsProcessed: 0,
		SignalsDetected:       0,
		ErrorsCount:           0,
	}
}
