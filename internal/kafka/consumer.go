package kafka

import (
	"context"
	"strings"

	"github.com/segmentio/kafka-go"
)

// Consumer wraps kafka.Reader.
type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID string, topics ...string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	return &Consumer{reader: r}
}

func (c *Consumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

func (c *Consumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return c.reader.CommitMessages(ctx, msgs...)
}

func (c *Consumer) Close() { c.reader.Close() }

func SplitBrokers(s string) []string {
	return strings.Split(s, ",")
}
