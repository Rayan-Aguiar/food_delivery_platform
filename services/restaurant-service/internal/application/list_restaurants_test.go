package application_test

import (
	"context"
	"errors"
	"testing"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/application"
	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

// repoReturning devolve n restaurantes e registra a query recebida.
func repoReturning(n int, gotQuery *ports.ListRestaurantsQuery) *fakeRestaurantRepo {
	return &fakeRestaurantRepo{
		listFn: func(_ context.Context, q ports.ListRestaurantsQuery) ([]entities.Restaurant, error) {
			if gotQuery != nil {
				*gotQuery = q
			}
			out := make([]entities.Restaurant, n)
			for i := range out {
				out[i] = mustRestaurant()
			}
			return out, nil
		},
	}
}

func TestListRestaurantsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name         string
		input        application.ListRestaurantsInput
		repoReturns  int
		wantCount    int
		wantPage     int
		wantPageSize int
		wantHasMore  bool
		wantQuery    ports.ListRestaurantsQuery
	}{
		{
			name:         "defaults when page and page_size are omitted",
			input:        application.ListRestaurantsInput{},
			repoReturns:  3,
			wantCount:    3,
			wantPage:     1,
			wantPageSize: application.DefaultRestaurantsPageSize,
			wantQuery:    ports.ListRestaurantsQuery{Offset: 0, Limit: application.DefaultRestaurantsPageSize + 1},
		},
		{
			name:         "empty list",
			input:        application.ListRestaurantsInput{},
			repoReturns:  0,
			wantCount:    0,
			wantPage:     1,
			wantPageSize: application.DefaultRestaurantsPageSize,
			wantQuery:    ports.ListRestaurantsQuery{Offset: 0, Limit: application.DefaultRestaurantsPageSize + 1},
		},
		{
			name:         "has_more when repository returns one extra record",
			input:        application.ListRestaurantsInput{Page: 1, PageSize: 2},
			repoReturns:  3,
			wantCount:    2,
			wantPage:     1,
			wantPageSize: 2,
			wantHasMore:  true,
			wantQuery:    ports.ListRestaurantsQuery{Offset: 0, Limit: 3},
		},
		{
			name:         "exact page size means no next page",
			input:        application.ListRestaurantsInput{Page: 1, PageSize: 2},
			repoReturns:  2,
			wantCount:    2,
			wantPage:     1,
			wantPageSize: 2,
			wantQuery:    ports.ListRestaurantsQuery{Offset: 0, Limit: 3},
		},
		{
			name:         "offset is derived from page",
			input:        application.ListRestaurantsInput{Page: 3, PageSize: 10},
			repoReturns:  1,
			wantCount:    1,
			wantPage:     3,
			wantPageSize: 10,
			wantQuery:    ports.ListRestaurantsQuery{Offset: 20, Limit: 11},
		},
		{
			name:         "max limits are accepted",
			input:        application.ListRestaurantsInput{Page: application.MaxRestaurantsPage, PageSize: application.MaxRestaurantsPageSize},
			wantPage:     application.MaxRestaurantsPage,
			wantPageSize: application.MaxRestaurantsPageSize,
			wantQuery: ports.ListRestaurantsQuery{
				Offset: (application.MaxRestaurantsPage - 1) * application.MaxRestaurantsPageSize,
				Limit:  application.MaxRestaurantsPageSize + 1,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotQuery ports.ListRestaurantsQuery
			uc := application.NewListRestaurantsUseCase(repoReturning(tc.repoReturns, &gotQuery))

			got, err := uc.Execute(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got.Restaurants) != tc.wantCount {
				t.Errorf("restaurants = %d, want %d", len(got.Restaurants), tc.wantCount)
			}
			if got.Page != tc.wantPage || got.PageSize != tc.wantPageSize {
				t.Errorf("page/page_size = %d/%d, want %d/%d", got.Page, got.PageSize, tc.wantPage, tc.wantPageSize)
			}
			if got.HasMore != tc.wantHasMore {
				t.Errorf("has_more = %v, want %v", got.HasMore, tc.wantHasMore)
			}
			if gotQuery != tc.wantQuery {
				t.Errorf("repository query = %+v, want %+v", gotQuery, tc.wantQuery)
			}
		})
	}
}

func TestListRestaurantsUseCase_Execute_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input application.ListRestaurantsInput
	}{
		{"negative page", application.ListRestaurantsInput{Page: -1}},
		{"page above max", application.ListRestaurantsInput{Page: application.MaxRestaurantsPage + 1}},
		{"negative page_size", application.ListRestaurantsInput{PageSize: -5}},
		{"page_size above max", application.ListRestaurantsInput{PageSize: application.MaxRestaurantsPageSize + 1}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			repo := &fakeRestaurantRepo{
				listFn: func(context.Context, ports.ListRestaurantsQuery) ([]entities.Restaurant, error) {
					called = true
					return nil, nil
				},
			}

			_, err := application.NewListRestaurantsUseCase(repo).Execute(context.Background(), tc.input)

			var appErr *apperrors.AppError
			if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidArgument {
				t.Fatalf("expected invalid argument error, got %v", err)
			}
			if called {
				t.Error("repository must not be called for invalid input")
			}
		})
	}
}

func TestListRestaurantsUseCase_Execute_RepositoryError(t *testing.T) {
	repo := &fakeRestaurantRepo{
		listFn: func(context.Context, ports.ListRestaurantsQuery) ([]entities.Restaurant, error) {
			return nil, errors.New("db unavailable")
		},
	}

	_, err := application.NewListRestaurantsUseCase(repo).Execute(context.Background(), application.ListRestaurantsInput{})

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInternal {
		t.Fatalf("expected internal error, got %v", err)
	}
}
