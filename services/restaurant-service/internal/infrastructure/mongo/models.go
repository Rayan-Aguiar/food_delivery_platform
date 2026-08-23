package mongo

import "time"

type restaurantDoc struct {
	ID          string  `bson:"_id"`
	Name        string  `bson:"name"`
	Address     string  `bson:"address"`
	PhoneNumber string  `bson:"phone_number"`
	Status      string  `bson:"status"`
	DeliveryFee float64 `bson:"delivery_fee"`
}

type menuItemDoc struct {
	ID           string  `bson:"_id"`
	RestaurantID string  `bson:"restaurant_id"`
	Name         string  `bson:"name"`
	Price        float64 `bson:"price"`
	Category     string  `bson:"category"`
	Available    bool    `bson:"available"`
}

type outboxMessageDoc struct {
	ID         string            `bson:"_id"`
	Exchange   string            `bson:"exchange"`
	RoutingKey string            `bson:"routing_key"`
	Body       []byte            `bson:"body"`
	Headers    map[string]string `bson:"headers"`
	CreatedAt  time.Time         `bson:"created_at"`
	LastError  string            `bson:"last_error,omitempty"`
}
