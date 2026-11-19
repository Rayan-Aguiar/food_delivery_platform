//go:build integration

// Testes de integracao contra um Mongo real. Rodam so com a build tag:
//
//	MONGO_TEST_URI=mongodb://localhost:27018 go test -tags=integration ./internal/infrastructure/mongo/...
//
// Cada teste usa um database proprio e o remove ao final, entao podem rodar
// contra o container de desenvolvimento sem tocar em restaurant_db.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"

	"go.mongodb.org/mongo-driver/bson"
	mdriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newTestDB(t *testing.T) *mdriver.Database {
	t.Helper()

	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping Mongo integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mdriver.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping: %v", err)
	}

	db := client.Database(fmt.Sprintf("restaurant_it_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Drop(ctx); err != nil {
			t.Errorf("drop test database: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("disconnect: %v", err)
		}
	})

	if err := EnsureIndexes(ctx, db); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	return db
}

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func mustNewRestaurant(t *testing.T, id string, status entities.RestaurantStatus) entities.Restaurant {
	t.Helper()
	r, err := entities.NewRestaurant(id, "Restaurante "+id, "Rua A, 1", "+5511999999999", status, 7.5)
	if err != nil {
		t.Fatalf("new restaurant: %v", err)
	}
	return r
}

func mustNewMenuItem(t *testing.T, id, restaurantID string, available bool) entities.MenuItem {
	t.Helper()
	item, err := entities.NewMenuItem(id, restaurantID, "Item "+id, 29.9, entities.MenuCategoryPizzaria, available)
	if err != nil {
		t.Fatalf("new menu item: %v", err)
	}
	return item
}

func TestIntegration_EnsureIndexes_IsIdempotent(t *testing.T) {
	db := newTestDB(t) // ja chamou EnsureIndexes uma vez
	ctx := testCtx(t)

	if err := EnsureIndexes(ctx, db); err != nil {
		t.Fatalf("second EnsureIndexes must succeed (startup roda em todo deploy): %v", err)
	}

	cursor, err := db.Collection(menuItemsCollection).Indexes().List(ctx)
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		t.Fatalf("decode indexes: %v", err)
	}

	names := map[string]bool{}
	for _, idx := range indexes {
		names[idx["name"].(string)] = true
	}
	for _, want := range []string{"ix_menu_items_restaurant_id", "ix_menu_items_available", "tx_menu_items_name"} {
		if !names[want] {
			t.Errorf("missing index %q on menu_items", want)
		}
	}
}

func TestIntegration_RestaurantRepository_SaveAndGetByID(t *testing.T) {
	repo := NewRestaurantRepository(newTestDB(t))
	ctx := testCtx(t)

	want := mustNewRestaurant(t, "rest-1", entities.RestaurantStatusActive)
	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.GetByID(ctx, "rest-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != want {
		t.Errorf("round trip mismatch:\n got  %+v\n want %+v", got, want)
	}

	// Save e' upsert: salvar de novo com o mesmo id substitui, nao duplica.
	updated := mustNewRestaurant(t, "rest-1", entities.RestaurantStatusInactive)
	if err := repo.Save(ctx, updated); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, err = repo.GetByID(ctx, "rest-1")
	if err != nil {
		t.Fatalf("get after upsert: %v", err)
	}
	if got.Status != entities.RestaurantStatusInactive {
		t.Errorf("status after upsert = %q, want inactive", got.Status)
	}
}

func TestIntegration_RestaurantRepository_GetByID_NotFound(t *testing.T) {
	repo := NewRestaurantRepository(newTestDB(t))

	_, err := repo.GetByID(testCtx(t), "missing")
	if !errors.Is(err, domainerrors.ErrRestaurantNotFound) {
		t.Fatalf("err = %v, want ErrRestaurantNotFound", err)
	}
}

func TestIntegration_RestaurantRepository_ListPaginates(t *testing.T) {
	repo := NewRestaurantRepository(newTestDB(t))
	ctx := testCtx(t)

	// Insere fora de ordem para provar que a ordenacao vem da query, nao da insercao.
	for _, id := range []string{"r-3", "r-1", "r-5", "r-2", "r-4"} {
		if err := repo.Save(ctx, mustNewRestaurant(t, id, entities.RestaurantStatusActive)); err != nil {
			t.Fatalf("save %s: %v", id, err)
		}
	}

	tests := []struct {
		name    string
		query   ports.ListRestaurantsQuery
		wantIDs []string
	}{
		{"first page", ports.ListRestaurantsQuery{Offset: 0, Limit: 2}, []string{"r-1", "r-2"}},
		{"second page", ports.ListRestaurantsQuery{Offset: 2, Limit: 2}, []string{"r-3", "r-4"}},
		{"last partial page", ports.ListRestaurantsQuery{Offset: 4, Limit: 2}, []string{"r-5"}},
		{"beyond the end", ports.ListRestaurantsQuery{Offset: 10, Limit: 2}, []string{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.List(ctx, tc.query)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(got) != len(tc.wantIDs) {
				t.Fatalf("got %d restaurants, want %d", len(got), len(tc.wantIDs))
			}
			for i, id := range tc.wantIDs {
				if got[i].ID != id {
					t.Errorf("position %d = %q, want %q", i, got[i].ID, id)
				}
			}
		})
	}
}

func TestIntegration_MenuRepository_CRUD(t *testing.T) {
	repo := NewMenuRepository(newTestDB(t))
	ctx := testCtx(t)

	item := mustNewMenuItem(t, "item-1", "rest-1", true)
	if err := repo.Save(ctx, item); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := repo.Save(ctx, mustNewMenuItem(t, "item-2", "rest-1", false)); err != nil {
		t.Fatalf("save item-2: %v", err)
	}
	if err := repo.Save(ctx, mustNewMenuItem(t, "item-x", "rest-2", true)); err != nil {
		t.Fatalf("save other restaurant item: %v", err)
	}

	t.Run("list filters by restaurant", func(t *testing.T) {
		items, err := repo.ListByRestaurantID(ctx, "rest-1")
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("got %d items, want 2 (item of rest-2 must not leak)", len(items))
		}
	})

	t.Run("get is scoped by restaurant", func(t *testing.T) {
		if _, err := repo.GetByID(ctx, "rest-2", "item-1"); !errors.Is(err, domainerrors.ErrMenuItemNotFound) {
			t.Fatalf("err = %v, want ErrMenuItemNotFound for item of another restaurant", err)
		}
	})

	t.Run("duplicate id maps to domain error", func(t *testing.T) {
		if err := repo.Save(ctx, item); !errors.Is(err, domainerrors.ErrMenuItemExists) {
			t.Fatalf("err = %v, want ErrMenuItemExists (base do replay idempotente)", err)
		}
	})

	t.Run("update persists changes", func(t *testing.T) {
		item.Available = false
		item.Price = 35
		if err := repo.Update(ctx, item); err != nil {
			t.Fatalf("update: %v", err)
		}
		got, err := repo.GetByID(ctx, "rest-1", "item-1")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Available || got.Price != 35 {
			t.Errorf("after update = %+v, want available=false price=35", got)
		}
	})
}
