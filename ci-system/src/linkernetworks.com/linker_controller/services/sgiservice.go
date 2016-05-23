package services

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"sync"
)

var (
	sgiService *SgiService = nil
	onceSgi    sync.Once

	SGI_STATUS_DEPLOYING   = "DEPLOYING"
	SGI_STATUS_DEPLOYED    = "DEPLOYED"
	SGI_STATUS_TERMINATING = "TERMINATING"
	SGI_STATUS_TERMINATED  = "TERMINATED"
	SGI_STATUS_MODIFYING   = "MODIFYING"
	SGI_STATUS_FAILED      = "FAILED"
	SGI_STATUS_REPARING    = "REPAIRING"
	SGI_STATUS             = []string{SGI_STATUS_REPARING, SGI_STATUS_FAILED, SGI_STATUS_TERMINATED,
		SGI_STATUS_DEPLOYING, SGI_STATUS_DEPLOYED, SGI_STATUS_TERMINATING, SGI_STATUS_MODIFYING}

	SGI_ERROR_CREATE string = "E11051"
	SGI_ERROR_UPDATE string = "E11053"
	SGI_ERROR_DELETE string = "E11052"
	SGI_ERROR_QUERY  string = "E11054"
)

type SgiService struct {
	collectionName string
}

func GetSgiService() *SgiService {
	onceSgi.Do(func() {
		logrus.Debugf("Once called from sgiService ......................................")
		sgiService = &SgiService{"service_group_instance"}
	})
	return sgiService
}

func (p *SgiService) QueryAll(skip int, limit int, x_auth_token string) (total int, sgis []entity.ServiceGroupInstance,
	errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_sgis", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	logrus.Debugf("auth query is %v", authQuery)

	selector := generateQueryWithAuth(bson.M{}, authQuery)
	logrus.Debugf("selector is %v", selector)
	sort := ""
	sgis = []entity.ServiceGroupInstance{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, sort}
	total, err = dao.HandleQueryAll(&sgis, queryStruct)
	if err != nil {
		logrus.Errorf("list sgi [token=%v] failed, error is %v", x_auth_token, err)
		errorCode = SGI_ERROR_QUERY
	}
	return
}

func (p *SgiService) QueryById(objectId string, x_auth_token string) (sgi entity.ServiceGroupInstance,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// do authorize first
	if authorized := GetAuthService().Authorize("get_sgi", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get sgi with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	sgi = entity.ServiceGroupInstance{}
	err = dao.HandleQueryOne(&sgi, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query sgi [objectId=%v] error is %v", objectId, err)
		errorCode = SGI_ERROR_QUERY
	}
	return
}

func (p *SgiService) Create(sgi entity.ServiceGroupInstance, x_auth_token string) (newSgi entity.ServiceGroupInstance,
	errorCode string, err error) {
	logrus.Infof("start to create sgi [%v]", sgi)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_sgi", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create sgi [%v] error is %v", sgi, err)
		return
	}
	// generate ObjectId
	if !bson.IsObjectIdHex(sgi.ObjectId.Hex()) {
		sgi.ObjectId = bson.NewObjectId()
	}

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = SGI_ERROR_CREATE
		logrus.Errorf("get token failed when create sgi [%v], error is %v", sgi, err)
		return
	}

	// set token_id and user_id from token
	sgi.Tenant_id = token.Tenant.Id
	sgi.User_id = token.User.Id

	// set created_time and updated_time
	sgi.TimeCreate = dao.GetCurrentTime()
	sgi.TimeUpdate = sgi.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, sgi)
	if err != nil {
		errorCode = SGI_ERROR_CREATE
		logrus.Errorf("insert sgi [%v] to db error is %v", sgi, err)
		return
	}

	newSgi = sgi
	return
}

func (p *SgiService) UpdateById(objectId string, sgi entity.ServiceGroupInstance, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update sgi [%v]", sgi)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_sgi", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update sgi with objectId [%v] error is %v", objectId, err)
		return
	}
	// validate sgi
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	sgi.ObjectId = bson.ObjectIdHex(objectId)
	sgi.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&sgi, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update sgi [%v] error is %v", sgi, err)
		errorCode = SGI_ERROR_UPDATE
	}
	return
}

func (p *SgiService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete sgi with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_sgi", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete sgi with objectId [%v] error is %v", objectId, err)
		return
	}

	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	err = dao.HandleDelete(p.collectionName, true, selector)
	if err != nil {
		logrus.Errorf("delete sgi [objectId=%v] error is %v", objectId, err)
		errorCode = SGI_ERROR_DELETE
	}
	return
}

func (p *SgiService) UpdateRepairIdAndStatusById(objectId string, repairId string, status string, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update sgi by objectId [%v] repairId to %v and status to %v", objectId, repairId, status)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_sgi", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update sgi with objectId [%v] repairId to [%v] and status to [%v] failed, error is %v", objectId, repairId, status, err)
		return
	}
	// validate objectId
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// get sgi by objectId
	sgi, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("get sgi by objeceId [%v] failed, error is %v", objectId, err)
		return
	}

	if sgi.LifeCycleStatus == status {
		logrus.Infof("this sgi [%v] is already in state [%v]", sgi, status)
		return false, "", nil
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	sgi.ObjectId = bson.ObjectIdHex(objectId)
	sgi.LifeCycleStatus = status
	sgi.RepairId = repairId
	sgi.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&sgi, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update sgi with objectId [%v] repairId to [%v] and status to [%v] failed, error is %v", objectId, repairId, status, err)
	}
	return
}

func (p *SgiService) UpdateStateById(objectId string, newState string, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update sgi by objectId [%v] status to %v", objectId, newState)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_sgi", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update sg with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
		return
	}
	// validate objectId
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// get sgi by objectId
	sgi, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("get sgi by objeceId [%v] failed, error is %v", objectId, err)
		return
	}

	if sgi.LifeCycleStatus == newState {
		logrus.Infof("this sgi [%v] is already in state [%v]", sgi, newState)
		return false, "", nil
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	sgi.ObjectId = bson.ObjectIdHex(objectId)
	sgi.LifeCycleStatus = newState
	sgi.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&sgi, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update sgi with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
	}
	return
}

func (p *SgiService) queryAllByquery(skip int, limit int, query bson.M) (total int, sgis []entity.ServiceGroupInstance,
	errorCode string, err error) {
	sort := ""
	sgis = []entity.ServiceGroupInstance{}
	queryStruct := dao.QueryStruct{p.collectionName, query, skip, limit, sort}
	total, err = dao.HandleQueryAll(&sgis, queryStruct)
	if err != nil {
		logrus.Errorf("list sgi by query [%v] failed, error is %v", query, err)
		errorCode = SGI_ERROR_QUERY
	}
	return
}

/**
* Method to change the instance number refine
* for the app in a certain service group instance
 */
func (p *SgiService) ChangeInstanceNumber(sgi *entity.ServiceGroupInstance,
	appId string, number int) (flag bool, err error) {

	app, err := entity.FindAppInSgi(sgi, appId)
	if err != nil {
		logrus.Errorf("Can not find the App in the service group instance, err is %v", err)
		flag = false
		return
	} else {
		app.Instances = number
		flag = true
	}
	return
}
