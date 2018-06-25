package adapter

import (
	"log"

	producer "github.com/a8m/kinesis-producer"
)

const (
	topicName = "helpshift"
)

type KinesisAdapter struct {
	producer *producer.Producer
}

// NewKinesisAdapter create new KinesisAdapter
func NewKinesisAdapter(producer *producer.Producer) *KinesisAdapter {
	adapter := new(KinesisAdapter)
	adapter.producer = producer
	return adapter
}

// Send send data to kinesis topic
func (adapter *KinesisAdapter) Send(partitionKey string, content []byte) error {
	log.Printf("Sending data to kinesis")
	err := adapter.producer.Put([]byte("foo"), "bar")
	return err
}
