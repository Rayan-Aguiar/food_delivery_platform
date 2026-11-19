package application

import (
	"context"
	"errors"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/restaurant-service/internal/domain/entities"
	domainerrors "food_delivery_platform/services/restaurant-service/internal/domain/errors"
	"food_delivery_platform/services/restaurant-service/internal/domain/ports"
)

type CreateMenuItemInput struct {
	RestaurantID string
	Name         string
	Price        float64
	Category     string
	Available    bool

	// IdempotencyKey, quando informado, e usado como ID do item: reenviar a
	// mesma chave devolve o item ja criado (sem sobrescrever nem republicar evento).
	IdempotencyKey string

	CorrelationID string
	CausationID   string
	Traceparent   string
}

type CreateMenuItemUseCase struct {
	restaurants ports.RestaurantRepository
	menu        ports.MenuRepository
	idGen       ports.IDGenerator
	events      ports.CatalogEventPublisher
}

func NewCreateMenuItemUseCase(
	restaurants ports.RestaurantRepository,
	menu ports.MenuRepository,
	idGen ports.IDGenerator,
	eventPublisher ...ports.CatalogEventPublisher,
) *CreateMenuItemUseCase {
	publisher := ports.CatalogEventPublisher(noopCatalogEventPublisher{})
	if len(eventPublisher) > 0 && eventPublisher[0] != nil {
		publisher = eventPublisher[0]
	}

	return &CreateMenuItemUseCase{restaurants: restaurants, menu: menu, idGen: idGen, events: publisher}
}

func (uc *CreateMenuItemUseCase) Execute(ctx context.Context, input CreateMenuItemInput) (entities.MenuItem, error) {
	if input.RestaurantID == "" {
		return entities.MenuItem{}, apperrors.InvalidArgument("restaurant id is required", nil)
	}

	if _, err := uc.restaurants.GetByID(ctx, input.RestaurantID); err != nil {
		return entities.MenuItem{}, mapNotFoundOrInternal(err, domainerrors.ErrRestaurantNotFound, "restaurant not found", "failed to get restaurant")
	}

	id := input.IdempotencyKey
	if id == "" {
		id = uc.idGen.NewID()
	}

	item, err := entities.NewMenuItem(id, input.RestaurantID, input.Name, input.Price, entities.MenuCategory(input.Category), input.Available)
	if err != nil {
		return entities.MenuItem{}, apperrors.InvalidArgument(err.Error(), err)
	}

	if err := uc.menu.Save(ctx, item); err != nil {
		if errors.Is(err, domainerrors.ErrMenuItemExists) && input.IdempotencyKey != "" {
			return uc.replay(ctx, input.RestaurantID, id)
		}
		return entities.MenuItem{}, apperrors.Internal("failed to save menu item", err)
	}

	// Best effort: a criacao do item continua bem-sucedida mesmo com falha de publicacao.
	_ = uc.events.PublishMenuUpdated(ctx, ports.MenuUpdatedEvent{
		RestaurantID:   item.RestaurantID,
		ItemID:         item.ID,
		Name:           item.Name,
		Price:          item.Price,
		Category:       string(item.Category),
		CorrelationID:  input.CorrelationID,
		CausationID:    input.CausationID,
		Traceparent:    input.Traceparent,
		IdempotencyKey: input.IdempotencyKey,
	})

	return item, nil
}

// replay atende o retry de uma criacao ja concluida: insert + lookup e' atomico o
// bastante aqui porque o _id unico do Mongo garante um unico vencedor sob
// concorrencia; quem perde a corrida le o documento do vencedor.
func (uc *CreateMenuItemUseCase) replay(ctx context.Context, restaurantID, id string) (entities.MenuItem, error) {
	existing, err := uc.menu.GetByID(ctx, restaurantID, id)
	if errors.Is(err, domainerrors.ErrMenuItemNotFound) {
		// A chave ja foi usada para um item de outro restaurante.
		return entities.MenuItem{}, apperrors.Conflict("idempotency key already used for another resource", nil)
	}
	if err != nil {
		return entities.MenuItem{}, apperrors.Internal("failed to get menu item", err)
	}
	return existing, nil
}
