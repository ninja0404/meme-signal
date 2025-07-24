package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ninja0404/meme-signal/internal/config"
	"github.com/ninja0404/meme-signal/internal/detector"
	"github.com/ninja0404/meme-signal/internal/detector/condition"
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/internal/pipeline"
	"github.com/ninja0404/meme-signal/internal/repo"
	"github.com/ninja0404/meme-signal/internal/source/database"
	"github.com/ninja0404/meme-signal/pkg/database/polardbx"
	"github.com/ninja0404/meme-signal/pkg/logger"
	"gorm.io/gorm"
)

// Application meme交易信号监听应用
type Application struct {
	configManager   *config.Manager
	pipeline        *pipeline.Pipeline
	db              *gorm.DB
	swapTxRepo      repo.SwapTxRepo
	tokenInfoRepo   repo.TokenInfoRepo
	tokenHolderRepo repo.TokenHolderRepo
}

// New 创建新的meme信号应用实例
func New() *Application {
	return &Application{
		configManager: config.NewManager(),
		pipeline:      pipeline.NewPipeline(),
	}
}

// Initialize 初始化应用
func (app *Application) Initialize(configPath string) error {
	// 1. 加载配置
	if err := app.configManager.Load(configPath); err != nil {
		return err
	}

	// 2. 初始化日志系统
	if err := app.configManager.InitLogger(); err != nil {
		return err
	}
	logger.Info("🚀 Meme交易信号监听服务初始化开始", logger.String("config_path", configPath))

	// 3. 初始化数据库
	if err := app.initDatabase(); err != nil {
		return err
	}

	// 4. 设置数据源
	app.setupDataSources()

	logger.Info("✅ Meme交易信号监听服务初始化完成")
	return nil
}

// initDatabase 初始化数据库连接
func (app *Application) initDatabase() error {
	// 从默认配置初始化数据库
	if err := polardbx.SetupDatabaseFromDefaultConfig(); err != nil {
		return err
	}

	// 获取数据库连接
	db, err := polardbx.GetDb()
	if err != nil {
		return err
	}
	app.db = db

	// 创建SwapTx仓储
	app.swapTxRepo = repo.NewSwapTxRepo(db)

	// 创建TokenInfo仓储
	app.tokenInfoRepo = repo.NewTokenInfoRepo(db)

	// 创建TokenHolder仓储
	app.tokenHolderRepo = repo.NewTokenHolderRepo(db)

	logger.Info("📊 数据库连接已建立")
	return nil
}

// setupDataSources 设置数据源
func (app *Application) setupDataSources() {
	// 创建数据库数据源配置
	sourceConfig := database.SourceConfig{
		QueryInterval:     1 * time.Second, // 每秒查询一次
		InitWindowMinutes: 5,               // 首次查询5分钟内数据
		BatchSize:         10000,           // 每批查询1000条
	}

	// 创建数据库数据源
	dbSource := database.NewSource(sourceConfig, app.swapTxRepo)
	app.pipeline.GetSourceManager().AddSource(dbSource)

	// 设置发布管理器配置
	publisherConfig := app.configManager.GetPublisherConfig()
	app.pipeline.SetPublisherConfig(publisherConfig)

	// 配置发布管理器的Repository
	app.pipeline.GetPublisherManager().SetRepositories(app.tokenInfoRepo, app.tokenHolderRepo, app.swapTxRepo)

	logger.Info("🗄️ 已配置数据库数据源",
		logger.String("query_interval", sourceConfig.QueryInterval.String()),
		logger.Int("init_window_minutes", sourceConfig.InitWindowMinutes),
		logger.Int("batch_size", sourceConfig.BatchSize))
}

// setupDetectors 设置检测器
func (app *Application) setupDetectors() *detector.DetectorRegistry {
	detectorRegistry := detector.NewDetectorRegistry()

	// 注册混合Meme信号检测器
	detectorRegistry.Register("meme_signal", func() detector.Detector {
		return detectorRegistry.CreateMemeSignalDetector()
	})

	// 注册巨鲸活动检测器 - 使用ConditionFactory保持架构一致性
	detectorRegistry.Register("whale_activity", func() detector.Detector {
		// 创建条件工厂
		factory := condition.NewConditionFactory()

		// 使用工厂方法创建突然性巨鲸活动条件
		whaleCondition := factory.CreateWhaleTransactionCondition(
			"突然性巨鲸活动检测",
			"检测平静期突然出现的大额交易",
			40000.0, // 平静期最大交易量: $40,000
			5000.0,  // 平静期最大单笔: $5,000
			10000.0, // 突破交易阈值: $10,000
			80,      // 平静期最大交易数: 80笔
		)

		return detector.NewDetectorBuilder().
			Name("突然性巨鲸活动检测器").
			Description("检测平静期突然出现的大额交易（平静期<$40k，突然>$10k）").
			Type("whale_activity").
			SignalType(model.SignalTypeWhaleActivity).
			Severity(7).
			Confidence(0.9).
			WithCondition(whaleCondition).
			Build()
	})

	logger.Info("🔧 检测器注册完成",
		logger.String("meme_signal", "复合Meme信号检测器"),
		logger.String("whale_activity", "巨鲸活动检测器"))

	return detectorRegistry
}

