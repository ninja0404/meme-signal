package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ninja0404/meme-signal/internal/config"
	"github.com/ninja0404/meme-signal/internal/pipeline"
	"github.com/ninja0404/meme-signal/internal/repo"
	"github.com/ninja0404/meme-signal/internal/source/database"
	"github.com/ninja0404/meme-signal/pkg/database/polardbx"
	"github.com/ninja0404/meme-signal/pkg/logger"
	"gorm.io/gorm"
)

// Application meme交易信号监听应用
type Application struct {
	configManager *config.Manager
	pipeline      *pipeline.Pipeline
	db            *gorm.DB
	swapTxRepo    repo.SwapTxRepo
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

	logger.Info("📊 数据库连接已建立")
	return nil
}

// setupDataSources 设置数据源
func (app *Application) setupDataSources() {
	// 创建数据库数据源配置
	sourceConfig := database.SourceConfig{
		QueryInterval:     1 * time.Second, // 每秒查询一次
		InitWindowMinutes: 5,               // 首次查询5分钟内数据
		BatchSize:         1000,            // 每批查询1000条
	}

	// 创建数据库数据源
	dbSource := database.NewSource(sourceConfig, app.swapTxRepo)
	app.pipeline.GetSourceManager().AddSource(dbSource)

	logger.Info("🗄️ 已配置数据库数据源",
		logger.String("query_interval", sourceConfig.QueryInterval.String()),
		logger.Int("init_window_minutes", sourceConfig.InitWindowMinutes),
		logger.Int("batch_size", sourceConfig.BatchSize))
}

// Run 运行应用
func (app *Application) Run() error {
	logger.Info("🎯 启动Meme交易信号检测管道")

	// 启动数据处理管道
	if err := app.pipeline.Start(); err != nil {
		return err
	}

	logger.Info("🔥 Meme交易信号监听服务已启动，开始监控DEX交易...")
	logger.Info("📊 复合信号检测: 代币涨幅≥15% + 交易量≥30k USD")
	logger.Info("⚡ 分片处理架构: 16个Worker协程 | 5分钟时间窗口 | 基于Token地址Hash分片")
	logger.Info("🗄️ 数据源: 数据库轮询 | 每秒查询 | 增量处理")

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
