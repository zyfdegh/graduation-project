package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type AppContainerInstance struct {
	ObjectId               bson.ObjectId     `bson:"_id" json:"_id"`
	AppContainerId         string            `bson:"app_container_id" json:"app_container_id"`
	Version                string            `bson:"version" json:"version"`
	ServiceGroupId         string            `bson:"service_group_id" json:"service_group_id"`
	ServiceOrderId         string            `bson:"service_order_id" json:"service_order_id"`
	ServiceGroupInstanceId string            `bson:"service_group_instance_id" json:"service_group_instance_id"`
	Cpus                   float32           `bson:"cpus" json:"cpus"`
	Mem                    int16             `bson:"mem" json:"mem"`
	MarathonAppPath        string            `bson:"marathon_app_path" json:"marathon_app_path"`
	MarathonAppVersion     string            `bson:"marathon_app_version" json:"marathon_app_version"`
	MesosSlave             string            `bson:"mesos_slave" json:"mesos_slave"`
	MesosSlaveIp           string            `bson:"mesos_slave_ip" json:"mesos_slave_ip"`
	MesosTaskId            string            `bson:"mesos_task_id" json:"mesos_task_id"`
	MesosSandbox           string            `bson:"mesos_sand_box" json:"mesos_sand_box"`
	MesosSlaveHostPort     string            `bson:"mesos_slave_host_port" json:"mesos_slave_host_port"`
	DockerContainerName    string            `bson:"docker_container_name" json:"docker_container_name"`
	DockerContainerIp      string            `bson:"docker_container_ip" json:"docker_container_ip"`
	DockerContainerLongID  string            `bson:"docker_container_long_id" json:"docker_container_long_id"`
	DockerContainerPort    string            `bson:"docker_container_port" json:"docker_container_port"`
	Volumes                map[string]string `bson:"volumes" json:"volumes"`
	LifeCycleStatus        string            `bson:"lifecycle_status" json:"lifecycle_status"`
	TimeCreate             string            `bson:"time_create" json:"time_create"`
	TimeUpdate             string            `bson:"time_update" json:"time_update"`
}
