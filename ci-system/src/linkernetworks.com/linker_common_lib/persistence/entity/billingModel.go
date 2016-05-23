package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type BillingModel struct {
	ObjectId    bson.ObjectId `bson:"_id" json:"_id"`
	TotalPrice  float64       `bson:"totalprice" json:"totalprice"`
	Price       float64       `bson:"price" json:"price"`
	ModelId     string        `bson:"modelid" json:"modelid"`
	User_id     string        `bson:"user_id" json:"user_id"`
	Tenant_id   string        `bson:"tenant_id" json:"tenant_id"`
	Refs        []string      `bson:"refs" json:"refs"`
	Description string        `bson:"description" json:"description"`
	TimeCreate  string        `bson:"time_create" json:"time_create"`
	TimeUpdate  string        `bson:"time_update" json:"time_update"`
}

type UserAccount struct {
	ObjectId           bson.ObjectId `bson:"_id" json:"_id"`
	Username           string        `bson:"username" json:"username"`
	User_id            string        `bson:"user_id" json:"user_id"`
	Tenant_id          string        `bson:"tenant_id" json:"tenant_id"`
	Ref_userid         string        `bson:"ref_userid" json:"ref_userid"`
	Ref_tenantid       string        `bson:"ref_tenantid" json:"ref_tenantid"`
	Transaction_type   string        `bson:"transaction_type" json:"transaction_type"`
	Transaction_desc   string        `bson:"transaction_desc" json:"transaction_desc"`
	Transaction_object string        `bson:"transaction_object" json:"transaction_object"`
	Transaction_status string        `bson:"transaction_status" json:"transaction_status"`
	Sg_id              string        `bson:"sg_id" json:"sg_id"`
	Sg_objId           string        `bson:"sg_objid" json:"sg_objid"`
	Price              float64       `bson:"price" json:"price"`
	Date               int           `bson:"date" json:"date"`
	TimeCreate         string        `bson:"time_create" json:"time_create"`
	TimeUpdate         string        `bson:"time_update" json:"time_update"`
}

type UserAccountBalance struct {
	ObjectId    bson.ObjectId `bson:"_id" json:"_id"`
	User_id     string        `bson:"user_id" json:"user_id"`
	Tenant_id   string        `bson:"tenant_id" json:"tenant_id"`
	Balance     float64       `bson:"balance" json:"balance"`
	Consume     float64       `bson:"consume" json:"consume"`
	Income      float64       `bson:"income" json:"income"`
	NotifyLevel int           `bson:"notifylevel" json:"notifylevel"`
	TimeCreate  string        `bson:"time_create" json:"time_create"`
	TimeUpdate  string        `bson:"time_update" json:"time_update"`
}
