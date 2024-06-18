package patient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	patientapi "github.com/ukane-philemon/labtracka-api/cmd/patient"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	patientDatabaseDev  = "patient-dev"
	patientDatabaseProd = "patient-prod"

	/**** Patient Collections ****/

	accountCollection       = "accounts"
	loginRecordCollection   = "login-records"
	ordersCollection        = "orders"
	notificationsCollection = "notifications"
	subAccountsCollection   = "sub-accounts"

	// Fields

	dbIDKey         = "_id"
	idKey           = "id"
	deviceIDKey     = "current_device_id"
	emailKey        = "email"
	readKey         = "read"
	lastLoginKey    = "last_login"
	statusKey       = "status"
	otherAddressKey = "other_address"
	patientIDKey    = "patient_id"
	passwordKey     = "password"
	profileImageKey = "profile_image"
	subAccountsKey  = "sub_accounts"

	// Special actions
	greaterThan = "$gt"
	equalTo     = "$eq"
	setAction   = "$set"
)

// Check that MongoDB implements patientapi.Database.
var _ patientapi.Database = (*MongoDB)(nil)

// MongoDB implements patient.Database.
type MongoDB struct {
	ctx context.Context
	log *slog.Logger

	client  *mongo.Client
	patient *mongo.Database
}

// New creates and connects a new *MongoDB instance.
func New(ctx context.Context, devMode bool, logger *slog.Logger, connectionURL string) (*MongoDB, error) {
	if logger == nil || connectionURL == "" {
		return nil, errors.New("missing required patient db arguments")
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

	patientDB := patientDatabaseDev
	if !devMode {
		patientDB = patientDatabaseProd
	}

	m := &MongoDB{
		ctx:     ctx,
		client:  client,
		patient: client.Database(patientDB),
		log:     logger,
	}

	// Create unique indexes.
	accountUniqueIndex := true
	m.patient.Collection(accountCollection).Indexes().CreateOne(m.ctx, mongo.IndexModel{
		Keys: bson.D{{Key: emailKey, Value: 1}},
		Options: &options.IndexOptions{
			Unique: &accountUniqueIndex,
		},
	})
	// TODO: Create indexes for other collections.

	m.log.Info("patient database connected successfully!")

	return m, nil
}

// Shutdown ends the database connection.
func (m *MongoDB) Shutdown() {
	if err := m.client.Disconnect(m.ctx); err != nil {
		m.log.Error("User database failed to disconnect: %v", err)
	}
}

func (m *MongoDB) patientID(email string) (string, error) {
	var patient *idOnly
	accountsColl := m.patient.Collection(accountCollection)
	projection := options.FindOne().SetProjection(bson.M{dbIDKey: 1})
	err := accountsColl.FindOne(m.ctx, bson.M{emailKey: email}, projection).Decode(&patient)
	if err != nil {
		return "", err
	}

	return patient.ID.Hex(), nil
}
