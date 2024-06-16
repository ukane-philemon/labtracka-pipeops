package patient

import (
	"errors"
	"fmt"
	"time"

	"github.com/ukane-philemon/labtracka-api/db"
	"github.com/ukane-philemon/labtracka-api/internal/funcs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PatientOrders returns a list of orders made by the patient with the
// provided email.
func (m *MongoDB) PatientOrders(email string) ([]*db.Order, error) {
	return nil, nil
}

// CreatePatientOrder creates a new order for the patient and returns the
// orderID and amount after validating the order.
func (m *MongoDB) CreatePatientOrder(email string, orderReq *db.OrderInfo) (string, error) {
	orderID, err := funcs.RandomToken(5)
	if err != nil {
		return "", fmt.Errorf("funcs.RandomToken error: %w", err)
	}

	order := &db.Order{
		ID:        orderID,
		OrderInfo: *orderReq,
		Timestamp: time.Now().Unix(),
	}

	accountsColl := m.patient.Collection(accountCollection)
	nPatient, err := accountsColl.CountDocuments(m.ctx, bson.M{emailKey: email})
	if err != nil {
		return "", err
	}

	if nPatient == 0 {
		return "", errors.New("patient does not exist")
	}

	ordersCollection := m.patient.Collection(ordersCollection)
	res, err := ordersCollection.InsertOne(m.ctx, order)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// UpdatePatientOrder updates the status for a patient order.
func (m *MongoDB) UpdatePatientOrder(email string, orderIDStr, status string) error {
	var patient *dbPatient
	accountsColl := m.patient.Collection(accountCollection)
	err := accountsColl.FindOne(m.ctx, bson.M{emailKey: email}).Decode(&patient)
	if err != nil {
		return err
	}

	orderID, err := primitive.ObjectIDFromHex(orderIDStr)
	if err != nil {
		return fmt.Errorf("primitive.ObjectIDFromHex error: %w", err)
	}

	ordersCollection := m.patient.Collection(ordersCollection)
	orderFilter := bson.M{dbIDKey: orderID, patientIDKey: patient.ID.Hex()}
	res, err := ordersCollection.UpdateOne(m.ctx, orderFilter, bson.M{statusKey: db.OrderStatusPaid}, options.Update().SetUpsert(false))
	if err != nil {
		return fmt.Errorf("ordersCollection.UpdateOne error: %w", err)
	}

	if res.ModifiedCount == 0 {
		return fmt.Errorf("patient order with id %s does not exist", orderIDStr)
	}

	return nil
}
