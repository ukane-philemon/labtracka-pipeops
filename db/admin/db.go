package admin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ukane-philemon/labtracka-api/cmd/admin"
	"github.com/ukane-philemon/labtracka-api/cmd/patient"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	adminDatabaseDev  = "admin-dev"
	adminDatabaseProd = "admin-prod"
)

var _ admin.Database = (*MongoDB)(nil)
var _ patient.AdminDatabase = (*MongoDB)(nil)

// MongoDB implements the adminapi.Database and patientapi.AdminDatabase
// interface.
type MongoDB struct {
	ctx context.Context
	log *slog.Logger

	client *mongo.Client
	admin  *mongo.Database
}

// New connects and creates a *MongoDB instance.
func New(ctx context.Context, devMode bool, logger *slog.Logger, connectionURL string) (*MongoDB, error) {
	if logger == nil || connectionURL == "" {
		return nil, errors.New("missing required admin server arguments")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connectionURL).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("mongo.Connect error: %w", err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Ping(ctx, readpref.Nearest()); err != nil {
		return nil, fmt.Errorf("client.Ping error: %w", err)
	}

	adminDatabase := adminDatabaseDev
	if !devMode {
		adminDatabase = adminDatabaseProd
	}
	m := &MongoDB{
		ctx:    ctx,
		client: client,
		admin:  client.Database(adminDatabase),
		log:    logger,
	}

	// Create unique indexes.
	// accountUniqueIndex := true
	// TODO: Create indexes.
	// m.admin.Collection(accountCollection).Indexes().CreateOne(m.ctx, mongo.IndexModel{
	// 	Keys: bson.D{{Key: emailKey, Value: 1}, {Key: idKey, Value: 1}},
	// 	Options: &options.IndexOptions{
	// 		Unique: &accountUniqueIndex,
	// 	},
	// })
	// TODO: Create indexes for other collections.

	m.log.Info("Admin database connected successfully!")

	return m, nil
}

// Shutdown ends the database connection.
func (m *MongoDB) Shutdown() {
	if err := m.client.Disconnect(m.ctx); err != nil {
		m.log.Error("User database failed to disconnect: %v", err)
	}
}
