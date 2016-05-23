package entity

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	OPSENV_STATUS_TERMINATED = "terminated"
)

type OpsEnv struct {
	ObjectId                  bson.ObjectId `bson:"_id" json:"-"`
	Id                        string        `bson:"id" json:"id"`
	Name                      string        `bson:"name" json:"name"`
	ServiceGroupInstanceId    string        `bson:"service_group_instance_id" json:"service_group_instance_id"`
	ServiceOrderId            string        `bson:"service_order_id" json:"service_order_id"`
	ServiceOfferingInstanceId string        `bson:"service_offering_instance_id" json:"service_offering_instance_id"`
	UserId                    string        `bson:"user_id" json:"user_id"`
	TenantId                  string        `bson:"tenant_id" json:"tenant_id"`
	GerritInfo                string        `bson:"gerrit_info" json:"gerrit_info"`
	JenkinsInfo               string        `bson:"jenkins_info" json:"jenkins_info"`
	NexusInfo                 string        `bson:"nexus_info" json:"nexus_info"`
	MysqlInfo                 string        `bson:"mysql_info" json:"mysql_info"`
	LdapInfo                  string        `bson:"ldap_info" json:"ldap_info"`
	GerritDockerIP			  string 		`bson:"gerrit_docker_ip" json:"gerrit_docker_ip"`
	JenkinsDockerIP			  string 		`bson:"jenkins_docker_ip" json:"jenkins_docker_ip"`
	NexusDockerIP			  string 		`bson:"nexus_docker_ip" json:"nexus_docker_ip"`
	MysqlDockerIP			  string 		`bson:"mysql_docker_ip" json:"mysql_docker_ip"`
	LdapDockerIP			  string 		`bson:"ldap_docker_ip" json:"ldap_docker_ip"`
	GerritInternalInfo        string        `bson:"gerrit_internal_info" json:"gerrit_internal_info"`
	JenkinsInternalInfo       string        `bson:"jenkins_internal_info" json:"jenkins_internal_info"`
	NexusInternalInfo         string        `bson:"nexus_internal_info" json:"nexus_internal_info"`
	MysqlInternalInfo         string        `bson:"mysql_internal_info" json:"mysql_internal_info"`
	LdapInternalInfo          string        `bson:"ldap_internal_info" json:"ldap_internal_info"`
	GerritHttpPort			  string		`bson:"gerrit_http_port" json:"gerrit_http_port"`
	JenkinsHttpPort			  string		`bson:"jenkins_http_port" json:"jenkins_http_port"`
	LdapHttpPort			  string		`bson:"ldap_http_port" json:"ldap_http_port"`
	NexusHttpPort			  string		`bson:"nexus_http_port" json:"nexus_http_port"`
	Status                    string        `bson:"status" json:"status"`
}
