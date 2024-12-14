package testhelpers

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBContainer struct {
	*mongodb.MongoDBContainer
	ConnectionString string
}

// CreateMongoDBContainer creates a MongoDB testcontainer.
func CreateMongoDBContainer(ctx context.Context) (*MongoDBContainer, error) {
	mdbContainer, err := mongodb.Run(ctx,
		"mongo:6",
		mongodb.WithReplicaSet("rs0"),
		testcontainers.WithWaitStrategy(wait.ForLog("Waiting for connections")),
	)

	if err != nil {
		return nil, err
	}

	connStr, err := mdbContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return &MongoDBContainer{
		MongoDBContainer: mdbContainer,
		ConnectionString: connStr,
	}, nil

}

// Client returns a MongoDB client.
func (m MongoDBContainer) Client(ctx context.Context) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(m.ConnectionString).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
