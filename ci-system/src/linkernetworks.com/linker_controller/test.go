package main

import (
	// "encoding/json"
	"fmt"
	// "github.com/compose/mejson"
	"encoding/json"
	// "flag"
	// "github.com/magiconair/properties"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	// "reflect"
	// "linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/persistence/session"
	"strings"
	"time"
)

// var (
// 	Props          *properties.Properties
// 	PropertiesFile = flag.String("config", "/home/centos/controller.properties", "the configuration file")
// 	MongoAlias     string
// )

// func init() {
// 	// get configuration
// 	flag.Parse()
// 	// fmt.Printf("PropertiesFile is %s\n", *PropertiesFile)
// 	var err error
// 	if Props, err = properties.LoadFile(*PropertiesFile, properties.UTF8); err != nil {
// 		fmt.Printf("[error] Unable to read properties:%v\n", err)
// 	}
// }

type Todo struct {
	ObjectId    bson.ObjectId     `bson:"_id" json:"_id"`
	Env         map[string]string `bson:"env" json:"env"`
	CreatedTime time.Time         `bson:"created_time" json:"created_time"`
	UpdatedTime time.Time         `bson:"updated_time" json:"updated_time"`
}

type Scale struct {
	MinNum    int
	ScaleStep int
}

func mainmath() {
	num := 5
	scale := Scale{3, 3}

	fmt.Print((num - scale.MinNum) % scale.ScaleStep)
}

