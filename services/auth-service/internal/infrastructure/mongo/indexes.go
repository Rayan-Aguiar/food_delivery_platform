package mongo

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    mdriver "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

const (
    credentialsCollection = "credentials"
    refreshTokensCollection = "refresh_tokens"
)

func EnsureIndexes(ctx context.Context, db *mdriver.Database) error {
    credentials := db.Collection(credentialsCollection)
    refresh := db.Collection(refreshTokensCollection)

    _, err := credentials.Indexes().CreateMany(ctx, []mdriver.IndexModel{
        {
            Keys:    bson.D{{Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true).SetName("ux_credentials_email"),
        },
        {
            Keys:    bson.D{{Key: "user_id", Value: 1}},
            Options: options.Index().SetName("ix_credentials_user_id"),
        },
    })
    if err != nil {
        return err
    }

    _, err = refresh.Indexes().CreateMany(ctx, []mdriver.IndexModel{
        {
            Keys:    bson.D{{Key: "token_hash", Value: 1}},
            Options: options.Index().SetUnique(true).SetName("ux_refresh_token_hash"),
        },
        {
            Keys:    bson.D{{Key: "user_id", Value: 1}},
            Options: options.Index().SetName("ix_refresh_user_id"),
        },
        {
            Keys:    bson.D{{Key: "expires_at", Value: 1}},
            Options: options.Index().SetExpireAfterSeconds(0).SetName("ttl_refresh_expires_at"),
        },
    })
    return err
}