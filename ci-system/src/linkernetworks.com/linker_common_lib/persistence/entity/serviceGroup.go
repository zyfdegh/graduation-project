package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type Parameter struct {
	Key         string `bson:"key" json:"key"`
	Value       string `bson:"value" json:"value"`
	Description string `bson:"description" json:"description"`
	Editable    bool   `bson:"editable" json:"editable"`
}

type PortMapping struct {
	ContainerPort int    `bson:"containerPort" json:"containerPort"`
	HostPort      int    `bson:"hostPort" json:"hostPort"`
	ServicePort   int    `bson:"servicePort" json:"servicePort"`
	Protocol      string `bson:"protocol" json:"protocol"`
}

type Docker struct {
	Network        string        `bson:"network" json:"network"`
	Image          string        `bson:"image" json:"image"`
	Privileged     bool          `bson:"privileged" json:"privileged"`
	ForcePullImage bool          `bson:"forcePullImage" json:"forcePullImage"`
	PortMappings   []PortMapping `bson:"portMappings" json:"portMappings"`
	Parameters     []Parameter   `bson:"parameters" json:"parameters"`
}

type Volume struct {
	ContainerPath string `bson:"containerPath" json:"containerPath"`
	HostPath      string `bson:"hostPath" json:"hostPath"`
	Mode          string `bson:"mode" json:"mode"`
}

type Container struct {
	Type    string   `bson:"type" json:"type"`
	Docker  Docker   `bson:"docker" json:"docker"`
	Volumes []Volume `bson:"volumes" json:"volumes"`
}

// type Env struct {
// 	// SGO_ID             string `bson:"LINKER_SERVICE_GROUP_ORDER_ID" json:"LINKER_SERVICE_GROUP_ORDER_ID,omitempty"`
// 	SGI_ID             string `bson:"LINKER_SERVICE_GROUP_INSTANCE_ID" json:"LINKER_SERVICE_GROUP_INSTANCE_ID,omitempty"`
// 	SO_ID              string `bson:"LINKER_SERVICE_OFFERING_ID" json:"LINKER_SERVICE_OFFERING_ID,omitempty"`
// 	SERVER_ID          string `bson:"SERVER_ID" json:"SERVER_ID,omitempty"`
// 	LINKER_ADDR        string `bson:"LINKER_ADDR" json:"LINKER_ADDR,omitempty"`
// 	LINKER_BR          string `bson:"LINKER_BR" json:"LINKER_BR,omitempty"`
// 	LINKER_CONF_SCRIPT string `bson:"LINKER_CONF_SCRIPT" json:"LINKER_CONF_SCRIPT,omitempty"`
// 	LINKER_CONF_PARAMS string `bson:"LINKER_CONF_PARAMS" json:"LINKER_CONF_PARAMS,omitempty"`
// 	SG_ID              string `bson:"LINKER_SERVICE_GROUP_ID" json:"LINKER_SERVICE_GROUP_ID,omitempty"`
// }

type App struct {
	ObjectId    bson.ObjectId     `bson:"_id,omitempty" json:"_id,omitempty"`
	Id          string            `bson:"id" json:"id"`
	Cpus        float32           `bson:"cpus" json:"cpus"`
	Mem         int16             `bson:"mem" json:"mem"`
	Instances   int               `bson:"instances" json:"instances"`
	Cmd         string            `bson:"cmd,omitempty" json:"cmd,omitempty"`
	Container   Container         `bson:"container,omitempty" json:"container,omitempty"`
	Env         map[string]string `bson:"env" json:"env"`
	Constraints [][]string        `json:"constraints"`
	Executor    string            `bson:"executor,omitempty" json:"executor,omitempty"`
	Scale       ScaleConfig       `bson:"scale" json:"scale"`
	Openstack   Openstack         `bson:"openstack,omitempty" json:"openstack,omitempty"`
	User_id     string            `bson:"user_id" json:"user_id"`
	Tenant_id   string            `bson:"tenant_id" json:"tenant_id"`
	TimeCreate  string            `bson:"time_create" json:"time_create"`
	TimeUpdate  string            `bson:"time_update" json:"time_update"`
}

type ScaleConfig struct {
	Enabled   bool `bson:"enabled" json:"enabled"`
	MinNum    int  `bson:"min_num" json:"min_num"`
	MaxNum    int  `bson:"max_num" json:"max_num"`
	ScaleStep int  `bson:"scale_step" json:"scale_step"`
}

type Group struct {
	Id           string   `bson:"id" json:"id"`
	Dependencies []string `bson:"dependencies" json:"dependencies"`
	Apps         []App    `bson:"apps,omitempty" json:"apps,omitempty"`
	Groups       []Group  `bson:"groups,omitempty" json:"groups,omitempty"`
	Billing      bool     `bson:"billing" json:"billing"`
}

type ServiceGroup struct {
	ObjectId      bson.ObjectId `bson:"_id" json:"_id"`
	Id            string        `bson:"id" json:"id"`
	LinkerVersion float32       `bson:"linker_version" json:"linker_version"`
	Groups        []Group       `bson:"groups" json:"groups"`
	State         string        `bson:"state" json:"state"`
	User_id       string        `bson:"user_id" json:"user_id"`
	Tenant_id     string        `bson:"tenant_id" json:"tenant_id"`
	TimeCreate    string        `bson:"time_create" json:"time_create"`
	TimeUpdate    string        `bson:"time_update" json:"time_update"`
}

type ServiceGroupNoContainer struct {
	Id            string             `bson:"id" json:"id"`
	LinkerVersion float32            `bson:"linker_version" json:"linker_version"`
	Groups        []GroupNoContainer `bson:"groups" json:"groups"`
}

type GroupNoContainer struct {
	Id           string             `bson:"id" json:"id"`
	Dependencies []string           `bson:"dependencies" json:"dependencies"`
	Apps         []AppNoContainer   `bson:"apps,omitempty" json:"apps,omitempty"`
	Groups       []GroupNoContainer `bson:"groups,omitempty" json:"groups,omitempty"`
}

type AppNoContainer struct {
	Id          string            `bson:"id" json:"id"`
	Cpus        float32           `bson:"cpus" json:"cpus"`
	Mem         int16             `bson:"mem" json:"mem"`
	Instances   int               `bson:"instances" json:"instances"`
	Cmd         string            `bson:"cmd,omitempty" json:"cmd,omitempty"`
	Env         map[string]string `bson:"env" json:"env"`
	Constraints [][]string        `json:"constraints"`
	Executor    string            `bson:"executor,omitempty" json:"executor,omitempty"`
}

type Openstack struct {
	Image  string `bson:"image" json "image"`
	Flavor string `bson:"flavor" json "flavor"`
}
