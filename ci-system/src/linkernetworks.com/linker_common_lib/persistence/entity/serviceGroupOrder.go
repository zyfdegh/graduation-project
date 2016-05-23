package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type ServiceGroupOrder struct {
	ObjectId               bson.ObjectId       `bson:"_id"  json:"_id"`
	OrderName              string              `bson:"name" json:"name"`
	ClusterId              string              `bson:"cluster_id" json:"cluster_id"`
	ClusterName            string              `bson:"cluster_name" json:"cluster_name"`
	Description            string              `bson:"description" json:"description"`
	ServiceGroupObjId      string              `bson:"service_group_obj_id" json:"service_group_obj_id"`
	ServiceGroupId         string              `bson:"service_group_id" json:"service_group_id"`
	ServiceGroupInstanceId string              `bson:"service_group_instance_id" json:"service_group_instance_id"`
	ServiceOfferingId      string              `bson:"service_offering_id" json:"service_offering_id"`
	MarathonGroupId        string              `bson:"marathon_group_id" json:"marathon_group_id"`
	LifeCycleStatus        string              `bson:"life_cycle_status" json:"life_cycle_status"`
	OfferingParameters     []OfferingParameter `bson:"parameters,omitempty" json:"parameters,omitempty"`
	User_id                string              `bson:"user_id" json:"user_id"`
	Tenant_id              string              `bson:"tenant_id" json:"tenant_id"`
	TimeCreate             string              `bson:"time_create" json:"time_create"`
	TimeUpdate             string              `bson:"time_update" json:"time_update"`
	TerminatedDate         string              `bson:"terminatedDate" json:"terminatedDate"`
}

type OfferingParameter struct {
	AppId      string `bson:"appId" json:"appId"`
	ParamName  string `bson:"paramName" json:"paramName"`
	ParamValue string `bson:"paramValue" json:"paramValue"`
}
