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

// PublisherConfig 发布器配置接口
type PublisherConfig interface {
	GetFeishuWebhookURL() string
}

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
	swapTxRepo      repo.SwapTxRepo
	config          PublisherConfig

	// 信号去重管理
	sentSignals    map[string]time.Time // key: tokenAddress_signalType, value: 发送时间
	signalCooldown time.Duration        // 信号冷却时间，防止重复发送

	// 跳过信号管理
	skippedSignals        map[string]time.Time // key: tokenAddress_signalType, value: 跳过时间
	skippedSignalCooldown time.Duration        // 跳过信号冷却时间，防止重复检测

	mutex sync.RWMutex // 保护sentSignals和skippedSignals的并发访问
}

// NewManager 创建发布管理器
func NewManager(config PublisherConfig) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &Manager{
		publishers:            make([]Publisher, 0),
		ctx:                   ctx,
		cancel:                cancel,
		config:                config,
		sentSignals:           make(map[string]time.Time),
		signalCooldown:        1 * time.Hour, // 1小时内同一代币同一类型信号只发送一次
		skippedSignals:        make(map[string]time.Time),
		skippedSignalCooldown: 30 * time.Minute, // 30分钟内被跳过的信号不再检测
	}

	return manager
}

// SetRepositories 设置Repository
func (m *Manager) SetRepositories(tokenInfoRepo repo.TokenInfoRepo, tokenHolderRepo repo.TokenHolderRepo, swapTxRepo repo.SwapTxRepo) {
	m.tokenInfoRepo = tokenInfoRepo
	m.tokenHolderRepo = tokenHolderRepo
	m.swapTxRepo = swapTxRepo
}

// registerDefaultPublishers 注册默认发布器
func (m *Manager) registerDefaultPublishers() {
	// 注册日志发布器
	m.AddPublisher(&LogPublisher{})

	// 注册飞书发布器
	if m.config != nil {
		webhookURL := m.config.GetFeishuWebhookURL()
		if webhookURL != "" {
			m.AddPublisher(NewFeishuPublisher(webhookURL, m.tokenInfoRepo, m.tokenHolderRepo))
		} else {
			logger.Warn("⚠️ 飞书发布器缺少webhook URL配置")
		}
	} else {
		logger.Warn("⚠️ Publisher配置为空")
	}

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

	// 清理过期的已发送信号
	for signalKey, sentTime := range m.sentSignals {
		if now.Sub(sentTime) > m.signalCooldown {
			delete(m.sentSignals, signalKey)
		}
	}

	// 清理过期的跳过信号
	for signalKey, skippedTime := range m.skippedSignals {
		if now.Sub(skippedTime) > m.skippedSignalCooldown {
			delete(m.skippedSignals, signalKey)
		}
	}
}

// shouldCheckSignal 检查是否应该检测信号（跳过信号检查）
func (m *Manager) shouldCheckSignal(signal *model.Signal) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 生成信号key：tokenAddress_signalType
	signalKey := fmt.Sprintf("%s_%s", signal.TokenAddress, string(signal.Type))

	// 检查是否在跳过信号冷却期内
	if lastSkippedTime, exists := m.skippedSignals[signalKey]; exists {
		if time.Since(lastSkippedTime) < m.skippedSignalCooldown {
			return false // 还在冷却期内，不检测
		}
	}

	return true // 可以检测
}

// recordSkippedSignal 记录被跳过的信号
func (m *Manager) recordSkippedSignal(signal *model.Signal, reason string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	signalKey := fmt.Sprintf("%s_%s", signal.TokenAddress, string(signal.Type))
	m.skippedSignals[signalKey] = time.Now()

	logger.Info("📝 已记录跳过信号状态",
		logger.String("token", signal.TokenAddress),
		logger.String("type", string(signal.Type)),
		logger.String("reason", reason),
		logger.String("cooldown", m.skippedSignalCooldown.String()))
}

