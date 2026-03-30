package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	w := &kafka.Writer{
		Addr: kafka.TCP(brokers...),
		// 将消息发送到当前分区中最少字节的分区，以实现负载均衡。
		Balancer: &kafka.LeastBytes{},
		// 如果主题不存在，自动创建主题。
		AllowAutoTopicCreation: true,
	}
	return &Producer{writer: w}
}

// Publish 将消息发布到指定的 Kafka 主题。
func (p *Producer) Publish(ctx context.Context, topic string, value any) error {
	// 将消息编码为 JSON 格式。
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: data,
	})
}

// Close 关闭 Kafka 生产者，释放资源。
func (p *Producer) Close() {
	err := p.writer.Close()
	if err != nil {
		return
	}
}
