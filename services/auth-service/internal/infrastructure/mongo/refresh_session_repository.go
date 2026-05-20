package mongo

import (
    "context"
    "errors"
    "time"

    "food_delivery_platform/services/auth-service/internal/domain/entities"

    "go.mongodb.org/mongo-driver/bson"
    mdriver "go.mongodb.org/mongo-driver/mongo"
)

type RefreshSessionRepository struct {
    col *mdriver.Collection
}

func NewRefreshSessionRepository(db *mdriver.Database) *RefreshSessionRepository {
    return &RefreshSessionRepository{
        col: db.Collection(refreshTokensCollection),
    }
}

func (r *RefreshSessionRepository) Create(ctx context.Context, session *entities.RefreshSession) error {
    doc := refreshSessionDoc{
        ID:                   session.ID,
        UserID:               session.UserID,
        TokenHash:            session.TokenHash,
        ExpiresAt:            session.ExpiresAt.UTC(),
        RevokedAt:            session.RevokedAt,
        RotatedFromSessionID: session.RotatedFromSessionID,
        UserAgent:            session.UserAgent,
        IPAddress:            session.IPAddress,
        CreatedAt:            session.CreatedAt.UTC(),
        UpdatedAt:            session.UpdatedAt.UTC(),
    }

    _, err := r.col.InsertOne(ctx, doc)
    return err
}

func (r *RefreshSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshSession, error) {
    var doc refreshSessionDoc
    err := r.col.FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&doc)
    if errors.Is(err, mdriver.ErrNoDocuments) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return mapRefreshDocToEntity(doc), nil
}

func (r *RefreshSessionRepository) GetByID(ctx context.Context, id string) (*entities.RefreshSession, error) {
    var doc refreshSessionDoc
    err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
    if errors.Is(err, mdriver.ErrNoDocuments) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return mapRefreshDocToEntity(doc), nil
}

func (r *RefreshSessionRepository) Revoke(ctx context.Context, sessionID string) error {
    now := time.Now().UTC()
    _, err := r.col.UpdateByID(ctx, sessionID, bson.M{
        "$set": bson.M{
            "revoked_at": now,
            "updated_at": now,
        },
    })
    return err
}

func (r *RefreshSessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
    now := time.Now().UTC()
    _, err := r.col.UpdateMany(ctx, bson.M{
        "user_id": userID,
        "revoked_at": bson.M{"$exists": false},
    }, bson.M{
        "$set": bson.M{
            "revoked_at": now,
            "updated_at": now,
        },
    })
    return err
}

func (r *RefreshSessionRepository) Update(ctx context.Context, session *entities.RefreshSession) error {
    _, err := r.col.UpdateByID(ctx, session.ID, bson.M{
        "$set": bson.M{
            "revoked_at":             session.RevokedAt,
            "rotated_from_session_id": session.RotatedFromSessionID,
            "updated_at":             session.UpdatedAt.UTC(),
        },
    })
    return err
}

func mapRefreshDocToEntity(doc refreshSessionDoc) *entities.RefreshSession {
    return &entities.RefreshSession{
        ID:                   doc.ID,
        UserID:               doc.UserID,
        TokenHash:            doc.TokenHash,
        ExpiresAt:            doc.ExpiresAt,
        RevokedAt:            doc.RevokedAt,
        RotatedFromSessionID: doc.RotatedFromSessionID,
        UserAgent:            doc.UserAgent,
        IPAddress:            doc.IPAddress,
        CreatedAt:            doc.CreatedAt,
        UpdatedAt:            doc.UpdatedAt,
    }
}