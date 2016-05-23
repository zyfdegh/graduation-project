package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type Quotas struct {
	ObjectId        bson.ObjectId `bson:"_id" json:"_id"`
	Cpus            float64       `bson:"cpus" json:"cpus"`
	Memory          float64       `bson:"memory" json:"memory"`
	Disk            float64       `bson:"disk" json:"disk"`
	VolumnGigabytes float64       `bson:"volumngigabytes" json:"volumngigabytes"`
	Orders          float64       `bson:"orders" json:"orders"`
	Instances       float64       `bson:"instances" json:"instances"`
	User_id         string        `bson:"user_id" json:"user_id"`
	Tenant_id       string        `bson:"tenant_id" json:"tenant_id"`
	TimeCreate      string        `bson:"time_create" json:"time_create"`
	TimeUpdate      string        `bson:"time_update" json:"time_update"`
}

type QuotaUsages struct {
	ObjectId   bson.ObjectId `bson:"_id" json:"_id"`
	Resource   string        `bson:"resource" json:"resource"`
	In_use     float64       `bson:"in_use" json:"in_use"`
	User_id    string        `bson:"user_id" json:"user_id"`
	Tenant_id  string        `bson:"tenant_id" json:"tenant_id"`
	TimeCreate string        `bson:"time_create" json:"time_create"`
	TimeUpdate string        `bson:"time_update" json:"time_update"`
}
