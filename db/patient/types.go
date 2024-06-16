package patient

import (
	"github.com/ukane-philemon/labtracka-api/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type dbPatient struct {
	ID                 primitive.ObjectID `bson:"_id"`
	*db.PatientInfo    `bson:"inline"`
	SubAccounts        []string `json:"sub_accounts" bson:"sub_accounts"`
	ProfileImage       string   `bson:"profile_image"`
	DeviceID           string   `bson:"device_id"`
	Password           string
	CreatedAtTimestamp int64 `bson:"created_at_timestamp"`
}
