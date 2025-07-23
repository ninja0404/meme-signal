package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/internal/repo"
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
	publishers      []Publisher
	ctx             context.Context
	cancel          context.CancelFunc
	tokenInfoRepo   repo.TokenInfoRepo
	tokenHolderRepo repo.TokenHolderRepo

	// 信号去重管理
	sentSignals    map[string]time.Time // key: tokenAddress_signalType, value: 发送时间
	signalCooldown time.Duration        // 信号冷却时间，防止重复发送
	mutex          sync.RWMutex         // 保护sentSignals的并发访问
}

// NewManager 创建发布管理器
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &Manager{
		publishers:     make([]Publisher, 0),
		ctx:            ctx,
		cancel:         cancel,
		sentSignals:    make(map[string]time.Time),
		signalCooldown: 1 * time.Hour, // 1小时内同一代币同一类型信号只发送一次
	}

	return manager
}

// SetRepositories 设置Repository
func (m *Manager) SetRepositories(tokenInfoRepo repo.TokenInfoRepo, tokenHolderRepo repo.TokenHolderRepo) {
	m.tokenInfoRepo = tokenInfoRepo
	m.tokenHolderRepo = tokenHolderRepo
}

// registerDefaultPublishers 注册默认发布器
func (m *Manager) registerDefaultPublishers() {
	// 注册日志发布器
	m.AddPublisher(&LogPublisher{})

	// 注册飞书发布器
	feishuWebhookURL := "https://open.larksuite.com/open-apis/bot/v2/hook/abacd303-7553-411b-b4db-fce9c2ef819c"
	m.AddPublisher(NewFeishuPublisher(feishuWebhookURL, m.tokenInfoRepo, m.tokenHolderRepo))

	// 可以添加更多发布器：Telegram、Discord、WebHook等
	// m.AddPublisher(&TelegramPublisher{token: "your-bot-token"})
	// m.AddPublisher(&WebHookPublisher{url: "https://your-webhook-url"})
}

// AddPublisher 添加发布器
func (m *Manager) AddPublisher(publisher Publisher) {
	m.publishers = append(m.publishers, publisher)
	// 延迟到Start时输出日志
}

// shouldSendSignal 检查是否应该发送信号（去重检查）
func (m *Manager) shouldSendSignal(signal *model.Signal) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 生成信号key：tokenAddress_signalType
	signalKey := fmt.Sprintf("%s_%s", signal.TokenAddress, string(signal.Type))

	// 检查是否在冷却期内
	if lastSentTime, exists := m.sentSignals[signalKey]; exists {
		if time.Since(lastSentTime) < m.signalCooldown {
			return false // 还在冷却期内，不发送
		}
	}

	return true // 可以发送
}

// recordSentSignal 记录已发送的信号
func (m *Manager) recordSentSignal(signal *model.Signal) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	signalKey := fmt.Sprintf("%s_%s", signal.TokenAddress, string(signal.Type))
	m.sentSignals[signalKey] = time.Now()
}

// cleanupExpiredSignals 清理过期的信号记录
func (m *Manager) cleanupExpiredSignals() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for signalKey, sentTime := range m.sentSignals {
		if now.Sub(sentTime) > m.signalCooldown {
			delete(m.sentSignals, signalKey)
		}
	}
}

// PublishSignal 发布信号到所有发布器
func (m *Manager) PublishSignal(signal *model.Signal) {
	// 检查信号去重
	if !m.shouldSendSignal(signal) {
		logger.Debug("⏭️ 信号已在冷却期内，跳过发送",
			logger.String("type", string(signal.Type)),
			logger.String("token", signal.TokenAddress))
		return
	}

	for _, publisher := range m.publishers {
		if err := publisher.Publish(signal); err != nil {
			logger.Error("发布信号失败",
				logger.String("publisher", publisher.GetType()),
				logger.String("signal_id", signal.ID),
				logger.FieldErr(err))
		} else {
			logger.Info("✅ 信号发布成功",
				logger.String("publisher", publisher.GetType()),
				logger.String("signal_id", signal.ID),
				logger.String("token", signal.TokenAddress))

			// 如果是飞书发布器且发送成功，记录已发送信号
			m.recordSentSignal(signal)
			logger.Debug("📝 已记录信号发送状态",
				logger.String("token", signal.TokenAddress),
				logger.String("type", string(signal.Type)))
		}
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

	// 启动定期清理过期信号记录的协程
	go m.startCleanupTask()

	return nil
}

// startCleanupTask 启动清理过期信号记录的定期任务
func (m *Manager) startCleanupTask() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupExpiredSignals()

			// 输出当前缓存的信号数量
			m.mutex.RLock()
			cachedCount := len(m.sentSignals)
			m.mutex.RUnlock()

			if cachedCount > 0 {
				logger.Debug("🧹 清理过期信号记录完成",
					logger.Int("cached_signals", cachedCount),
					logger.String("cooldown", m.signalCooldown.String()))
			}
		}
	}
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
