package mongo

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
