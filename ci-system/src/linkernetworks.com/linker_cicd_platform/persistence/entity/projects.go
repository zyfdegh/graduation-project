package entity

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	PROJECT_STATUS_RUNNING    = "running"
	PROJECT_STATUS_TERMINATED = "terminated"
)

type Project struct {
	ObjectId       bson.ObjectId `bson:"_id" json:"-"`
	Id             string        `bson:"id" json:"id"`
	Name           string        `bson:"name" json:"name"`
	GitUrl         string        `bson:"git_url" json:"git_url"`
	OpsEnvId       string        `bson:"opsenv_id" json:"opsenv_id"`
	ServiceModelId string        `bson:"sm_id" json:"sm_id"` /*eg: com.linkernetworks*/
	Artifacts      []Artifact    `bson:"artifacts" json:"artifacts"`
	Status         string        `bson:"status" json:"status"`
}

type Artifact struct {
	ObjectId      bson.ObjectId `bson:"_id" json:"-"`
	Id            string        `bson:"id" json:"id"`
	GroupId       string        `bson:"group_id" json:"group_id"`
	Name          string        `bson:"name" json:"name"`
	Type          string        `bson:"type" json:"type"`
	DockerfileIds []string      `bson:"df_ids" json:"df_ids"`
}
