package services

import (
	"errors"
	"strconv"
	"sync"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
)

var (
	sgoService *SgoService = nil
	onceSgo    sync.Once

	SGO_STATUS_DEPLOYING   = "DEPLOYING"
	SGO_STATUS_DEPLOYED    = "DEPLOYED"
	SGO_STATUS_TERMINATING = "TERMINATING"
	SGO_STATUS_TERMINATED  = "TERMINATED"
	SGO_STATUS_MODIFYING   = "MODIFYING"
	SGO_STATUS_FAILED      = "FAILED"
	SGO_STATUS             = []string{SGO_STATUS_DEPLOYING,
		SGO_STATUS_DEPLOYED, SGO_STATUS_TERMINATING,
		SGO_STATUS_TERMINATED, SGO_STATUS_MODIFYING,
		SGO_STATUS_FAILED}

	SGO_ERROR_CREATE string = "E11061"
	SGO_ERROR_SCALE  string = "E11063"
	SGO_ERROR_UPDATE string = "E11064"
	SGO_ERROR_DELETE string = "E11062"
	SGO_ERROR_QUERY  string = "E11065"
)

type SgoService struct {
	collectionName string
}

func GetSgoService() *SgoService {
	onceSgo.Do(func() {
		logrus.Debugf("Once called from sgoService ......................................")
		sgoService = &SgoService{"service_group_order"}
	})
	return sgoService
}

