package entity

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	JOB_STATUS_TERMINATED = "terminated"
)

type Job struct {
	ObjectId    bson.ObjectId `bson:"_id" json:"-"`
	Id          string        `bson:"id" json:"id"`
	Name        string        `bson:"name" json:"name"`
	GitUrl      string        `bson:"giturl" json:"giturl"`
	Project     string        `bson:"project" json:"project"`
	BuildNumber int           `bson:"buildnumber" json:"buildnumber"`
	Branch      string        `bson:"branch" json:"branch"`
	Status      string        `bson:"status" json:"status"`
	Version     string        `bson:"version" json:"version"`
	AutoDeploy  bool          `bson:"autodeploy" json:"autodeploy"`
}

type JobEnv struct {
	ObjectId                  bson.ObjectId `bson:"_id" json:"-"`
	Id                        string        `bson:"id" json:"id"`
	JobId                     string        `bson:"jobid" json:"jobid"`
	ProjectId                 string        `bson:"projectid" json:"projectid"`
	EnvId                     string        `bson:"envid" json:"envid"`
	ServiceGroupInstanceId    string        `bson:"service_group_instance_id" json:"service_group_instance_id"`
	ServiceOrderId            string        `bson:"service_order_id" json:"service_order_id"`
	ServiceOfferingInstanceId string        `bson:"service_offering_instance_id" json:"service_offering_instance_id"`
	UserId                    string        `bson:"user_id" json:"user_id"`
	TenantId                  string        `bson:"tenant_id" json:"tenant_id"`
}