// Run 运行应用
func (app *Application) Run() error {
	logger.Info("🎯 启动Meme交易信号检测管道")

	// 设置检测器注册表
	detectorRegistry := app.setupDetectors()

	// 将检测器注册表传递给引擎，实现统一管理
	detectorEngine := app.pipeline.GetDetectorEngine()
	detectorEngine.SetDetectorRegistry(detectorRegistry)

	// 启动数据处理管道
	if err := app.pipeline.Start(); err != nil {
		return err
	}

	logger.Info("🔥 Meme交易信号监听服务已启动，开始监控DEX交易...")
	logger.Info("📊 复合信号检测: 5分钟内涨幅≥20% + 最后30秒涨幅≥15% + 5分钟内交易次数>300笔 + 独立钱包数>50个 + 大额交易条件")
	logger.Info("💰 大额交易条件: 30秒内>1000U交易的用户数≥5个 + 大额买卖比≥2:1")
	logger.Info("🐋 巨鲸检测: 单笔交易金额>10,000USD的大额交易监控")
	logger.Info("⚡ 分片处理架构: 16个Worker协程 | 5分钟时间窗口 | 基于Token地址Hash分片")
	logger.Info("🗄️ 数据源: 数据库轮询 | 每秒查询 | 增量处理")
	logger.Info("🔄 信号去重: 1小时冷却期 | 防止重复发送 | 每个代币每种信号类型限制")

	// 等待终止信号
	app.waitForShutdown()

	return nil
}

// waitForShutdown 等待关闭信号
func (app *Application) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	sig := <-quit
	logger.Info("📤 收到终止信号，开始优雅关闭应用...", logger.String("signal", sig.String()))

	// 优雅关闭
	app.Shutdown()
}

// Shutdown 优雅关闭应用
func (app *Application) Shutdown() {
	logger.Info("🛑 开始关闭Meme交易信号监听服务...")

	// 停止数据处理管道
	if err := app.pipeline.Stop(); err != nil {
		logger.Error("停止数据处理管道失败", logger.FieldErr(err))
	}

	// 关闭数据库连接
	if err := polardbx.Stop(); err != nil {
		logger.Error("关闭数据库连接失败", logger.FieldErr(err))
	}

	// 获取统计信息
	stats := app.pipeline.GetStats()
	workerStats := app.pipeline.GetDetectorEngine().GetWorkerStats()

	// 计算worker负载均衡情况
	totalTokens := 0
	for _, count := range workerStats {
		totalTokens += count
	}

	logger.Info("📈 服务运行统计",
		logger.Int64("transactions_processed", stats.TransactionsProcessed),
		logger.Int64("signals_detected", stats.SignalsDetected),
		logger.Int64("errors_count", stats.ErrorsCount),
		logger.Int("total_tokens_tracked", totalTokens))

	logger.Info("⚡ Worker负载分布", logger.Any("worker_token_counts", workerStats))

	logger.Info("✨ Meme交易信号监听服务已成功关闭")
}

// Start 启动应用的便捷方法
func (app *Application) Start(configPath string) error {
	// 初始化
	if err := app.Initialize(configPath); err != nil {
		logger.Error("❌ Meme交易信号服务初始化失败", logger.FieldErr(err))
		return err
	}

	// 运行
	if err := app.Run(); err != nil {
		logger.Error("❌ Meme交易信号服务运行失败", logger.FieldErr(err))
		return err
	}

	return nil
}

// GetPipeline 获取数据处理管道（用于调试和监控）
func (app *Application) GetPipeline() *pipeline.Pipeline {
	return app.pipeline
}

// GetConfigManager 获取配置管理器
func (app *Application) GetConfigManager() *config.Manager {
	return app.configManager
}

// GetDatabase 获取数据库连接
func (app *Application) GetDatabase() *gorm.DB {
	return app.db
}

// GetSwapTxRepo 获取SwapTx仓储
func (app *Application) GetSwapTxRepo() repo.SwapTxRepo {
	return app.swapTxRepo
}

// GetTokenInfoRepo 获取TokenInfo仓储
func (app *Application) GetTokenInfoRepo() repo.TokenInfoRepo {
	return app.tokenInfoRepo
}

// GetTokenHolderRepo 获取TokenHolder仓储
func (app *Application) GetTokenHolderRepo() repo.TokenHolderRepo {
	return app.tokenHolderRepo
}
