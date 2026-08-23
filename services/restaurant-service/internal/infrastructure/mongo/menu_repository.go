package mongo

import (
	"context"
	"errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"

	"go.mongodb.org/mongo-driver/bson"
	mdriver "go.mongodb.org/mongo-driver/mongo"
)

type MenuRepository struct {
	col *mdriver.Collection
}

func NewMenuRepository(db *mdriver.Database) *MenuRepository {
	return &MenuRepository{
		col: db.Collection(menuItemsCollection),
	}
}

func (r *MenuRepository) ListByRestaurantID(ctx context.Context, restaurantID string) ([]entities.MenuItem, error) {
	cursor, err := r.col.Find(ctx, bson.M{"restaurant_id": restaurantID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]entities.MenuItem, 0)
	for cursor.Next(ctx) {
		var doc menuItemDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		item, err := mapMenuItemDocToEntity(doc)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *MenuRepository) GetByID(ctx context.Context, restaurantID, itemID string) (entities.MenuItem, error) {
	var doc menuItemDoc
	err := r.col.FindOne(ctx, bson.M{"_id": itemID, "restaurant_id": restaurantID}).Decode(&doc)
	if errors.Is(err, mdriver.ErrNoDocuments) {
		return entities.MenuItem{}, domainerrors.ErrMenuItemNotFound
	}
	if err != nil {
		return entities.MenuItem{}, err
	}

	return mapMenuItemDocToEntity(doc)
}

func (r *MenuRepository) Save(ctx context.Context, item entities.MenuItem) error {
	_, err := r.col.InsertOne(ctx, mapMenuItemEntityToDoc(item))
	return err
}

func (r *MenuRepository) Update(ctx context.Context, item entities.MenuItem) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": item.ID, "restaurant_id": item.RestaurantID},
		bson.M{"$set": bson.M{
			"name":      item.Name,
			"price":     item.Price,
			"category":  string(item.Category),
			"available": item.Available,
		}},
	)
	return err
}

func mapMenuItemEntityToDoc(item entities.MenuItem) menuItemDoc {
	return menuItemDoc{
		ID:           item.ID,
		RestaurantID: item.RestaurantID,
		Name:         item.Name,
		Price:        item.Price,
		Category:     string(item.Category),
		Available:    item.Available,
	}
}

func mapMenuItemDocToEntity(doc menuItemDoc) (entities.MenuItem, error) {
	return entities.NewMenuItem(
		doc.ID,
		doc.RestaurantID,
		doc.Name,
		doc.Price,
		entities.MenuCategory(doc.Category),
		doc.Available,
	)
}
