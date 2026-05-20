package mongo

import (
    "context"
    "errors"

    domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
    "food_delivery_platform/services/auth-service/internal/domain/entities"
    "food_delivery_platform/services/auth-service/internal/domain/valueobjects"

    "go.mongodb.org/mongo-driver/bson"
    mdriver "go.mongodb.org/mongo-driver/mongo"
)

type CredentialRepository struct {
    col *mdriver.Collection
}

func NewCredentialRepository(db *mdriver.Database) *CredentialRepository {
    return &CredentialRepository{
        col: db.Collection(credentialsCollection),
    }
}

func (r *CredentialRepository) Create(ctx context.Context, credential *entities.Credential) error {
    doc := credentialDoc{
        ID:                  credential.ID,
        UserID:              credential.UserID,
        Email:               credential.Email.String(),
        PasswordHash:        credential.PasswordHash,
        Status:              string(credential.Status),
        FailedLoginAttempts: credential.FailedLoginAttempts,
        LastLoginAt:         credential.LastLoginAt,
        CreatedAt:           credential.CreatedAt.UTC(),
        UpdatedAt:           credential.UpdatedAt.UTC(),
    }

    _, err := r.col.InsertOne(ctx, doc)
    if isDuplicateKeyError(err) {
        return domainerrors.ErrEmailAlreadyRegistered
    }
    return err
}

func (r *CredentialRepository) GetByEmail(ctx context.Context, email string) (*entities.Credential, error) {
    var doc credentialDoc
    err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
    if errors.Is(err, mdriver.ErrNoDocuments) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return mapCredentialDocToEntity(doc)
}

func (r *CredentialRepository) GetByUserID(ctx context.Context, userID string) (*entities.Credential, error) {
    var doc credentialDoc
    err := r.col.FindOne(ctx, bson.M{"user_id": userID}).Decode(&doc)
    if errors.Is(err, mdriver.ErrNoDocuments) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return mapCredentialDocToEntity(doc)
}

func (r *CredentialRepository) Update(ctx context.Context, credential *entities.Credential) error {
    update := bson.M{
        "$set": bson.M{
            "password_hash":          credential.PasswordHash,
            "status":                 string(credential.Status),
            "failed_login_attempts":  credential.FailedLoginAttempts,
            "last_login_at":          credential.LastLoginAt,
            "updated_at":             credential.UpdatedAt.UTC(),
        },
    }

    _, err := r.col.UpdateByID(ctx, credential.ID, update)
    return err
}

func mapCredentialDocToEntity(doc credentialDoc) (*entities.Credential, error) {
    emailVO, err := valueobjects.NewEmail(doc.Email)
    if err != nil {
        return nil, err
    }

    status, err := valueobjects.NewCredentialStatus(doc.Status)
    if err != nil {
        return nil, err
    }

    return &entities.Credential{
        ID:                  doc.ID,
        UserID:              doc.UserID,
        Email:               emailVO,
        PasswordHash:        doc.PasswordHash,
        Status:              status,
        FailedLoginAttempts: doc.FailedLoginAttempts,
        LastLoginAt:         doc.LastLoginAt,
        CreatedAt:           doc.CreatedAt,
        UpdatedAt:           doc.UpdatedAt,
    }, nil
}