// PublishSignal 发布信号到所有发布器
func (m *Manager) PublishSignal(signal *model.Signal) {
	// 检查是否在跳过信号冷却期内
	if !m.shouldCheckSignal(signal) {
		logger.Info("⏭️ 信号在跳过冷却期内，不再检测",
			logger.String("type", string(signal.Type)),
			logger.String("token", signal.TokenAddress),
			logger.String("cooldown", m.skippedSignalCooldown.String()))
		return
	}

	// 检查信号去重
	if !m.shouldSendSignal(signal) {
		logger.Debug("⏭️ 信号已在冷却期内，跳过发送",
			logger.String("type", string(signal.Type)),
			logger.String("token", signal.TokenAddress))
		return
	}

	// 检查捆绑交易占比
	bundleRatio := 0.0
	if m.swapTxRepo != nil {
		if ratio, err := m.swapTxRepo.GetTokenBundleRatio(signal.TokenAddress); err == nil {
			bundleRatio = ratio
			// 如果捆绑交易占比超过30%，跳过发送并记录
			if bundleRatio > 0.3 {
				logger.Info("🚫 捆绑交易占比过高，跳过发送信号",
					logger.String("token", signal.TokenAddress),
					logger.Float64("bundle_ratio", bundleRatio*100),
					logger.String("type", string(signal.Type)))
				m.recordSkippedSignal(signal, "捆绑交易占比过高")
				return
			}
		} else {
			logger.Warn("⚠️ 查询捆绑交易占比失败",
				logger.String("token", signal.TokenAddress),
				logger.FieldErr(err))
		}
	}

	// 检查钓鱼钱包占比
	phishingRatio := 0.0
	if m.swapTxRepo != nil && m.tokenInfoRepo != nil {
		// 获取代币信息（包含总供应量）
		if tokenInfo, err := m.tokenInfoRepo.GetTokenInfo(signal.TokenAddress); err == nil {
			// 将代币信息添加到信号数据中，供发布器使用，避免重复查询
			if signal.Data == nil {
				signal.Data = make(map[string]interface{})
			}
			signal.Data["token_symbol"] = tokenInfo.Symbol
			signal.Data["token_supply"] = tokenInfo.Supply
			signal.Data["current_price"] = tokenInfo.CurrentPrice.String()

			// 获取持仓地址列表
			var holderAddresses []string
			if holders, err := m.tokenHolderRepo.GetTokenHolders(signal.TokenAddress); err == nil {
				// 提取持仓地址
				holderAddresses = make([]string, len(holders))
				for i, holder := range holders {
					holderAddresses[i] = holder.WalletAddress
				}
			}

			// 查询钓鱼钱包持仓占比
			if ratio, err := m.swapTxRepo.GetTokenPhishingRatio(signal.TokenAddress, holderAddresses, tokenInfo.Supply); err == nil {
				phishingRatio = ratio
				// 如果钓鱼钱包占比超过20%，跳过发送并记录
				if phishingRatio > 20.0 {
					logger.Info("🚫 钓鱼钱包占比过高，跳过发送信号",
						logger.String("token", signal.TokenAddress),
						logger.Float64("phishing_ratio", phishingRatio),
						logger.String("type", string(signal.Type)))
					m.recordSkippedSignal(signal, "钓鱼钱包占比过高")
					return
				}
			} else {
				logger.Warn("⚠️ 查询钓鱼钱包占比失败",
					logger.String("token", signal.TokenAddress),
					logger.FieldErr(err))
			}
		} else {
			logger.Warn("⚠️ 查询代币信息失败",
				logger.String("token", signal.TokenAddress),
				logger.FieldErr(err))
		}
	}

	// 将占比信息添加到信号数据中，供发布器使用
	if signal.Data == nil {
		signal.Data = make(map[string]interface{})
	}
	signal.Data["bundle_ratio"] = bundleRatio
	signal.Data["phishing_ratio"] = phishingRatio

	if count, err := m.tokenHolderRepo.GetHolderCount(signal.TokenAddress); err == nil {
		if count < 200 {
			logger.Info("持仓人数小于200，跳过发送信号",
				logger.String("token", signal.TokenAddress),
				logger.Int64("holder_count", count))
			return
		}
		signal.Data["holder_count"] = count
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
				logger.String("token", signal.TokenAddress),
				logger.Float64("bundle_ratio", bundleRatio*100),
				logger.Float64("phishing_ratio", phishingRatio*100))

			// 如果是飞书发布器且发送成功，记录已发送信号
			m.recordSentSignal(signal)
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
			sentCount := len(m.sentSignals)
			skippedCount := len(m.skippedSignals)
			m.mutex.RUnlock()

			if sentCount > 0 || skippedCount > 0 {
				logger.Debug("🧹 清理过期信号记录完成",
					logger.Int("sent_signals", sentCount),
					logger.Int("skipped_signals", skippedCount),
					logger.String("sent_cooldown", m.signalCooldown.String()),
					logger.String("skipped_cooldown", m.skippedSignalCooldown.String()))
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
