package publisher

import (
	"context"
	"encoding/json"

	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/pkg/logger"
)

// Publisher 信号发布器接口
type Publisher interface {
	// Publish 发布信号
	Publish(signal *model.Signal) error

	// GetType 获取发布器类型
	GetType() string

	// Close 关闭发布器
	Close() error
}

// Manager 信号发布管理器
type Manager struct {
	publishers []Publisher
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewManager 创建发布管理器
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &Manager{
		publishers: make([]Publisher, 0),
		ctx:        ctx,
		cancel:     cancel,
	}

	return manager
}

// registerDefaultPublishers 注册默认发布器
func (m *Manager) registerDefaultPublishers() {
	// 注册日志发布器
	m.AddPublisher(&LogPublisher{})

	// 可以添加更多发布器：Telegram、Discord、WebHook等
	// m.AddPublisher(&TelegramPublisher{token: "your-bot-token"})
	// m.AddPublisher(&WebHookPublisher{url: "https://your-webhook-url"})
}

// AddPublisher 添加发布器
func (m *Manager) AddPublisher(publisher Publisher) {
	m.publishers = append(m.publishers, publisher)
	// 延迟到Start时输出日志
}

// PublishSignal 发布信号到所有发布器
func (m *Manager) PublishSignal(signal *model.Signal) {
	for _, publisher := range m.publishers {
		go func(p Publisher) {
			if err := p.Publish(signal); err != nil {
				logger.Error("发布信号失败",
					logger.String("publisher", p.GetType()),
					logger.String("signal_id", signal.ID),
					logger.FieldErr(err))
			}
		}(publisher)
	}
}

// Start 启动发布管理器
func (m *Manager) Start() error {
	// 注册默认发布器
	m.registerDefaultPublishers()

	// 输出已注册的发布器信息
	for _, publisher := range m.publishers {
		logger.Info("✅ 已加载信号发布器", logger.String("type", publisher.GetType()))
	}

	logger.Info("📡 信号发布管理器已启动")
	return nil
}

// Stop 停止发布管理器
func (m *Manager) Stop() error {
	m.cancel()

	// 关闭所有发布器
	for _, publisher := range m.publishers {
		if err := publisher.Close(); err != nil {
			logger.Error("关闭发布器失败",
				logger.String("type", publisher.GetType()),
				logger.FieldErr(err))
		}
	}

	logger.Info("信号发布管理器已停止")
	return nil
}

// LogPublisher 日志发布器 - 将信号输出到日志
type LogPublisher struct{}

func (p *LogPublisher) GetType() string {
	return "log"
}

func (p *LogPublisher) Publish(signal *model.Signal) error {
	logger.Info("🚨 发现交易信号",
		logger.String("signal_id", signal.ID),
		logger.String("type", string(signal.Type)),
		logger.String("token", signal.TokenAddress),
		logger.String("symbol", signal.TokenSymbol),
		logger.Int("severity", signal.Severity),
		logger.Float64("confidence", signal.Confidence),
		logger.String("message", signal.Message))

	// 如果有源交易，也记录相关信息
	if signal.SourceTx != nil {
		logger.Info("信号源交易信息",
			logger.String("signature", signal.SourceTx.Signature),
			logger.String("user", signal.SourceTx.UserWallet),
			logger.String("amount_usd", signal.SourceTx.AmountUSD.String()))
	}

	return nil
}

func (p *LogPublisher) Close() error {
	return nil
}

// ConsolePublisher 控制台发布器 - 格式化输出到控制台
type ConsolePublisher struct{}

func (p *ConsolePublisher) GetType() string {
	return "console"
}

func (p *ConsolePublisher) Publish(signal *model.Signal) error {
	// 格式化输出信号信息
	signalJSON, err := json.MarshalIndent(signal, "", "  ")
	if err != nil {
		return err
	}

	logger.Info("🚨 交易信号详情", logger.String("signal", string(signalJSON)))
	return nil
}

func (p *ConsolePublisher) Close() error {
	return nil
}