func (p *SgoService) Create(sgo entity.ServiceGroupOrder, x_auth_token string) (newSgo entity.ServiceGroupOrder,
	errorCode string, err error) {
	logrus.Infof("start to create sgo [%v]", sgo)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_sgo", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create sgo [%v] error is %v", sgo, err)
		return
	}

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		logrus.Errorf("get token failed when create sgo [%v], error is %v", sgo, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	// reset sg_id, sg_obj_id
	sgSelector := bson.M{}
	if sgo.ServiceGroupObjId == "" {
		sgSelector["id"] = sgo.ServiceGroupId

	} else {
		sgSelector["_id"] = bson.ObjectIdHex(sgo.ServiceGroupObjId)
	}

	_, sgs, _, err := GetSgService().queryByQuery(sgSelector, 0, 1, x_auth_token, true)
	if err != nil {
		logrus.Errorf("get service group failed, error is %v", err)
		errorCode = SGO_ERROR_CREATE
		return
	}

	if len(sgs) < 1 {
		logrus.Errorf("can not find service group by query [%v]", sgSelector)
		err = errors.New("invalidate service group")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	sg := sgs[0]
	sgo.ServiceGroupId = sg.Id
	sgo.ServiceGroupObjId = sg.ObjectId.Hex()

	// generate ObjectId
	sgo.ObjectId = bson.NewObjectId()

	// set marathon_group_id, for mulit instance order
	sgo.MarathonGroupId = makeMarathonGroupId(sg.Id)

	// set token_id and user_id from token
	sgo.Tenant_id = token.Tenant.Id
	sgo.User_id = token.User.Id

	// set created_time and updated_time
	sgo.TimeCreate = dao.GetCurrentTime()
	sgo.TimeUpdate = sgo.TimeCreate

	// quota check
	errorCode, err = GetQuotaService().CheckOrder(&sgo)
	if err != nil {
		return
	}

	// generate service group instance
	sgi := entity.ServiceGroupInstance{
		ObjectId:        bson.NewObjectId(),
		ServiceGroupId:  sg.Id,
		LifeCycleStatus: SGI_STATUS_DEPLOYING,
	}
	setServiceGroupIntanceGroups(&sgi, &sg)
	// save sgi
	sgi, _, err = GetSgiService().Create(sgi, x_auth_token)
	if err != nil {
		logrus.Errorf("save service group instance failed, err is %v", err)
		errorCode = SGO_ERROR_CREATE
		return
	}

	// make service offering, it's a copy of service group, only change the id with marathonId, and reset parameters with order's parameters.
	so := sg
	so.ObjectId = bson.NewObjectId()
	so.Id = sgo.MarathonGroupId
	//add parameter and instance info to service group and call marathon API
	sgo.ServiceGroupInstanceId = sgi.ObjectId.Hex()
	sgo.ServiceOfferingId = so.ObjectId.Hex()
	refineEnv(sgo, &so)
	refineParam(sgo, &so)

	// save so
	so, _, err = GetSoService().Create(so, x_auth_token)
	if err != nil {
		logrus.Errorf("save service offering failed, err is %v", err)
		errorCode = SGO_ERROR_CREATE
		// need to delete sgi
		_, _, delerr := GetSgiService().UpdateStateById(sgi.ObjectId.Hex(), SGI_STATUS_FAILED, x_auth_token)
		if delerr != nil {
			logrus.Errorf("delete sgi err is %v", delerr)
		}
		return
	}

	// insert sgo to mongodb
	sgo.LifeCycleStatus = SGO_STATUS_DEPLOYING
	err = dao.HandleInsert(p.collectionName, sgo)
	if err != nil {
		logrus.Errorf("insert sgo [%v] to db error is %v", sgo, err)
		errorCode = SGO_ERROR_CREATE
		// need to delete sgi
		_, _, delerr := GetSgiService().UpdateStateById(sgi.ObjectId.Hex(), SGI_STATUS_FAILED, x_auth_token)
		if delerr != nil {
			logrus.Errorf("delete sgi err is %v", delerr)
		}
		return
	}

	// remove container section in json which send to marathon
	marGroup, err := generateMarathonGroup(so, sgo)
	if err != nil {
		logrus.Errorf("convert so to marathon group failed, error is %v", err)
		errorCode = SGO_ERROR_CREATE
		// need to delete sgi and sgo
		_, _, delerr := GetSgiService().UpdateStateById(sgi.ObjectId.Hex(), SGI_STATUS_FAILED, x_auth_token)
		if delerr != nil {
			logrus.Errorf("set sgi to failed err is %v", delerr)
		}
		_, _, delerr = p.UpdateStateBySgiId(sgi.ObjectId.Hex(), SGO_STATUS_FAILED, x_auth_token)
		if delerr != nil {
			logrus.Errorf("set sgo to failed err is %v", delerr)
		}
		return
	}

	err = postToMarathonGroup(marGroup)
	if err != nil {
		logrus.Errorf("create group in marathon failed, error is %v", err)
		errorCode = SGO_ERROR_CREATE
		// need to delete sgi and sgo
		_, _, delerr := GetSgiService().UpdateStateById(sgi.ObjectId.Hex(), SGI_STATUS_FAILED, x_auth_token)
		if delerr != nil {
			logrus.Errorf("set sgi to failed err is %v", delerr)
		}
		_, _, delerr = p.UpdateStateBySgiId(sgi.ObjectId.Hex(), SGO_STATUS_FAILED, x_auth_token)
		if delerr != nil {
			logrus.Errorf("set sgo to failed err is %v", delerr)
		}
		return
	}

	newSgo = sgo
	return
}

func (p *SgoService) QueryAll(skip int, limit int, x_auth_token string) (total int,
	sgos []entity.ServiceGroupOrder, errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_sgos", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	logrus.Debugf("auth query is %v", authQuery)

	selector := generateQueryWithAuth(bson.M{}, authQuery)
	logrus.Debugf("selector is %v", selector)
	sort := ""
	sgos = []entity.ServiceGroupOrder{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, sort}
	total, err = dao.HandleQueryAll(&sgos, queryStruct)
	if err != nil {
		logrus.Errorf("list sgo [token=%v] failed, error is %v", x_auth_token, err)
		errorCode = SGO_ERROR_QUERY
	}
	return
}

func (p *SgoService) QueryAllUnterminated(skip int, limit int, x_auth_token string) (total int,
	sgos []entity.ServiceGroupOrder, errorCode string, err error) {
	query := bson.M{}
	query["life_cycle_status"] = bson.M{"$ne": "TERMINATED"}
	return p.queryByQuery(query, skip, limit, x_auth_token, false)
}

func (p *SgoService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int,
	sgos []entity.ServiceGroupOrder, errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth service first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_sgos", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	sgos = []entity.ServiceGroupOrder{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, ""}
	total, err = dao.HandleQueryAll(&sgos, queryStruct)
	if err != nil {
		logrus.Errorf("query sgo by query [%v] error is %v", query, err)
		errorCode = SG_ERROR_QUERY
	}
	return
}

func (p *SgoService) QueryById(objectId string, x_auth_token string) (sgo entity.ServiceGroupOrder,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// do authorize first
	if authorized := GetAuthService().Authorize("get_sgo", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get app with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	sgo = entity.ServiceGroupOrder{}
	err = dao.HandleQueryOne(&sgo, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query sgo [objectId=%v] error is %v", objectId, err)
		errorCode = SGO_ERROR_QUERY
	}
	return
}

func (p *SgoService) DeleteById(objectId, x_auth_token string) (errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// do authorize first
	if authorized := GetAuthService().Authorize("delete_sgo", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete order [%v] error is %v", objectId, err)
		return
	}

	// get order by objId
	sgo, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("get order by id [%v] failed, error is %v", objectId, err)
		errorCode = SGO_ERROR_DELETE
		return
	}
	sgo_oriState := sgo.LifeCycleStatus

	// get sgi by objId
	sgi, _, err := GetSgiService().QueryById(sgo.ServiceGroupInstanceId, x_auth_token)
	if err != nil {
		logrus.Errorf("get sgi by id [%v] failed, error is %v", objectId, err)
		errorCode = SGO_ERROR_DELETE
		return
	}
	sgi_oriState := sgi.LifeCycleStatus

	// change sgo status to terminating
	_, _, err = p.UpdateStateById(objectId, SGO_STATUS_TERMINATING, x_auth_token)
	if err != nil {
		logrus.Errorf("change order status failed, error is %v", err)
		errorCode = SGO_ERROR_DELETE
		return
	}

	// change sgi status to terminating
	_, _, err = GetSgiService().UpdateStateById(sgo.ServiceGroupInstanceId, SGI_STATUS_TERMINATING, x_auth_token)
	if err != nil {
		logrus.Errorf("change sgi status failed, error is %v", err)
		errorCode = SGO_ERROR_DELETE
		return
	}

	// send delete request to marathon to delete group by marathongroupId
	err = deleteToMarathonGroup(sgo.MarathonGroupId)
	if err != nil {
		logrus.Errorf("send delete request to marathon failed, error is %v", err)
		errorCode = SGO_ERROR_DELETE
		// need to change sgo and sgi status to oristate
		_, _, backErr := GetSgiService().UpdateStateById(sgo.ServiceGroupInstanceId, sgi_oriState, x_auth_token)
		if backErr != nil {
			logrus.Errorf("change sgi status failed, error is %v", backErr)
		}
		_, _, backErr = p.UpdateStateById(objectId, sgo_oriState, x_auth_token)
		if backErr != nil {
			logrus.Errorf("change order status failed, error is %v", backErr)
		}
		return
	}
	// change sgi status to terminated
	_, _, err = GetSgiService().UpdateStateById(sgo.ServiceGroupInstanceId, SGI_STATUS_TERMINATED, x_auth_token)
	if err != nil {
		logrus.Errorf("change sgi status to terminated failed, error is %v", err)
		errorCode = SGO_ERROR_DELETE
		return
	}
	// change sgo status to terminated
	_, _, err = p.UpdateStateById(objectId, SGO_STATUS_TERMINATED, x_auth_token)
	if err != nil {
		logrus.Errorf("change order status to terminated failed, error is %v", err)
		errorCode = SGO_ERROR_DELETE
		return
	}
	return
}

func (p *SgoService) ScaleAppByOrderId(objectId, appPath, appNum,
	x_auth_token string) (errorCode string, err error) {
	// validate parameter
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	num, err := strconv.Atoi(appNum)
	if err != nil {
		logrus.Errorf("convert app number [%v] failed, error is %v", appNum, err)
		err = errors.New("invalide app number.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// appNum can not less than minNumber,
	// and can not greater than maxNumber,
	// and match step interval
	app, _, err := p.GetAppInOrder(appPath, objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("get app [%v] from order [%v] failed, error is %v", appPath, objectId, err)
		errorCode = SGO_ERROR_SCALE
		return
	}

	if !app.Scale.Enabled {
		logrus.Errorf("scale app [%v] in order [%v] failed, caused by this app cannot be scaled")
		err = errors.New("this app can not be scaled")
		errorCode = SGO_ERROR_SCALE
		return
	}

	if num < app.Scale.MinNum || num > app.Scale.MaxNum {
		err = errors.New("the number of app must be in range [" +
			strconv.Itoa(app.Scale.MinNum) + "-" + strconv.Itoa(app.Scale.MaxNum) + "]")
		errorCode = COMMON_ERROR_INVALIDATE
		logrus.Errorf("scale app [%v] in order [%v] failed, error is %v", appPath, objectId, err)
		return
	}
	//the size must be an integral multiple of the scale step
	if matchStep := (num - app.Scale.MinNum) % app.Scale.ScaleStep; matchStep != 0 {
		err = errors.New("the scale number of app must be an integral multiple of " + strconv.Itoa(app.Scale.ScaleStep))
		errorCode = COMMON_ERROR_INVALIDATE
		logrus.Errorf("scale app [%v] in order [%v] failed, error is %v", appPath, objectId, err)
		return
	}

	// do authorize first, scaleapp_sgo
	if authorized := GetAuthService().Authorize("scaleapp_sgo", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("scale app [%v] in sgo [%v] error is %v", appPath, objectId, err)
		return
	}
	// find sgo
	sgo, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("find sgo [%v] failed, error is %v", objectId, err)
		errorCode = SGO_ERROR_SCALE
		return
	}
	// quota check
	errorCode, err = GetQuotaService().CheckScale(&sgo, num)
	if err != nil {
		logrus.Errorf("check quota failed, error is %v", err)
		return
	}

	// change sgo status to modifying
	_, _, err = p.UpdateStateById(objectId, SGO_STATUS_MODIFYING, x_auth_token)
	if err != nil {
		logrus.Errorf("update sgo [%v] status to modifying failed, error is %v", objectId, err)
		errorCode = SGO_ERROR_SCALE
		return
	}

	// change sgi's status to modifying and change app instances num
	defer GetLockService().ReleaseInstanceLock(sgo.ServiceGroupInstanceId)
	if locked := GetLockService().CreateInstanceLock(sgo.ServiceGroupInstanceId); locked {
		sgi, _, updateErr := GetSgiService().QueryById(sgo.ServiceGroupInstanceId, x_auth_token)
		if updateErr != nil {
			err = updateErr
			logrus.Errorf("find sgi by id [%v] failed, error is %v", sgo.ServiceGroupInstanceId, err)
			errorCode = SGO_ERROR_SCALE
			return
		}
		// change app instances
		flag, updateErr := GetSgiService().ChangeInstanceNumber(&sgi, appPath, num)
		if updateErr != nil {
			err = updateErr
			logrus.Errorf("change app [%v] num [%v] in sgi [%v] failed, error is %v",
				appPath, num, sgi, err)
			errorCode = SGO_ERROR_SCALE
			return
		}
		if !flag {
			err = errors.New("can not update the instances number in service group instance")
			logrus.Errorf("can not update the instances number in service group instance")
			errorCode = SGO_ERROR_SCALE
			return
		}
		// modify sgi status
		if sgi.LifeCycleStatus != SGI_STATUS_REPARING {
			sgi.LifeCycleStatus = SGI_STATUS_MODIFYING
		}

		sgi.TimeUpdate = dao.GetCurrentTime()
		_, _, updateErr = GetSgiService().UpdateById(sgi.ObjectId.Hex(), sgi, x_auth_token)
		if updateErr != nil {
			err = updateErr
			logrus.Errorf("update sgi [%v] failed, error is %v", sgi, err)
			errorCode = SGO_ERROR_SCALE
			return

		}
	} else {
		// throw error
		err = errors.New("can't update app instance number to service group instance, maybe locked")
		errorCode = SGO_ERROR_SCALE
		logrus.Errorf("refine appInstance err is %v", err)
		return
	}

	// invoke marathon app api to do scale operation
	marathonAppId := entity.GetMarathonAppIdFromSgo(&sgo, appPath)
	app.Id = marathonAppId
	app.Instances = num
	app.Env["LINKER_SERVICE_GROUP_INSTANCE_ID"] = sgo.ServiceGroupInstanceId
	app.Env["LINKER_SERVICE_ORDER_ID"] = sgo.ObjectId.Hex()
	app.Env["LINKER_SERVICE_OFFERING_ID"] = sgo.ServiceOfferingId
	app.Env["LINKER_SERVICE_GROUP_ID"] = sgo.ServiceGroupId
	appjson, _ := removeDockerContainerInfoInApp(app)

	err = putToMarathonApp(app.Id, appjson)
	if err != nil {
		logrus.Errorf("put app [%v] to marathon failed, error is %v", app, err)
		errorCode = SGO_ERROR_SCALE
		//TODO: should callback sgi, sgo status and sgi's app instances num.
	}
	return
}

func (p *SgoService) GetAppInOrder(appPath, orderId, x_auth_token string) (app *entity.App,
	errorCode string, err error) {
	// validate orderId
	if !bson.IsObjectIdHex(orderId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// get sgo
	sgo, errorCode, err := p.QueryById(orderId, x_auth_token)
	if err != nil {
		logrus.Errorf("query order by objectId [%v] failed, error is %v", orderId, err)
		return
	}
	// get so
	so, _, err := GetSoService().QueryById(sgo.ServiceOfferingId, x_auth_token)
	if err != nil {
		logrus.Errorf("query offering by objectId [%v] failed, error is %v", sgo.ServiceOfferingId, err)
		errorCode = SGO_ERROR_QUERY
		return
	}

	// get app from so
	app, err = entity.GetAppFromServiceGroup(&so, appPath)
	if err != nil {
		logrus.Errorf("get app [%v] from offering [%v] failed, error is %v", appPath, so, err)
		errorCode = SGO_ERROR_QUERY
	}
	return
}

func (p *SgoService) GetOperationById(objectId string, x_auth_token string) (operations map[string]int,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	operationList := []string{"delete_sgo", "scaleapp_sgo", "metering_sgo"}
	operations, err = GetAuthService().AuthOperation(operationList, x_auth_token, objectId, p.collectionName)
	if err != nil {
		logrus.Errorf("get auth operation of [objectId=%v] error is %v", objectId, err)
		errorCode = COMMON_ERROR_INTERNAL
	}
	return
}

func (p *SgoService) UpdateStateById(objectId string, newState string, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update sgo by objectId [%v] status to %v", objectId, newState)
	// validate objectId
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// get sgo by objectId
	query := bson.M{}
	query["_id"] = bson.ObjectIdHex(objectId)
	sgo, _, err := p.queryOneByQuery(query)
	if err != nil {
		logrus.Errorf("get sgo by objectId [%v] failed, error is %v", objectId, err)
		return
	}

	// do authorize
	if authorized := GetAuthService().Authorize("update_sgo", x_auth_token, sgo.ObjectId.Hex(), p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update sgo with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
		return
	}

	if sgo.LifeCycleStatus == newState {
		err = errors.New("this order is still in " + newState + " state.")
		errorCode = SGO_ERROR_UPDATE
		logrus.Errorf("this sgo [%v] is already in state [%v]", sgo, newState)
		return
	}

	var selector = bson.M{}
	selector["_id"] = sgo.ObjectId

	// updated_time
	sgo.LifeCycleStatus = newState
	sgo.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&sgo, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update sgo with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
	}
	return
}

func (p *SgoService) UpdateStateBySgiId(sgiId string, newState string, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update sgo by sgiId [%v] status to %v", sgiId, newState)
	// validate sgiId
	if !bson.IsObjectIdHex(sgiId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// get sgo by sgiId
	query := bson.M{}
	query["service_group_instance_id"] = sgiId
	sgo, _, err := p.queryOneByQuery(query)
	if err != nil {
		logrus.Errorf("get sgo by sgiId [%v] failed, error is %v", sgiId, err)
		return
	}

	// do authorize
	if authorized := GetAuthService().Authorize("update_sgo", x_auth_token, sgo.ObjectId.Hex(), p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update sgo with sgiId [%v] status to [%v] failed, error is %v", sgiId, newState, err)
		return
	}

	if sgo.LifeCycleStatus == newState {
		logrus.Infof("this sgo [%v] is already in state [%v]", sgo, newState)
		return false, "", nil
	}

	var selector = bson.M{}
	selector["_id"] = sgo.ObjectId

	// updated_time
	sgo.LifeCycleStatus = newState
	sgo.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&sgo, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update sgo with sgiId [%v] status to [%v] failed, error is %v", sgiId, newState, err)
	}
	return
}

func (p *SgoService) queryOneByQuery(query bson.M) (sgo entity.ServiceGroupOrder,
	errorCode string, err error) {
	sgo = entity.ServiceGroupOrder{}
	err = dao.HandleQueryOne(&sgo, dao.QueryStruct{p.collectionName, query, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query sgo [query=%v] error is %v", query, err)
		errorCode = SGO_ERROR_QUERY
	}
	return
}
