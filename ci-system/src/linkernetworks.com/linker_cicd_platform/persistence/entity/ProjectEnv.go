package entity

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	PROJECTENV_STATUS_TERMINATED = "terminated"
)

type ProjectEnv struct {
	ObjectId                  bson.ObjectId `bson:"_id" json:"-"`
	Id                        string        `bson:"id" json:"id"`
	OpsEnvId                  string        `bson:"ops_env_id" json:"ops_env_id"`
	ProjectId                 string        `bson:"project_id" json:"project_id"`
	JobId                     string        `bson:"job_id" json:"job_id"`
	ServiceGroupInstanceId    string        `bson:"service_group_instance_id" json:"service_group_instance_id"`
	ServiceOrderId            string        `bson:"service_order_id" json:"service_order_id"`
	ServiceOfferingInstanceId string        `bson:"service_offering_instance_id" json:"service_offering_instance_id"`
	UserId                    string        `bson:"user_id" json:"user_id"`
	TenantId                  string        `bson:"tenant_id" json:"tenant_id"`
	Status                    string        `bson:"status" json:"status"`
}
