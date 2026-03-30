// internal/model/event.go
package model

// OrderCreatedEvent is published to Kafka topic "order.created"
type OrderCreatedEvent struct {
	OrderID    string `json:"order_id"`
	ActivityID string `json:"activity_id"`
	UserID     string `json:"user_id"`
}

// OrderPaidEvent is published to "order.paid"
type OrderPaidEvent struct {
	OrderID string `json:"order_id"`
}

// OrderCancelledEvent is published to "order.cancelled"
type OrderCancelledEvent struct {
	OrderID    string `json:"order_id"`
	ActivityID string `json:"activity_id"` // needed for stock rollback
}
