package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type KafkaConfig struct {
	Brokers []string `yaml:"KAFKA_BROKERS" env-required:"true"`
	Topic   string   `yaml:"KAFKA_TOPIC" env-required:"true"`
}

func NewSyncProducer(ctx context.Context, log *zap.Logger, cfg KafkaConfig) (sarama.SyncProducer, error) {
	brokers := cfg.Brokers

	if len(brokers) == 0 {
		return nil, sarama.ErrBrokerNotFound
	}

	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner // распределяем сообщения по партициям случайным образом
	config.Producer.Retry.Max = 5                             // максимальное количество попыток отправки сообщения
	config.Producer.Retry.Backoff = 100 * time.Millisecond    // время ожидания между попытками
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll // ждём подтверждения от всех брокеров
	producer, err := sarama.NewSyncProducer(brokers, config)

	if err != nil {
		log.Error("failed to create kafka producer", zap.Error(err))
		return nil, err
	}

	log.Info("Kafka producer created", zap.Strings("brokers", brokers), zap.String("topic", cfg.Topic))
	return producer, err
}

func NewAsyncProducer(ctx context.Context, cfg KafkaConfig) (sarama.AsyncProducer, error) {
	brokers := cfg.Brokers

	if len(brokers) == 0 {
		return nil, sarama.ErrBrokerNotFound
	}

	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewAsyncProducer(brokers, config)

	return producer, err
}

func PrepareMessage(topic string, message []byte) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(message),
	}

	return msg
}
