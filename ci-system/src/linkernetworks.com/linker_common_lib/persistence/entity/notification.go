package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type Notification struct {
	ObjectId        bson.ObjectId `bson:"_id" json:"_id"`
	ServiceGroupId  string        `bson:"service_group_id" json:"service_group_id"`
	NotifyAddress    string         `bson:"notify_address" json:"notify_address"`
	TimeCreate      string        `bson:"time_create" json:"time_create"`
	TimeUpdate    string        `bson:"time_update" json:"time_update"`
}


