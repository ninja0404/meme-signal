package kafka

import (
	"context"
	"fmt"

	"github.com/ninja0404/meme-signal/internal/common"
	"github.com/ninja0404/meme-signal/internal/model"
	"github.com/ninja0404/meme-signal/pkg/logger"
	"github.com/ninja0404/meme-signal/pkg/mq/kafka"
)

// Source Kafka数据源实现
type Source struct {
	txChan       chan *model.Transaction
	errChan      chan error
	ctx          context.Context
	cancel       context.CancelFunc
	config       SourceConfig
	consumerName string
}

// SourceConfig Kafka数据源配置
type SourceConfig struct {
	Topic       string
	Brokers     []string
	KafkaConfig kafka.KafkaConsumerConfig // 直接使用完整配置
}

// NewSource 创建Kafka数据源
func NewSource(config SourceConfig) *Source {
	ctx, cancel := context.WithCancel(context.Background())

	return &Source{
		txChan:       make(chan *model.Transaction, 1000),
		errChan:      make(chan error, 100),
		ctx:          ctx,
		cancel:       cancel,
		config:       config,
		consumerName: fmt.Sprintf("meme-signal-%s", config.KafkaConfig.GroupId),
	}
}

// Start 启动Kafka数据源
func (s *Source) Start(ctx context.Context) error {
	// 使用完整的Kafka配置，只覆盖Topic
	kafkaConfig := s.config.KafkaConfig
	kafkaConfig.Topics = []string{s.config.Topic}

	// 设置命名的Kafka消费者
	if err := kafka.SetupNamedKafkaConsumer(s.consumerName, s.config.Brokers, kafkaConfig); err != nil {
		return fmt.Errorf("设置Kafka消费者失败: %w", err)
	}

	// 注册消息处理器
	if err := kafka.RegisterTopicHandlerForConsumer(s.consumerName, s.config.Topic, s.handleMessage); err != nil {
		return fmt.Errorf("注册消息处理器失败: %w", err)
	}

	// 启动消费者
	if err := kafka.StartNamedConsumer(s.consumerName); err != nil {
		return fmt.Errorf("启动Kafka消费者失败: %w", err)
	}

	logger.Info("✅ Kafka数据源已启动",
		logger.String("topic", s.config.Topic),
		logger.String("group_id", s.config.KafkaConfig.GroupId),
		logger.String("consumer_name", s.consumerName))

	return nil
}

// Stop 停止Kafka数据源
func (s *Source) Stop() error {
	logger.Info("🛑 停止Kafka数据源")
	s.cancel()

	// 关闭命名的Kafka消费者
	if err := kafka.CloseNamedConsumer(s.consumerName); err != nil {
		logger.Error("关闭Kafka消费者失败", logger.FieldErr(err))
	}

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

// handleMessage 处理Kafka消息 - 使用MessageHandler签名
func (s *Source) handleMessage(data []byte) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("上下文已取消")
	default:
	}

	// 使用DecodeEvent解析二进制消息
	event, err := common.DecodeEvent(data)
	if err != nil {
		err = fmt.Errorf("解析事件数据失败: %w", err)
		select {
		case s.errChan <- err:
		case <-s.ctx.Done():
		}
		return err
	}

	// 只处理交易事件，忽略其他事件类型
	if event.Type == common.TradeEventType {
		// 处理交易事件
		tradeEvent := event.InnerEvent.(*common.TradeEvent)
		if tradeEvent.Action != common.BuyAction && tradeEvent.Action != common.SellAction {
			return nil
		}
		if !tradeEvent.PriceUSD.IsPositive() {
			return nil
		}
		transaction := s.convertTradeEventToTransaction(tradeEvent)

		select {
		case s.txChan <- transaction:
			logger.Debug("📨 处理交易事件",
				logger.String("signature", transaction.Signature),
				logger.String("token", transaction.TokenAddress),
				logger.String("amount_usd", transaction.AmountUSD.String()))
		case <-s.ctx.Done():
			return fmt.Errorf("上下文已取消")
		}
	}
	// 其他事件类型直接忽略，不处理也不记录日志

	return nil
}

// convertTradeEventToTransaction 将TradeEvent转换为model.Transaction
func (s *Source) convertTradeEventToTransaction(trade *common.TradeEvent) *model.Transaction {
	// 生成交易ID（使用signature和指令索引）
	txID := fmt.Sprintf("%s_%s", trade.Signature, trade.InstIdxInTx)

	return &model.Transaction{
		ID:           txID,
		Signature:    trade.Signature,
		Slot:         trade.Slot,
		BlockTime:    trade.BlockTime,
		UserWallet:   trade.UserWallet,
		TokenAddress: trade.TokenAddress,
		AmountUSD:    trade.AmountUSD,
		PriceUSD:     trade.PriceUSD,
		Action:       trade.Action,
	}
}

// String 数据源名称
func (s *Source) String() string {
	return fmt.Sprintf("kafka(%s)", s.config.Topic)
}

// GetStats 获取数据源统计信息
func (s *Source) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"topic":            s.config.Topic,
		"group_id":         s.config.KafkaConfig.GroupId,
		"consumer_name":    s.consumerName,
		"tx_channel_size":  len(s.txChan),
		"err_channel_size": len(s.errChan),
	}
}
