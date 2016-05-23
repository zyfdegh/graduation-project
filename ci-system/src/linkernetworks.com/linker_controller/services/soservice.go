package services

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"strings"
	"sync"
)

var (
	soService *SoService = nil
	onceSo    sync.Once
)

type SoService struct {
	collectionName string
}

func GetSoService() *SoService {
	onceSo.Do(func() {
		logrus.Debugf("Once called from soService ......................................")
		soService = &SoService{"service_group_offering"}
	})
	return soService
}

func (p *SoService) Create(so entity.ServiceGroup, x_auth_token string) (newSo entity.ServiceGroup,
	errorCode string, err error) {
	logrus.Infof("start to create so [%v]", so)

	// generate ObjectId
	if !bson.IsObjectIdHex(so.ObjectId.Hex()) {
		so.ObjectId = bson.NewObjectId()
	}

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		logrus.Errorf("get token failed when create so [%v], error is %v", so, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	// set token_id and user_id from token
	so.Tenant_id = token.Tenant.Id
	so.User_id = token.User.Id

	// set created_time and updated_time
	so.TimeCreate = dao.GetCurrentTime()
	so.TimeUpdate = so.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, so)
	if err != nil {
		logrus.Errorf("insert so [%v] to db error is %v", so, err)
		errorCode = SG_ERROR_CREATE
		return
	}

	newSo = so
	return
}

func (p *SoService) QueryById(objectId string, x_auth_token string) (so entity.ServiceGroup,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	so = entity.ServiceGroup{}
	err = dao.HandleQueryOne(&so, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query so [objectId=%v] error is %v", objectId, err)
		errorCode = SG_ERROR_QUERY
	}
	return
}

func (p *SoService) GetSOContainerInfo(soId, sgiId, appId,
	x_auth_token string) (app *entity.App, errorCode string, err error) {
	// validate
	if !bson.IsObjectIdHex(soId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	if !bson.IsObjectIdHex(sgiId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// get so and sgi object from db
	so, _, err := p.QueryById(soId, x_auth_token)
	if err != nil {
		logrus.Errorf("find service grou offering [%v] failed, error is %v",
			soId, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}
	sgi, _, err := GetSgiService().QueryById(sgiId, x_auth_token)
	if err != nil {
		logrus.Errorf("find service grou instance [%v] failed, error is %v",
			sgiId, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	// get app from service group offering
	app, err = entity.GetAppFromServiceGroup(&so, appId)
	if err != nil {
		logrus.Errorf("can not find app [%v] from so %v",
			appId, so)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	logrus.Debugf("found app [%v] in so [%v]", app, so)

	// refine docker parameters
	if app.Container.Docker.Parameters != nil {
		for numS, element := range app.Container.Docker.Parameters {
			logrus.Debugf("key: " + element.Key)
			logrus.Debugf("value: " + element.Value)
			neededParams := strings.Split(element.Value, "=")
			envname := neededParams[0]
			envvalue := neededParams[1]
			logrus.Debugf("envname: " + envname)
			logrus.Debugf("envvalue: " + envvalue)
			if strings.HasPrefix(envvalue, "%") && strings.HasSuffix(envvalue, "%") {
				envvalue = strings.Trim(envvalue, "%")
				var tmpAppContainerInfo = entity.AppContainerInstance{
					// here, the ObjectId is just a placeholder, not used in logic
					ObjectId: bson.ObjectIdHex("56580c14a51f255a5fc1bed1"),
				}
				replacedenvvalue := getGroupInstanceInfo(&tmpAppContainerInfo, envvalue, &sgi, "ALL")
				logrus.Infoln("replacedenvvalue: " + replacedenvvalue)
				app.Container.Docker.Parameters[numS].Value = envname + "=" + replacedenvvalue
			}
		}
	}

	return
}
