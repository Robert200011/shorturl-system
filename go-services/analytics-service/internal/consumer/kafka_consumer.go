package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"

	"analytics-service/internal/model"
	"analytics-service/internal/service"
)

// KafkaConsumer Kafka消费者
type KafkaConsumer struct {
	consumer   sarama.ConsumerGroup
	aggregator *service.Aggregator
	topic      string
}

// NewKafkaConsumer 创建Kafka消费者
func NewKafkaConsumer(brokers []string, groupID, topic string, aggregator *service.Aggregator) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Version = sarama.V2_6_0_0

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Kafka consumer connected to %v, group: %s, topic: %s", brokers, groupID, topic)

	return &KafkaConsumer{
		consumer:   consumer,
		aggregator: aggregator,
		topic:      topic,
	}, nil
}

// Start 开始消费
func (c *KafkaConsumer) Start(ctx context.Context) error {
	handler := &consumerGroupHandler{
		aggregator: c.aggregator,
	}

	for {
		if err := c.consumer.Consume(ctx, []string{c.topic}, handler); err != nil {
			log.Printf("❌ Consumer error: %v", err)
			return err
		}

		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() error {
	if c.consumer != nil {
		return c.consumer.Close()
	}
	return nil
}

// consumerGroupHandler 消费者组处理器
type consumerGroupHandler struct {
	aggregator *service.Aggregator
}

// Setup 初始化
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("🔄 Consumer group setup")
	return nil
}

// Cleanup 清理
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("🔄 Consumer group cleanup")
	return nil
}

// ConsumeClaim 消费消息
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// 解析消息
		var event model.VisitEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("⚠️  Failed to unmarshal message: %v", err)
			session.MarkMessage(message, "")
			continue
		}

		log.Printf("📥 Received event: short_code=%s, device=%s, browser=%s",
			event.ShortCode, event.DeviceType, event.Browser)

		// 聚合数据
		ctx := context.Background()
		if err := h.aggregator.ProcessVisitEvent(ctx, &event); err != nil {
			log.Printf("⚠️  Failed to aggregate event: %v", err)
			// 继续处理下一条消息，不阻塞
		}

		// 标记消息已处理
		session.MarkMessage(message, "")
	}

	return nil
}
