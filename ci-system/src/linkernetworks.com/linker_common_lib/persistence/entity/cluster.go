package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type Cluster struct {
	ObjectId    bson.ObjectId `bson:"_id" json:"_id"`
	ClusterName string        `bson:"cluster_name" json:"cluster_name"`
	Supplier    string        `bson:"supplier" json:"supplier"`
	Instances   int           `bson:"instances" json:"instances"`
	User_id     string        `bson:"user_id" json:"user_id"`
	Tenant_id   string        `bson:"tenant_id" json:"tenant_id"`
	Status      string        `bson:"status" json:"status"`
	ClusterType string        `bson:"cluster_type" json:"cluster_type"`
	Disk        int           `bson:"disk" json:"disk"`
	Flavor      Flavor        `bson:"flavor" json:"flavor"`

	TimeCreate string `bson:"time_create" json:"time_create"`
	TimeUpdate string `bson:"time_update" json:"time_update"`
}

type Flavor struct {
	ObjectId     bson.ObjectId `bson:"_id" json:"_id"`
	Cpu          int           `bson:"cpu" json:"cpu"`
	Mem          int           `bson:"mem" json:"mem"`
	ProviderType string        `bson:"provider_type" json:"provider_type"`
	Price        float64       `bson:"price" json:"price"`
	Type        string          `bson:"type" json:"type"`
	FlavorName   string        `bson:"flavor_name" json:"flavor_name"`

	TimeCreate string `bson:"time_create" json:"time_create"`
	TimeUpdate string `bson:"time_update" json:"time_update"`
}

type Host struct {
	ObjectId      bson.ObjectId `bson:"_id" json:"_id"`
	HostName      string        `bson:"host_name" json:"host_name"`
	ClusterId     string        `bson:"cluster_id" json:"cluster_id"`
	ClusterName   string        `bson:"cluster_name" json:"cluster_name"`
	Status        string        `bson:"status" json:"status"`
	IP            string        `bson:"ip" json:"ip"`
	IsMasterNode  bool          `bson:"ismasternode" json:"ismasternode"`
	CloudProxyId   string        `bson:"cloudproxy_id" json:"cloudproxy_id"`
	DockerVersion string        `bson:"docker_version" json:"docker_version"`
	User_id       string        `bson:"user_id" json:"user_id"`
	Tenant_id     string        `bson:"tenant_id" json:"tenant_id"`
	Flavor        Flavor        `bson:"flavor" json:"flavor"`

	Lable      int       `bson:"lable" json:"lable"`	
	
	Date       int    `bson:"date" json:"date"`
	TimeCreate string `bson:"time_create" json:"time_create"`
	TimeUpdate string `bson:"time_update" json:"time_update"`
}

type ClusterOrder struct {
	ObjectId    bson.ObjectId `bson:"_id" json:"_id"`
	ClusterId   string        `bson:"cluster_id" json:"cluster_id"`
	ClusterName string        `bson:"cluster_name" json:"cluster_name"`
	SgoId       string        `bson:"sgo_id" json:"sgo_id"`
	SgId        string        `bson:"sgid" json:"sgid"`
	SgiId       string        `bson:"sgi_id" json:"sgi_id"`
	User_id     string        `bson:"user_id" json:"user_id"`
	Tenant_id   string        `bson:"tenant_id" json:"tenant_id"`

	TimeCreate string `bson:"time_create" json:"time_create"`
	TimeUpdate string `bson:"time_update" json:"time_update"`
}


type HostLog struct {
	ObjectId    bson.ObjectId `bson:"_id" json:"_id"`
	ClusterId   string        `bson:"cluster_id" json:"cluster_id"`
	ClusterName string        `bson:"cluster_name" json:"cluster_name"`
	HostId      string         `bson:"host_id" json:"host_id"`
	HostName      string        `bson:"host_name" json:"host_name"`
	Operation  string        `bson:"operation" json:"operation"`
	Content    string         `bson:"content" json:"content"`
	HostStatus   string    `bson:"host_status" josn:"host_status"`
	User_id     string        `bson:"user_id" json:"user_id"`
	Tenant_id   string        `bson:"tenant_id" json:"tenant_id"`
	
	Date               int           `bson:"date" json:"date"`
	TimeCreate string `bson:"time_create" json:"time_create"`
	TimeUpdate string `bson:"time_update" json:"time_update"`
}

type CloudProxy struct {
	CloudProvider      string    `bson:"cloudProvider" json:"cloudProvider"`
	SecurityGroupId    string    `bson:"securityGroupId" json:"securityGroupId"`
	SecuriyKey          string     `bson:"securityKey" json:"securityKey"`
	InstanceType       string	 `bson:"instanceType" json:"instanceType"`
	HostName        string        `bson:"hostName" json:"hostName"`
	RequestId       string          `bson:"requestId" json:"requestId"`
	InstanceId        string         `bson:"instanceId" json:"instanceId"`
}

type HostDeacribe struct {
	CloudProvider      string    `bson:"cloudProvider" json:"cloudProvider"`
	CreationTime  string      `bson:"creationTime" json:"creationTime"`
	HostName      string        `bson:"hostName" json:"hostName"`
	InstanceId        string         `bson:"instanceId" json:"instanceId"`
	InstanceName  string        `bson:"instanceName" json:"instanceName"`
	KeyName      string          `bson:"keyName" json:"keyName"`
	InstanceType   string        `bson:"instanceType" json:"instanceType"`
	RegionId         string        `bson:"regionId" json:"regionId"`
	ZoneId        string           `bson:"zoneId" json:"zoneId"`
	IpAddress     string       `bson:"ipAddress" json:"ipAddress"`
	PublicIpAddressList    []string   `bson:"publicIdAddressList" json:"publicIdAddressList"`
	PrivateIpAddressList  []string    `bson:"privateIpAddressList" json:"privateIpAddressList"`
	SecurityGroupIdList   []string     `bson:"securityGroupIdList" json:"securityGroupIdList"`
	status    string    `bson:"status" json:"status"`
}

