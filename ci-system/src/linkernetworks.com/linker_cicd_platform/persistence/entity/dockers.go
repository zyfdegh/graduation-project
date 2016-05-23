package entity

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	DOCKER_FILE_STATUS_UPLOADED  = "uploaded"
	DOCKER_FILE_STATUS_PUBLISHED = "published"
	DOCKER_FILE_STATUS_REJECTED  = "rejected"
)

type DockerFile struct {
	ObjectId   		bson.ObjectId `bson:"_id" json:"-"`
	Id         		string        `bson:"id" json:"id"`
	DockerFile 		string        `bson:"dockerfile" json:"dockerfile"`
	ZipFile    		string        `bson:"zipfile" json:"zipfile"`
	ImageName  		string        `bson:"imagename" json:"imagename"`
	UserId     		string        `bson:"user_id" json:"user_id"`
	TenantId   		string        `bson:"tenant_id" json:"tenant_id"`
	Status     		string        `bson:"status" json:"status"`
	BuildStatus		string		  `bson:"build_status" json:"build_status"`
	Version    		string        `bson:"version" json:"version"`
}
