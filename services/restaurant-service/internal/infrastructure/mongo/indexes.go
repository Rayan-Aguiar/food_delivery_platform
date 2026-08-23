package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	mdriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	restaurantsCollection = "restaurants"
	menuItemsCollection   = "menu_items"
	outboxCollection      = "restaurant_outbox"
)

func EnsureIndexes(ctx context.Context, db *mdriver.Database) error {
	restaurants := db.Collection(restaurantsCollection)
	menuItems := db.Collection(menuItemsCollection)
	outbox := db.Collection(outboxCollection)

	_, err := restaurants.Indexes().CreateOne(ctx, mdriver.IndexModel{
		Keys:    bson.D{{Key: "name", Value: "text"}},
		Options: options.Index().SetName("tx_restaurants_name"),
	})
	if err != nil {
		return err
	}

	_, err = menuItems.Indexes().CreateMany(ctx, []mdriver.IndexModel{
		{
			Keys:    bson.D{{Key: "restaurant_id", Value: 1}},
			Options: options.Index().SetName("ix_menu_items_restaurant_id"),
		},
		{
			Keys:    bson.D{{Key: "available", Value: 1}},
			Options: options.Index().SetName("ix_menu_items_available"),
		},
		{
			Keys:    bson.D{{Key: "name", Value: "text"}},
			Options: options.Index().SetName("tx_menu_items_name"),
		},
	})
	if err != nil {
		return err
	}

	_, err = outbox.Indexes().CreateMany(ctx, []mdriver.IndexModel{
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetName("ix_outbox_created_at"),
		},
		{
			Keys:    bson.D{{Key: "routing_key", Value: 1}},
			Options: options.Index().SetName("ix_outbox_routing_key"),
		},
	})
	return err
}
