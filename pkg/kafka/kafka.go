package kafka

import (
	"context"

	"github.com/IBM/sarama"
)

type KafkaConfig struct {
	Brokers []string `yaml:"KAFKA_BROKERS" env-required:"true"`
	Topic   string   `yaml:"KAFKA_TOPIC" env-required:"true"`
}

func NewSyncProducer(ctx context.Context, cfg KafkaConfig) (sarama.SyncProducer, error) {
	brokers := cfg.Brokers

	if len(brokers) == 0 {
		return nil, sarama.ErrBrokerNotFound
	}

	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner // распределяем сообщения по партициям случайным образом
	// config.Producer.Partitioner = sarama.NewRoundRobinPartitioner // распределяем сообщения по партициям по кругу
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll // ждём подтверждения от всех брокеров
	producer, err := sarama.NewSyncProducer(brokers, config)

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
