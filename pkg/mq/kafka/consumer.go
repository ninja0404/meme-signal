package kafka

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ninja0404/meme-signal/pkg/logger"
)

type MessageHandler func(message []byte) error
type wrapperMessageHandler func(message *kafka.Message) error

type KafkaConsumer struct {
	consumer    *kafka.Consumer
	config      *kafka.ConfigMap
	srcConfig   *KafkaConsumerConfig
	brokers     []string
	topics      []string
	groupId     string
	readTimeout int

	handlers map[string]wrapperMessageHandler

	done        chan struct{}
	closed      chan struct{}
	disablePool bool

	cancelCtx  context.Context
	cancelFunc context.CancelFunc
}

func NewKafkaConsumer(brokers []string, cfg KafkaConsumerConfig) (*KafkaConsumer, error) {
	config := newConsumerConfig(brokers, cfg)
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	instance := &KafkaConsumer{
		consumer:    consumer,
		config:      config,
		srcConfig:   &cfg,
		brokers:     brokers,
		topics:      cfg.Topics,
		groupId:     cfg.GroupId,
		disablePool: false,
		readTimeout: cfg.ReadTimeout,
		done:        make(chan struct{}),
		closed:      make(chan struct{}),
		handlers:    make(map[string]wrapperMessageHandler, 0),
		cancelCtx:   ctx,
		cancelFunc:  cancel,
	}
	return instance, nil
}

func (kc *KafkaConsumer) RegisterTopicHandler(t string, h MessageHandler) error {
	wrapperHandler := func(msg *kafka.Message) error {
		var err error
		defer func() {
			if r := recover(); r != nil {
				logger.Error("recovery from kafka message handler",
					logger.String("topic", *msg.TopicPartition.Topic),
					logger.Int32("partition", msg.TopicPartition.Partition),
					logger.String("offset", msg.TopicPartition.Offset.String()),
					logger.String("stack", string(debug.Stack())),
				)

				err = fmt.Errorf("panic in message handler: %v", r)
			}
		}()

		err = h(msg.Value)
		if err != nil {
			logger.Error("kafka message handler error",
				logger.FieldErr(err),
				logger.String("topic", *msg.TopicPartition.Topic))
			return err
		}
		return nil
	}
	for _, topic := range kc.topics {
		if topic == t {
			kc.handlers[t] = wrapperHandler
			return nil
		}
	}
	return errors.New("topic not in consumer list")
}

func (kc *KafkaConsumer) Close() error {
	// 关闭消费者
	if err := kc.consumer.Close(); err != nil {
		return fmt.Errorf("close consumer error: %w", err)
	}
	close(kc.done)

	logger.Info("consumer closed successfully")
	return nil
}

func (kc *KafkaConsumer) Start() error {
	subErr := kc.consumer.SubscribeTopics(kc.topics, nil)
	if subErr != nil {
		return subErr
	}

	// 如果是手动提交模式，启动定期提交goroutine
	//if !kc.srcConfig.EnableAutoCommit {
	//	go kc.periodicCommit()
	//}

	// 开始消费
	go func() {
		for {
			select {
			case <-kc.done:
				return
			default:
				msg, err := kc.consumer.ReadMessage(-1)

				if err != nil {
					logger.Error("kafka consumer read message error", logger.FieldErr(err), logger.Int("read_timeout", kc.readTimeout))
					continue
				}

				topic := *msg.TopicPartition.Topic

				h, ok := kc.handlers[topic]
				if !ok {
					logger.Warn("kafka consumer no handler for topic", logger.String("topic", topic))
					continue
				}

				// 处理消息
				for { // ⬅️ 内层重试循环，直到成功
					if err = h(msg); err == nil {
						// 处理成功：立即提交 offset
						if _, cErr := kc.consumer.CommitMessage(msg); cErr != nil {
							logger.Error("commit offset err", logger.FieldErr(cErr))
						}
						break // 跳出重试 -> 读下一条
					}

					// 处理失败：Seek 回当前 offset（依赖 Kafka 重发）
					tp := msg.TopicPartition
					if sErr := kc.consumer.Seek(tp, -1); sErr != nil {
						logger.Error("seek err", logger.FieldErr(sErr))
						time.Sleep(time.Second) // 避免忙等
					}

					// 重新阻塞式 ReadMessage，会再次拿到同一 offset
					msg, err = kc.consumer.ReadMessage(-1)
					if err != nil { // 网络等异常
						logger.Error("read msg err during retry", logger.FieldErr(err))
						continue
					}
				}
			}
		}
	}()
	return nil
}

// periodicCommit 定期提交offset
func (kc *KafkaConsumer) periodicCommit() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-kc.done:
			if _, err := kc.consumer.Commit(); err != nil {
				logger.Error("Final commit failed during shutdown", logger.FieldErr(err))
			}
			return
		case <-ticker.C:
			if _, err := kc.consumer.Commit(); err != nil {
				logger.Error("Periodic commit failed", logger.FieldErr(err))
			}
		}
	}
}