func maindot() {
	env := make(map[string]string)
	env["a.b"] = "c.d"
	todo := Todo{
		ObjectId:    bson.NewObjectId(),
		Env:         env,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	fmt.Printf("%v\n", todo)
	newDoc := remove_dots(todo)

	newTodo := Todo{}
	in, _ := bson.Marshal(newDoc)
	bson.Unmarshal(in, &newTodo)
	fmt.Printf("%v\n", newTodo)
}

func remove_dots(doc interface{}) (newDoc bson.M) {
	// get document type
	// st := reflect.TypeOf(doc)
	// fmt.Println(st)

	document := bson.M{}
	in, _ := bson.Marshal(doc)
	bson.Unmarshal(in, &document)

	newDoc = bson.M{}
	// logrus.Debugf("remove dot from document %v", document)
	for key, subdoc := range document {
		if subdoc, ok := subdoc.(bson.M); ok {
			newDoc[key] = remove_dots(subdoc)
		} else {
			newDoc[key] = document[key]
		}

		if strings.Contains(key, ".") {
			// logrus.Infof("replace dot for key %v", key)
			fmt.Println("replace dot for key ", key)
			newDoc[strings.Replace(key, ".", "\u2024", -1)] = newDoc[key]
			delete(newDoc, key)
			// } else {
			// logrus.Debugf("key %v don't have dot.", key)
		}
	}
	return
}

// func maindb() {
// 	MongoAlias = Props.GetString("db.alias", "dev")
// 	sessMng := session.NewSessionManager(Props.FilterPrefix("mongod."), MongoAlias)
// 	defer sessMng.CloseAll()
// 	// handleInsert(sessMng)
// 	handleQuery(sessMng)
// }

func handleInsert(sessMng *session.SessionManager) {
	todo := Todo{
		ObjectId:    bson.NewObjectId(),
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}

	session, _, err := sessMng.GetDefault()
	if err != nil {
		return
	}
	session.DB("test").C("todo").Insert(&todo)
	fmt.Printf("todo is %v\n", todo)

	appjson, _ := json.Marshal(todo)
	fmt.Printf("json is %v\n", string(appjson))
}

func handleQuery(sessMng *session.SessionManager) {
	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex("565d38faa9a58d806f01cd41")
	todo := Todo{}
	session, _, err := sessMng.GetDefault()
	if err != nil {
		return
	}
	session.DB("test").C("todo").Find(selector).One(&todo)

	fmt.Printf("todo is %v\n", todo)

	appjson, _ := json.Marshal(todo)
	fmt.Printf("json is %v\n", string(appjson))
}

// func main() {
// 	str := "/etc/ldap/slapd.d abc "
// 	fmt.Println("start:" + strings.TrimSpace(str) + "end.")
// }

func main1() {
	jsonstr := `{
		"_id":                      "5621b1f695f80f2d27f61b25",
		"app_container_id":         "/cliu-openldap/server/openldap",
		"cpus":                     0.5,
		"docker_container_ip":      "172.17.31.12",
		"docker_container_long_id": "aabd3eafa433d686819ab967059dc91d3abd27af134f7bd195294c98f5e3d313",
		"docker_container_name":    "mesos_cliu-openldap-8b61445b-7476-11e5-8357-525400be0aaf_server_openldap-5621b1f495f80f36ab7824ed",
		"docker_container_port":    "389",
		"id":                   "5621b1f695f80f2d27f61b25",
		"lifecycle_status":     "CONFIGED",
		"marathon_app_path":    "/cliu-openldap-8b61445b-7476-11e5-8357-525400be0aaf/server/openldap",
		"marathon_app_version": "2015-10-17T02:27:08.885Z",
		"mem.mem":                       512,
		"mesos_sand_box":            "/mnt/ceph/client",
		"mesos_slave":               "172.17.2.44",
		"mesos_slave_host_port":     "32769",
		"mesos_slave_ip":            "",
		"mesos_task_id":             "cliu-openldap-8b61445b-7476-11e5-8357-525400be0aaf_server_openldap.912e4634-7476-11e5-af37-b61761037964",
		"service_group_id":          "/cliu-openldap",
		"service_group_instance_id": "5621b1f395f80f2d27f61b22",
		"service_order_id":          "5621b1f395f80f2d27f61b21",
		"time_create":               "2015-10-17T10:27:02+08:00",
		"time_update":               "2015-10-17T11:39:47+08:00",
		"version":                   "",
		"volumes": {
			"/etc/ldap/slapd.d": "/var/lib/docker/volumes/ff2af682d730b8591d1a27ffa3d59fe9988958b025ff37849a0d228e06bca5ce/_data",
			"/var/lib/ldap":     "/var/lib/docker/volumes/0a7cd2eacf3789f740b084833dfae693c51d264c2739e93421ee1f7e1cbcb843/_data"
		}
	}`

	document := bson.M{}
	// Handle JSON parsing manually here, instead of relying on go-restful's
	// req.ReadEntity. This is because ReadEntity currently parses JSON with
	// UseNumber() which turns all numbers into strings. See:
	// https://github.com/emicklei/mora/pull/31
	decoder := json.NewDecoder(ioutil.NopCloser(strings.NewReader(jsonstr)))
	err := decoder.Decode(&document)
	if err != nil {
		fmt.Printf("Error is %v\n", err)
	}

	// fmt.Printf("document=%v\n", document)

	newDoc := remove_dots(document)
	// remove_dots(document)

	fmt.Printf("newDoc is %v", newDoc)
}

// func remove_dots(document *bson.M) {
// 	for key, subdoc := range *document {
// 		fmt.Printf("%v\n", key)
// 		if subdoc, ok := subdoc.(bson.M); ok {
// 			remove_dots(&subdoc)
// 		}

// 		if strings.Contains(key, ".") {
// 			*document[strings.Replace(key, ".", "\u2024", -1)] = subdoc
// 			delete(*document, key)
// 		}
// 	}
// }

// func _remove_dots(document *bson.M) (newDoc *bson.M) {
// 	for key := range document {
// 		fmt.Printf("%v\n", key)
// 		if subdoc, ok := document[key].(map[string]interface{}); ok {
// 			if strings.Contains(key, ".") {
// 				document[strings.Replace(key, ".", "\u2024", -1)] = document[key]
// 				delete(document, key)
// 			} else {

// 			}
// 			document[key] = remove_dots(subdoc)
// 			delete(document, key)
// 		} else {

// 		}

// 		if strings.Contains(key, ".") {
// 			document[strings.Replace(key, ".", "\u2024", -1)] = document[key]
// 			delete(document, key)
// 		}
// 	}
// }

// func ConvertToBson(obj interface{}) (document bson.M, err error) {
// 	b, err := json.Marshal(obj)
// 	if err != nil {
// 		return
// 	}
// 	reader := strings.NewReader(string(b))
// 	decoder := json.NewDecoder(reader)
// 	err = decoder.Decode(&document)
// 	if err != nil {
// 		return
// 	}

// 	document, err = mejson.Unmarshal(document)
// 	return
// }
