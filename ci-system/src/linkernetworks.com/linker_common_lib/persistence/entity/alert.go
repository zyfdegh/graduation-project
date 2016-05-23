package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type AlertMessage struct {
	ObjectId   bson.ObjectId `bson:"_id,omitempty" json:"_id,omitempty"`
	Version    string        `bson:"version" json:"version"`
	Status     string        `bson:"status" json:"status"`
	Alert      []Alert      `bson:"alert" json:"alert"`
	TimeCreate string        `bson:"time_create" json:"time_create"`
	TimeUpdate string        `bson:"time_update" json:"time_update"`
}

type Alert struct {
	Summary     string  `bson:"summary" json:"summary"`
	Description string  `bson:"description" json:"description"`
	Labels      Label   `bson:"labels" json:"labels"`
	Payload     PayLoad `bson:"payload" json:"payload"`
}

type Label struct {
	AlertName		  string `bson:"alert_name" json:"alert_name"`
	Image 			  string `bson:"image" json:"image"`
	Name  			  string `bson:"name" json:"name"`
	Id    			  string `bson:"id"  json:"id"`
	ServiceGroupId    string `bson:"service_group_id"  json:"service_group_id"`
	ServiceGroupInstanceId   string `bson:"service_group_instance_id"  json:"service_group_instance_id"`
	ServiceOrderId   string `bson:"service_order_id"  json:"service_order_id"`
	AppContainerId   string `bson:"app_container_id"  json:"app_container_id"`
	MesosTaskId		 string `bson:"mesos_task_id"  json:"mesos_task_id"`
	CpuUsage    string `bson:"cpu_usage"  json:"cpu_usage"`
	CpuUsageLowResult    string `bson:"cpu_usage_low_result"  json:"cpu_usage_low_result"`
	CpuUsageHighResult    string `bson:"cpu_usage_high_result"  json:"cpu_usage_high_result"`
	MemoryUsage     string `bson:"memory_usage"  json:"memory_usage"`
	MemoryUsageLowResult    string `bson:"memory_usage_low_result"  json:"memory_usage_low_result"`
	MemoryUsageHighResult    string `bson:"memory_usage_high_result"  json:"memory_usage_high_result"`
}

type PayLoad struct {
	ActiveSince  string `bson:"activeSince" json:"activeSince"`
	AlertingRule string `bson:"alertingRule" json:"alertingRule"`
	GeneratorURL string `bson:"generatorURL" json:"generatorURL"`
	Value        string `bson:"value" json:"value"`
}
