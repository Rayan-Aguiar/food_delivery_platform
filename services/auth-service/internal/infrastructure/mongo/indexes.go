package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	mdriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	credentialsCollection   = "credentials"
	refreshTokensCollection = "refresh_tokens"
	outboxCollection        = "auth_outbox"
)

func EnsureIndexes(ctx context.Context, db *mdriver.Database) error {
	credentials := db.Collection(credentialsCollection)
	refresh := db.Collection(refreshTokensCollection)
	outbox := db.Collection(outboxCollection)

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
	if err != nil {
		return err
	}

	_, err = outbox.Indexes().CreateMany(ctx, []mdriver.IndexModel{
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetName("ix_outbox_created_at"),
		},
		{
			Keys:    bson.D{{Key: "routing_key", Value: 1}},
			Options: options.Index().SetName("ix_outbox_routing_key"),
		},
	})
	return err
}
