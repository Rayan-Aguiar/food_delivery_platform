package application

import (
	"context"
	"errors"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/domain/entities"
	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type RegisterUserInput struct {
	Email    string
	Password string
}

type RegisterUserOutput struct {
	UserID       string
	CredentialID string
}

type RegisterUserUseCase struct {
	credentials    ports.CredentialRepository
	hasher         ports.PasswordHasher
	clock          ports.Clock
	idGen          ports.IDGenerator
	passwordPolicy valueobjects.PasswordPolicy
}

func NewRegisterUserUseCase(
	credentials ports.CredentialRepository,
	hasher ports.PasswordHasher,
	clock ports.Clock,
	idGen ports.IDGenerator,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		credentials:    credentials,
		hasher:         hasher,
		clock:          clock,
		idGen:          idGen,
		passwordPolicy: valueobjects.NewDefaultPasswordPolicy(),
	}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (RegisterUserOutput, error) {
	email, err := valueobjects.NewEmail(input.Email)
	if err != nil {
		return RegisterUserOutput{}, apperrors.InvalidArgument("invalid email", err)
	}

	if err := uc.passwordPolicy.Validate(input.Password); err != nil {
		return RegisterUserOutput{}, apperrors.InvalidArgument("password does not meet requirements", err)
	}

	existing, err := uc.credentials.GetByEmail(ctx, email.String())
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to check email uniqueness", err)
	}
	if existing != nil {
		return RegisterUserOutput{}, apperrors.Conflict("email already registered", nil)
	}

	hash, err := uc.hasher.Hash(ctx, input.Password)
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to hash password", err)
	}

	now := uc.clock.Now()
	credentialID := uc.idGen.NewID()
	userID := uc.idGen.NewID()

	cred, err := entities.NewCredential(credentialID, userID, email, hash, now)
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to build credential", err)
	}

	if err := uc.credentials.Create(ctx, cred); err != nil {
		if errors.Is(err, domainerrors.ErrEmailAlreadyRegistered) {
			return RegisterUserOutput{}, apperrors.Conflict("email already registered", nil)
		}
		return RegisterUserOutput{}, apperrors.Internal("failed to persist credential", err)
	}

	return RegisterUserOutput{
		UserID:       userID,
		CredentialID: credentialID,
	}, nil
}
