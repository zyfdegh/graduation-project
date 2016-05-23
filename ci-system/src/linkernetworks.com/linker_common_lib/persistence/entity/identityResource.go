package entity

import (
	"gopkg.in/mgo.v2/bson"
)

type LoginResponse struct {
	Id       string `json:"id"`
	Alias    string `json:"alias"`
	Rolename string `json:"rolename"`
	UserId   string `json:"userid"`
}

type Token struct {
	ObjectId   bson.ObjectId `bson:"_id" json:"_id"`
	Expire     float64       `bson:"expiretime" json:"expiretime"`
	User       UserPart      `bson:"user" json:"user"`
	Tenant     TenantPart    `bson:"tenant" json:"tenant"`
	Role       RolePart      `bson:"role" json:"role"`
	TimeCreate string        `bson:"time_create" json:"time_create"`
	TimeUpdate string        `bson:"time_update" json:"time_update"`
}

type UserPart struct {
	Id       string `bson:"id" json:"id"`
	Username string `bson:"username" json:"username"`
}

type TenantPart struct {
	Id         string `bson:"id" json:"id"`
	Tenantname string `bson:"tenantname" json:"tenantname"`
}

type RolePart struct {
	Id       string `bson:"id" json:"id"`
	Rolename string `bson:"rolename" json:"rolename"`
}

type User struct {
	ObjectId   bson.ObjectId `bson:"_id" json:"_id"`
	Username   string        `bson:"username" json:"username"`
	Alias      string        `bson:"alias" json:"alias"`
	Password   string        `bson:"password" json:"password"`
	Tenantname string        `bson:"tenantname" json:"tenantname"`
	Rolename   string        `bson:"rolename" json:"rolename"`
	Email      string        `bson:"email" json:"email"`
	State      int           `bson:"state" json:"state"`
	ActiveCode string        `bson:"activecode" json:"activecode"`
	Address    string        `bson:"address" json:"address"`
	Company    string        `bson:"company" json:"company"`
	PhoneNum   string        `bson:"phonenum" json:"phonenum"`
	Source     string        `bson:"source" json:"source"`
	TimeCreate string        `bson:"time_create" json:"time_create"`
	TimeUpdate string        `bson:"time_update" json:"time_update"`
}

type Tenant struct {
	ObjectId    bson.ObjectId `bson:"_id" json:"_id"`
	Tenantname  string        `bson:"tenantname" json:"tenantname"`
	Description string        `bson:"desc" json:"desc"`
	TimeCreate  string        `bson:"time_create" json:"time_create"`
	TimeUpdate  string        `bson:"time_update" json:"time_update"`
}

type Role struct {
	ObjectId    bson.ObjectId `bson:"_id" json:"_id"`
	Rolename    string        `bson:"rolename" json:"rolename"`
	Description string        `bson:"desc" json:"desc"`
	TimeCreate  string        `bson:"time_create" json:"time_create"`
	TimeUpdate  string        `bson:"time_update" json:"time_update"`
}

type OrCheck struct {
	Basechecks []BaseCheck `json:"basechecks"`
	Andchecks  []AndCheck  `json:"andchecks"`
	Orchecks   []OrCheck   `json:"orchecks"`
}

type AndCheck struct {
	Basechecks []BaseCheck `json:"basechecks"`
	Andchecks  []AndCheck  `json:"andchecks"`
	Orchecks   []OrCheck   `json:"orchecks"`
}

type BaseCheck struct {
	Checktype string `json:"checktype"`
	Value     string `json:"value"`
}
