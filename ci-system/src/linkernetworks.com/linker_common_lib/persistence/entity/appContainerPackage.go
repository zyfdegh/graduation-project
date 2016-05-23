package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type Notify struct {
	NotifyPath string `bson:"notify_path" json:"notify_path"`
	Scope      string `bson:"scope" json:"scope"`
}

type ConfigStep struct {
	ConfigType string `bson:"config_type" json:"config_type"`
	Execute    string `bson:"execute" json:"execute"`
	Scope      string `bson:"scope" json:"scope"`
}

type PreCondition struct {
	Condition string `bson:"condition" json:"condition"`
}

type Configuration struct {
	Name          string         `bson:"name" json:"name"`
	Preconditions []PreCondition `bson:"preconditions" json:"preconditions"`
	Steps         []ConfigStep   `bson:"steps" json:"steps"`
}

type AppContainerPackage struct {
	ObjectId       bson.ObjectId   `bson:"_id" json:"_id"`
	AppContainerId string          `bson:"app_container_id" json:"app_container_id"`
	Version        float32         `bson:"version" json:"version"`
	ServiceGroupId string          `bson:"service_group_id" json:"service_group_id"`
	Image          string          `bson:"image" json:"image"`
	Configurations []Configuration `bson:"configurations" json:"configurations"`
	Notifies       []Notify        `bson:"notifies" json:"notifies"`
	User_id        string          `bson:"user_id" json:"user_id"`
	Tenant_id      string          `bson:"tenant_id" json:"tenant_id"`
	TimeCreate     string          `bson:"time_create" json:"time_create"`
	TimeUpdate     string          `bson:"time_update" json:"time_update"`
}
