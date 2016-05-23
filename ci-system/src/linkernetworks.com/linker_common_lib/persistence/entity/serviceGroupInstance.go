package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type InstanceLock struct {
	ObjectId        bson.ObjectId `bson:"_id" json:"-"`
	GroupInstanceId string        `bson:"instance_id" json:"instance_id"`
}

type RefinedApp struct {
	Id          string   `bson:"id" json:"id"`
	Cpus        float32  `bson:"cpus" json:"cpus"`
	Mem         int16    `bson:"mem" json:"mem"`
	Instances   int      `bson:"instances" json:"instances"`
	InstanceIds []string `bson:"instance_ids" json:"instance_ids"`
}

type RefinedGroup struct {
	Id            string         `bson:"id" json:"id"`
	Dependencies  []string       `bson:"dependencies" json:"dependencies"`
	RefinedApps   []RefinedApp   `bson:"apps" json:"apps"`
	RefinedGroups []RefinedGroup `bson:"groups" json:"groups"`
}

type ServiceGroupInstance struct {
	ObjectId        bson.ObjectId  `bson:"_id" json:"_id"`
	ServiceGroupId  string         `bson:"service_group_id" json:"service_group_id"`
	Version         float32        `bson:"version" json:"version"`
	Groups          []RefinedGroup `bson:"groups" json:"groups"`
	LifeCycleStatus string         `bson:"life_cycle_status" json:"life_cycle_status"`
	User_id         string         `bson:"user_id" json:"user_id"`
	Tenant_id       string         `bson:"tenant_id" json:"tenant_id"`
	TimeCreate      string         `bson:"time_create" json:"time_create"`
	TimeUpdate      string         `bson:"time_update" json:"time_update"`
	RepairId        string         `bson:"repair_id" json:"repair_id"`
}
