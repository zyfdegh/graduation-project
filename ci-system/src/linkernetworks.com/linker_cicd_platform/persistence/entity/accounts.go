package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type Accounts struct {
	ObjectId        bson.ObjectId   `bson:"_id" json:"-"`
	Id              string          `bson:"id" json:"id"`
	OpsEnvId        string          `bson:"opsenv_id" json:"opsenv_id"`
	GerritAccount   GerritAccount   `bson:"gerrit_account" json:"gerrit_account"`
	OpenldapAccount OpenldapAccount `bson:"openldap_account" json:"openldap_account"`
}

type GerritAccount struct {
	Username    string `bson:"username" json:"username"`
	Password    string `bson:"password" json:"password"`
	NewPassword string `bson:"new_password" json:"new_password"`
}

type OpenldapAccount struct {
	DN          string `bson:"dn" json:"dn"`
	Password    string `bson:"password" json:"password"`
	NewPassword string `bson:"new_password" json:"new_password"`
}
