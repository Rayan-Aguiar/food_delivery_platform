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
)

func EnsureIndexes(ctx context.Context, db *mdriver.Database) error {
	restaurants := db.Collection(restaurantsCollection)
	menuItems := db.Collection(menuItemsCollection)

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
	return err
}
