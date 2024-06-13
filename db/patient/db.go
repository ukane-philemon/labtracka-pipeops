package patient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ukane-philemon/labtracka-api/cmd/patient"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	customerDatabase = "customer"

	/**** Customer Collections ****/

	accountCollection     = "accounts"
	loginRecordCollection = "login-records"

	// Fields

	dbIDKey         = "_id"
	idKey           = "id"
	deviceIDKey     = "current_device_id"
	emailKey        = "email"
	lastLoginKey    = "last_login"
	statusKey       = "status"
	customerInfoKey = "customerinfo" // embedded
	otherAddressKey = "other_address"

	customerResultsKey = "customer_results"
	customerIDKey      = "customer_id"

	// Special actions
	greaterThan = "$gt"
	equalTo     = "$eq"
	setAction   = "$set"
)

// Check that MongoDB implements patient.Database.
var _ patient.Database = (*MongoDB)(nil)

// MongoDB implements patient.Database.
type MongoDB struct {
	ctx context.Context
	log *slog.Logger

	client   *mongo.Client
	customer *mongo.Database
}

// New creates and connects a new *MongoDB instance.
func New(ctx context.Context, logger *slog.Logger, connectionURL string) (*MongoDB, error) {
	if logger == nil || connectionURL == "" {
		return nil, errors.New("missing required arguments")
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

	m := &MongoDB{
		ctx:      ctx,
		client:   client,
		customer: client.Database(customerDatabase),
		log:      logger,
	}

	// Create unique indexes.
	accountUniqueIndex := true
	m.customer.Collection(accountCollection).Indexes().CreateOne(m.ctx, mongo.IndexModel{
		Keys: bson.D{{Key: emailKey, Value: 1}, {Key: idKey, Value: 1}},
		Options: &options.IndexOptions{
			Unique: &accountUniqueIndex,
		},
	})
	// TODO: Create indexes for other collections.

	m.log.Info("Customer database connected successfully!")

	return m, nil
}

// Shutdown ends the database connection.
func (m *MongoDB) Shutdown() {
	if err := m.client.Disconnect(m.ctx); err != nil {
		m.log.Error("User database failed to disconnect: %v", err)
	}
}
