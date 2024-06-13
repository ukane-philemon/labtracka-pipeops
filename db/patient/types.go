package patient

import (
	"github.com/ukane-philemon/labtracka-api/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type dbCustomer struct {
	ID primitive.ObjectID `bson:"_id"`
	*db.CustomerInfo
	ProfileImage       string `bson:"profile_image"`
	DeviceID           string `bson:"device_id"`
	Password           string
	CreatedAtTimestamp int64 `bson:"created_at_timestamp"`
}
