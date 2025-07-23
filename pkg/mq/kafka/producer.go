package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/ninja0404/meme-signal/pkg/logger"
)

func newProducerConfig(brokers []string, cfg KafkaProducerConfig) *kafka.ConfigMap {
	var kafkaconf = &kafka.ConfigMap{
		"api.version.request":           "true",
		"message.max.bytes":             10 * MB,
		"linger.ms":                     5,
		"sticky.partitioning.linger.ms": 0,
		"retries":                       3,
		"retry.backoff.ms":              1000,
		"acks":                          "1",
		"compression.type":              "snappy",
	}
	if cfg.MessageMaxBytes != 0 {
		kafkaconf.SetKey("message.max.bytes", cfg.MessageMaxBytes)
	}
	if cfg.LingerMs != 0 {
		kafkaconf.SetKey("linger.ms", cfg.LingerMs)
	}
	if cfg.PartitionLingerMs != 0 {
		kafkaconf.SetKey("sticky.partitioning.linger.ms", cfg.PartitionLingerMs)
	}
	if cfg.RetryBackoffMs != 0 {
		kafkaconf.SetKey("retry.backoff.ms", cfg.RetryBackoffMs)
	}
	if cfg.RequiredAcks != 0 {
		kafkaconf.SetKey("acks", cfg.RequiredAcks)
	}

	if cfg.ClientID != "" {
		kafkaconf.SetKey("client.id", cfg.ClientID+getClientID())
	}
	bootstrapServers := strings.Join(brokers, ",")
	kafkaconf.SetKey("bootstrap.servers", bootstrapServers)

	switch cfg.SecurityProtocol {
	case "PLAINTEXT", "":
		kafkaconf.SetKey("security.protocol", "plaintext")
	case "SASL_SSL":
		kafkaconf.SetKey("security.protocol", "sasl_ssl")
		//kafkaconf.SetKey("sasl.mechanism", cfg.SaslMechanism) // 或 SCRAM-SHA-256, SCRAM-SHA-512
		kafkaconf.SetKey("sasl.username", cfg.SaslUsername)
		kafkaconf.SetKey("sasl.password", cfg.SaslPassword)
		kafkaconf.SetKey("ssl.ca.location", cfg.SslCaLocation)
		kafkaconf.SetKey("ssl.certificate.location", cfg.SslCertificateLocation)
		kafkaconf.SetKey("ssl.key.location", cfg.SslKeyLocation)

		// hostname校验改成空,
		kafkaconf.SetKey("enable.ssl.certificate.verification", "false")
		kafkaconf.SetKey("ssl.endpoint.identification.algorithm", "None")
	case "SSL":
		kafkaconf.SetKey("security.protocol", "ssl")
		kafkaconf.SetKey("ssl.ca.location", cfg.SslCaLocation)
		kafkaconf.SetKey("ssl.certificate.location", cfg.SslCertificateLocation)
		kafkaconf.SetKey("ssl.key.location", cfg.SslKeyLocation)
		kafkaconf.SetKey("enable.ssl.certificate.verification", "false")
	case "SASL_PLAINTEXT":
		kafkaconf.SetKey("security.protocol", "sasl_plaintext")
		kafkaconf.SetKey("sasl.username", cfg.SaslUsername)
		kafkaconf.SetKey("sasl.password", cfg.SaslPassword)
		kafkaconf.SetKey("sasl.mechanism", cfg.SaslMechanism)
	default:
		panic(kafka.NewError(kafka.ErrUnknownProtocol, "unknown protocol", true))
	}

	return kafkaconf
}

type KafkaProducer struct {
	producer   *kafka.Producer
	eventsDone chan struct{}
}

func NewKafkaProducer(brokers []string, cfg KafkaProducerConfig) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(newProducerConfig(brokers, cfg))
	if err != nil {
		return nil, err
	}

	p := &KafkaProducer{
		producer:   producer,
		eventsDone: make(chan struct{}),
	}

	// 在创建时就启动事件处理
	go func() {
		for event := range producer.Events() {
			switch ev := event.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logger.Error("kafka_error: topic partition error",
						logger.FieldErr(ev.TopicPartition.Error),
						logger.String("topic", *ev.TopicPartition.Topic),
					)
				}
			case kafka.Error:
				logger.Error("kafka_error:error",
					logger.String("code", ev.Code().String()),
					logger.String("message", ev.Error()),
				)
			default:
				logger.Debug("kafka_event",
					logger.String("event", fmt.Sprintf("%T", ev)),
				)
			}
		}
		logger.Info("kafka events done")
		p.eventsDone <- struct{}{}
	}()

	return p, nil
}

func (p *KafkaProducer) SendMessage(topic string, value []byte) error {
	// 发送消息
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          value,
	}
	return p.producer.Produce(kafkaMsg, nil)
}

func (p *KafkaProducer) SendMessageWithKey(topic string, key string, value []byte) error {
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          value,
	}
	return p.producer.Produce(kafkaMsg, nil)
}

// Close 函数只负责关闭处理
func (p *KafkaProducer) Close() error {
	logger.Info("closing producer...")

	// 1. 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. 执行关闭
	done := make(chan error, 1)
	go func() {
		// Flush 消息
		remaining := p.producer.Flush(5 * 1000)
		if remaining > 0 {
			done <- fmt.Errorf("flush incomplete: %d messages remaining", remaining)
			return
		}

		// 关闭 producer
		p.producer.Close()

		// 等待eventsDone
		<-p.eventsDone
		done <- nil
	}()

	// 3. 等待完成或超时
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("close failed: %w", err)
		}
		logger.Info("producer closed successfully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("close timeout after 10s")
	}
}

func CloseProducer() error {
	return defaultProducer.Close()
}

func SendMessage(topic string, value []byte) error {
	return defaultProducer.SendMessage(topic, value)
}

func SendMessageWithKey(topic string, key string, value []byte) error {
	return defaultProducer.SendMessageWithKey(topic, key, value)
}
