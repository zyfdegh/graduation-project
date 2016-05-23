package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type IpAddressResource struct {
	ObjectId   bson.ObjectId `bson:"_id,omitempty" json:"_id,omitempty"`
	IpAddress  string        `bson:"ipAddress" json:"ipAddress"`
	Subnet     string        `bson:"subnet" json:"subnet"`
	Gateway    string        `bson:"gateway" json:"gateway"`
	Allocated  string        `bson:"allocated" json:"allocated"`
	PoolName   string        `bson:"pool_name" json:"pool_name"`
	User_id    string        `bson:"user_id" json:"user_id"`
	Tenant_id  string        `bson:"tenant_id" json:"tenant_id"`
	TimeCreate string        `bson:"time_create" json:"time_create"`
	TimeUpdate string        `bson:"time_update" json:"time_update"`
}

type IpAddressPool struct {
	ObjectId    bson.ObjectId       `bson:"_id" json:"_id"`
	PoolName    string              `bson:"pool_name" json:"pool_name"`
	Subnet      string              `bson:"subnet" json:"subnet"`
	Gateway     string              `bson:"gateway" json:"gateway"`
	IpResources []IpAddressResource `bson:"ip_resources" json:"ip_resources"`
}
