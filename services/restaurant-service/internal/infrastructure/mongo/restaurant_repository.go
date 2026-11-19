package mongo

import (
	"context"
	"errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"

	"go.mongodb.org/mongo-driver/bson"
	mdriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RestaurantRepository struct {
	col *mdriver.Collection
}

func NewRestaurantRepository(db *mdriver.Database) *RestaurantRepository {
	return &RestaurantRepository{
		col: db.Collection(restaurantsCollection),
	}
}

func (r *RestaurantRepository) List(ctx context.Context, query ports.ListRestaurantsQuery) ([]entities.Restaurant, error) {
	// Ordena por _id (indice padrao) para a paginacao ser deterministica sem indice extra.
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetSkip(int64(query.Offset)).
		SetLimit(int64(query.Limit))

	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	restaurants := make([]entities.Restaurant, 0, query.Limit)
	for cursor.Next(ctx) {
		var doc restaurantDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		restaurant, err := mapRestaurantDocToEntity(doc)
		if err != nil {
			return nil, err
		}
		restaurants = append(restaurants, restaurant)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return restaurants, nil
}

func (r *RestaurantRepository) GetByID(ctx context.Context, id string) (entities.Restaurant, error) {
	var doc restaurantDoc
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if errors.Is(err, mdriver.ErrNoDocuments) {
		return entities.Restaurant{}, domainerrors.ErrRestaurantNotFound
	}
	if err != nil {
		return entities.Restaurant{}, err
	}

	return mapRestaurantDocToEntity(doc)
}

func (r *RestaurantRepository) Save(ctx context.Context, restaurant entities.Restaurant) error {
	doc := restaurantDoc{
		ID:          restaurant.ID,
		Name:        restaurant.Name,
		Address:     restaurant.Address,
		PhoneNumber: restaurant.PhoneNumber,
		Status:      string(restaurant.Status),
		DeliveryFee: restaurant.DeliveryFee,
	}

	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc, options.Replace().SetUpsert(true))
	return err
}

func mapRestaurantDocToEntity(doc restaurantDoc) (entities.Restaurant, error) {
	return entities.NewRestaurant(
		doc.ID,
		doc.Name,
		doc.Address,
		doc.PhoneNumber,
		entities.RestaurantStatus(doc.Status),
		doc.DeliveryFee,
	)
}
