package patient

import (
	"github.com/ukane-philemon/labtracka-api/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type idOnly struct {
	ID primitive.ObjectID `bson:"_id"`
}

type dbPatient struct {
	ID                 primitive.ObjectID `bson:"_id"`
	*db.PatientInfo    `bson:"inline"`
	SubAccounts        []string `json:"sub_accounts" bson:"sub_accounts"`
	ProfileImage       string   `bson:"profile_image"`
	DeviceID           string   `bson:"device_id"`
	Password           string
	CreatedAtTimestamp int64 `bson:"created_at_timestamp"`
}

type subAccountRecord struct {
	PatientID   string                        `bson:"patient_id"`
	SubAccounts map[string]*db.SubAccountInfo `bson:"sub_accounts"`
}
