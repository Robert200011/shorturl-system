package producer

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// VisitEvent 访问事件
type VisitEvent struct {
	ShortCode  string `json:"short_code"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	Referer    string `json:"referer"`
	DeviceType string `json:"device_type"`
	Browser    string `json:"browser"`
	OS         string `json:"os"`
	Timestamp  int64  `json:"timestamp"` // Unix timestamp
}

// KafkaProducer Kafka生产者
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaProducer 创建Kafka生产者
func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	log.Printf("✅ Kafka producer connected to %v, topic: %s", brokers, topic)

	return &KafkaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

// SendVisitEvent 发送访问事件
func (p *KafkaProducer) SendVisitEvent(event *VisitEvent) error {
	// 序列化为JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 创建消息
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.ShortCode), // 使用短链码作为key，保证同一短链的消息有序
		Value: sarama.ByteEncoder(data),
	}

	// 发送消息
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("📨 Sent visit event to Kafka: partition=%d, offset=%d, short_code=%s",
		partition, offset, event.ShortCode)

	return nil
}

// Close 关闭生产者
func (p *KafkaProducer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
