package kafka

import (
	"github.com/IBM/sarama"
)

// var brokers = []string{"127.0.0.1:9095", "127.0.0.1:9096", "127.0.0.1:9097"}

func NewSyncProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)

	return producer, err
}

func SendMessage(producer sarama.SyncProducer, topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(message),
	}

	_, _, err := producer.SendMessage(msg)

	return err
}
