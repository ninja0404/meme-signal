package kafka

import (
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/ninja0404/meme-signal/pkg/logger"
)

var defaultProducer *KafkaProducer
var defaultConsumer *KafkaConsumer

var consumers = make(map[string]*KafkaConsumer)
var consumerMutex sync.RWMutex

var startOnce sync.Once

func initKafka() {
	startOnce.Do(func() {
		sarama.Logger = NewLoggerKafka(logger.DefaultL1().Named("kafka-core"), LOGGER_INFO)
		sarama.DebugLogger = NewLoggerKafka(logger.DefaultL1().Named("kafka-core-debug"), LOGGER_DEBUG)
	})
}

func SetupKafkaProducer(brokers []string, cfg KafkaProducerConfig) error {
	initKafka()
	producer, err := NewKafkaProducer(brokers, cfg)
	if err != nil {
		return err
	}
	defaultProducer = producer
	//defaultProducer.Start()
	return nil
}

func SetupKafkaConsumer(brokers []string, cfg KafkaConsumerConfig) error {
	initKafka()
	instance, err := NewKafkaConsumer(brokers, cfg)
	if err != nil {
		return err
	}
	defaultConsumer = instance
	return nil
}

func CloseConsumer() error {
	return defaultConsumer.Close()
}

func StartConsumer() error {
	return defaultConsumer.Start()
}

func RegisterTopicHandler(t string, h MessageHandler) error {
	return defaultConsumer.RegisterTopicHandler(t, h)
}

func SetupNamedKafkaConsumer(name string, brokers []string, cfg KafkaConsumerConfig) error {
	initKafka()
	instance, err := NewKafkaConsumer(brokers, cfg)
	if err != nil {
		return err
	}

	consumerMutex.Lock()
	consumers[name] = instance
	consumerMutex.Unlock()

	return nil
}

func GetNamedConsumer(name string) *KafkaConsumer {
	consumerMutex.RLock()
	consumer := consumers[name]
	consumerMutex.RUnlock()
	return consumer
}

func StartNamedConsumer(name string) error {
	consumerMutex.RLock()
	consumer := consumers[name]
	consumerMutex.RUnlock()

	if consumer == nil {
		logger.Error("命名消费者不存在", logger.String("consumer", name))
		return fmt.Errorf("命名消费者不存在: %s", name)
	}

	return consumer.Start()
}

func CloseNamedConsumer(name string) error {
	consumerMutex.RLock()
	consumer := consumers[name]
	consumerMutex.RUnlock()

	if consumer == nil {
		logger.Error("命名消费者不存在", logger.String("consumer", name))
		return fmt.Errorf("命名消费者不存在: %s", name)
	}

	return consumer.Close()
}

func RegisterTopicHandlerForConsumer(name string, topic string, handler MessageHandler) error {
	consumerMutex.RLock()
	consumer := consumers[name]
	consumerMutex.RUnlock()

	if consumer == nil {
		logger.Error("命名消费者不存在", logger.String("consumer", name))
		return fmt.Errorf("命名消费者不存在: %s", name)
	}

	return consumer.RegisterTopicHandler(topic, handler)
}
