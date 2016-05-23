package services

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"strconv"
	"strings"
	"sync"
)

var (
	repairPolicyService           *RepairPolicyService = nil
	onceRepairPolicy              sync.Once
	REPAIR_POLICY_ERROR_CREATE    string = "E11130"
	REPAIR_POLICY_ERROR_UPDATE    string = "E11131"
	REPAIR_POLICY_ERROR_DELETE    string = "E11132"
	REPAIR_POLICY_ERROR_QUERY     string = "E11133"
	REPAIR_RECORD_ERROR_CREATE    string = "E11134"
	REPAIR_RECORD_ERROR_UPDATE    string = "E11135"
	REPAIR_RECORD_ERROR_QUERY     string = "E11136"
	REPAIR_ERROR                  string = "E11137"
	repairRecordCollection        string = "repairRecord"
	repairPolicyCollection        string = "repairPolicy"
	REPAIR_ACTION_TYPE_SCALE      string = "SCALE"
	REPAIR_ACTION_TYPE_SCALE_STEP string = "SCALESTEP"
	REPAIR_ACTION_FAILURE         string = "REPAIR_ACTION_FAILURE"
	REPAIR_ACTION_SUCCESS         string = "REPAIR_ACTION_SUCCESS"
)

type RepairPolicyService struct {
	collectionName string
}

func GetRepairPolicyService() *RepairPolicyService {
	onceRepairPolicy.Do(func() {
		logrus.Debugf("Once called from repairPolicyService ......................................")
		repairPolicyService = &RepairPolicyService{repairPolicyCollection}
	})
	return repairPolicyService
}

func (p *RepairPolicyService) Create(repairPolicy entity.RepairPolicy, x_auth_token string) (newRepairPolicy entity.RepairPolicy,
	errorCode string, err error) {
	logrus.Infof("start to create repairPolicy [%v]", repairPolicy)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_repairpolicy", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create repairPolicy [%v] error is %v", repairPolicy, err)
		return
	}
	// generate ObjectId
	repairPolicy.ObjectId = bson.NewObjectId()

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = REPAIR_POLICY_ERROR_CREATE
		logrus.Errorf("get token failed when create repairPolicy [%v], error is %v", repairPolicy, err)
		return
	}

	// set token_id and user_id from token
	repairPolicy.Tenant_id = token.Tenant.Id
	repairPolicy.User_id = token.User.Id

	// set created_time and updated_time
	repairPolicy.TimeCreate = dao.GetCurrentTime()
	repairPolicy.TimeUpdate = repairPolicy.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, repairPolicy)
	if err != nil {
		errorCode = REPAIR_POLICY_ERROR_CREATE
		logrus.Errorf("create repairPolicy [%v] to bson error is %v", repairPolicy, err)
		return
	}

	newRepairPolicy = repairPolicy
	return
}

func (p *RepairPolicyService) UpdateById(objectId string, repairPolicy entity.RepairPolicy, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update repairPolicy [%v]", repairPolicy)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_repairpolicy", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update repairPolicy with objectId [%v] error is %v", objectId, err)
		return
	}
	// validate repairPolicy
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	repairPolicy.ObjectId = bson.ObjectIdHex(objectId)
	repairPolicy.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&repairPolicy, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update repairPolicy [%v] error is %v", repairPolicy, err)
		errorCode = REPAIR_POLICY_ERROR_UPDATE
	}
	return
}

func (p *RepairPolicyService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete repairPolicy with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_repairpolicy", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete repairPolicy with objectId [%v] error is %v", objectId, err)
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
		logrus.Errorf("delete repairPolicy [objectId=%v] error is %v", objectId, err)
		errorCode = REPAIR_POLICY_ERROR_DELETE
	}
	return
}

func (p *RepairPolicyService) QueryAll(skip int, limit int, x_auth_token string) (total int, repairPolicys []entity.RepairPolicy,
	errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_repairpolicies", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	logrus.Debugf("auth query is %v", authQuery)

	selector := generateQueryWithAuth(bson.M{}, authQuery)
	logrus.Debugf("selector is %v", selector)
	sort := ""
	repairPolicys = []entity.RepairPolicy{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, sort}
	total, err = dao.HandleQueryAll(&repairPolicys, queryStruct)
	if err != nil {
		logrus.Errorf("list repairPolicy [token=%v] failed, error is %v", x_auth_token, err)
		errorCode = REPAIR_POLICY_ERROR_QUERY
	}
	return
}

func (p *RepairPolicyService) QueryById(objectId string, x_auth_token string) (repairPolicy entity.RepairPolicy,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// do authorize first
	if authorized := GetAuthService().Authorize("get_repairpolicy", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get repairPolicy with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	repairPolicy = entity.RepairPolicy{}
	err = dao.HandleQueryOne(&repairPolicy, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query repairPolicy [objectId=%v] error is %v", objectId, err)
		errorCode = REPAIR_POLICY_ERROR_QUERY
	}
	return
}

func (p *RepairPolicyService) GetOperationById(objectId string, x_auth_token string) (operations map[string]int,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	operationList := []string{"update_repairpolicy", "delete_repairpolicy"}
	operations, err = GetAuthService().AuthOperation(operationList, x_auth_token, objectId, p.collectionName)
	if err != nil {
		logrus.Errorf("get auth operation of [objectId=%v] error is %v", objectId, err)
		errorCode = COMMON_ERROR_INTERNAL
	}
	return
}

/*
Get Alert and find the correct repair actions
*/
func (p *RepairPolicyService) AnalyzeAlert(serviceGroupId, AppCointainerId,
	serviceGroupInsanceId, orderId, alertId, alertName, alertValue string) (errorCode string, err error) {
	// get the xtoken
	x_auth_token, err := GenerateToken()
	if err != nil {
		logrus.Errorf("get token for repair [serviceGroupId=%v] and [AppCointainerId=%v] error is %v", serviceGroupId, AppCointainerId, err)
		GetAlertService().NotifyRepairResult(alertId, SGI_STATUS_FAILED)
		return
	}

	//find the repair policy
	_, polices, _, err := queryRepairPolicy(serviceGroupId, AppCointainerId, x_auth_token)
	if err != nil || len(polices) <= 0 {
		logrus.Errorf("query policy for repair [serviceGroupId=%v] and [AppCointainerId=%v] and [x_auth_token=%v] error is %v", serviceGroupId, AppCointainerId, x_auth_token, err)
		GetAlertService().NotifyRepairResult(alertId, SGI_STATUS_FAILED)
		return
	}
	repairPolicy := polices[0]

	action := entity.RepairAction{}

	//find the detailed action,current only support the first action
	for _, policy := range repairPolicy.Polices {
		for _, condition := range policy.Conditions {
			if condition.Name == alertName {
				action = policy.Actions[0]
			}
		}
	}

	//call the do repair operation
	p.doRepairOperation(serviceGroupId, alertId, alertName, orderId, serviceGroupInsanceId, action, x_auth_token)
	return
}

/*
Doing Repair Operation
*/
func (p *RepairPolicyService) doRepairOperation(serviceGroupId, alertId string, alertName string, orderId string,
	serviceGroupInsanceId string, action entity.RepairAction, x_auth_token string) (errorCode string, err error) {
	if action.Type == REPAIR_ACTION_TYPE_SCALE {
		//get the current status
		sgi, _, err1 := GetSgiService().QueryById(serviceGroupInsanceId, x_auth_token)

		if err1 != nil {
			errorCode = REPAIR_ERROR
			err = errors.New("repair action failed while loading the instance!")
			logrus.Errorf("query instance for repair [objectId=%v] error is %v", serviceGroupInsanceId, err)
			GetAlertService().NotifyRepairResult(alertId, SGI_STATUS_FAILED)
			return
		}
		currentStatus := sgi.LifeCycleStatus

		//start the scaleout repair
		appPath := action.AppContainerId
		appNum := ""

		for _, parameter := range action.Parameters {
			if parameter.Name == REPAIR_ACTION_TYPE_SCALE_STEP {
				appNum = parameter.Value
			}
		}

		if len(appNum) <= 0 {
			return
		}

		//insert repair record
		repairRecord := entity.RepairRecord{}
		repairRecord.AppCointainerId = appPath
		repairRecord.ServiceGroupId = serviceGroupId
		repairRecord.ServiceGroupInstanceId = serviceGroupInsanceId
		repairRecord.Status = SGI_STATUS_REPARING
		repairRecord.AlertId = alertId
		repairRecord.AlertName = alertName
		repairRecord.Action = strings.Join([]string{REPAIR_ACTION_TYPE_SCALE, appNum}, ":")

		newRepairRecord, _, err1 := createRepairRecord(repairRecord)

		if err1 != nil {
			err = errors.New("repair action failed while save repair record!")
			errorCode = REPAIR_ERROR
			logrus.Errorf("create repair record failed [instanceid=%v] error is %v", serviceGroupInsanceId, err)
			GetAlertService().NotifyRepairResult(alertId, SGI_STATUS_FAILED)
			return
		}

		//update the instance lifecycle to repairing
		if locked := GetLockService().CreateInstanceLock(serviceGroupInsanceId); locked {
			_, errorCode, err = GetSgiService().UpdateRepairIdAndStatusById(serviceGroupInsanceId, newRepairRecord.RepairId, SGI_STATUS_REPARING, x_auth_token)

			GetLockService().ReleaseInstanceLock(serviceGroupInsanceId)

			if err != nil {
				GetLockService().ReleaseInstanceLock(serviceGroupInsanceId)
				logrus.Errorf("update instance for repairing [objectId=%v] error is %v", serviceGroupInsanceId, err)
				GetAlertService().NotifyRepairResult(alertId, SGI_STATUS_FAILED)
				return
			}
		} else {
			err = errors.New("can't update instance to reparing, maybe locked")
			errorCode = REPAIR_ERROR
			logrus.Errorf("can't update instance to reparing err is %v", err)
			GetAlertService().NotifyRepairResult(alertId, SGI_STATUS_FAILED)
			return
		}

		//start scale
		number, err2 := caculateNewAppNumber(sgi, appPath, appNum)
		if err2 != nil {
			err = errors.New("repair action failed while caculate the number for scale")
			errorCode = REPAIR_ERROR
			logrus.Errorf("can't caculate new instance number for reparing scaling  err is %v", err)
			GetAlertService().NotifyRepairResult(alertId, SGI_STATUS_FAILED)
			return
		}

		errorCode, err = GetSgoService().ScaleAppByOrderId(orderId, appPath, number, x_auth_token)

		//if repair failed change the instance status back
		if err != nil {
			logrus.Errorf("scale instance for repairing [orderId=%v] ,[appPath=%v] error is %v", orderId, appPath, err)
			//update the instance lifecycle to repairing
			if locked := GetLockService().CreateInstanceLock(serviceGroupInsanceId); locked {
				_, errorCode, err = GetSgiService().UpdateRepairIdAndStatusById(serviceGroupInsanceId, "", currentStatus, x_auth_token)
				if err != nil {
					logrus.Errorf("update instance for repairing [objectId=%v] to previous status error is %v", serviceGroupInsanceId, err)
				}
				GetLockService().ReleaseInstanceLock(serviceGroupInsanceId)
			} else {
				err = errors.New("can't update instance to back from reparing, maybe locked")
				errorCode = REPAIR_ERROR
				logrus.Errorf("can't update instance to back from reparing err is %v", err)
			}

			p.AnalyzeNotify(REPAIR_ACTION_FAILURE, newRepairRecord.RepairId)
			return
		}
	} else {
		logrus.Errorf("can't find the repaie action type, type is %v", action.Type)
		GetAlertService().NotifyRepairResult(alertId, SGI_STATUS_FAILED)
		return
	}
	return
}

/*
Get instance repair finished Notify and Notify repair and Alert Finished
*/
func (p *RepairPolicyService) AnalyzeNotify(repairStatus string, repairid string) (errorCode string, err error) {
	//split the repairId to alert Id and recordId
	valueArrays := strings.Split(repairid, "-")
	if len(valueArrays) != 2 {
		logrus.Errorln("a generic repairid format error! value is %v:", repairid)
		return
	}
	alertId := valueArrays[0]
	recordId := valueArrays[1]

	//update repair record status to finished
	_, _, err = updateStatusById(recordId, repairStatus)
	if err != nil {
		errorCode = REPAIR_ERROR
		logrus.Errorf("can't update record record with reparing result err is %v", err)
		GetAlertService().NotifyRepairResult(alertId, REPAIR_ACTION_FAILURE)
		return
	}

	//call the alert of reparing finished/Failed
	GetAlertService().NotifyRepairResult(alertId, repairStatus)

	return
}

func createRepairRecord(repairRecord entity.RepairRecord) (newRepairRecord entity.RepairRecord,
	errorCode string, err error) {
	logrus.Infof("start to create repairRecord [%v]", repairRecord)

	// generate ObjectId
	repairRecord.ObjectId = bson.NewObjectId()
	repairRecord.RepairId = strings.Join([]string{repairRecord.AlertId, repairRecord.ObjectId.Hex()}, "-")

	// set created_time and updated_time
	repairRecord.TimeCreate = dao.GetCurrentTime()
	repairRecord.TimeUpdate = repairRecord.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(repairRecordCollection, repairRecord)
	if err != nil {
		errorCode = REPAIR_RECORD_ERROR_CREATE
		logrus.Errorf("create repairRecord [%v] to bson error is %v", repairRecord, err)
		return
	}

	newRepairRecord = repairRecord
	return
}

func queryRecordById(objectId string) (repairRecord entity.RepairRecord,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	repairRecord = entity.RepairRecord{}
	err = dao.HandleQueryOne(&repairRecord, dao.QueryStruct{repairRecordCollection, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query repairRecord [objectId=%v] error is %v", objectId, err)
		errorCode = REPAIR_RECORD_ERROR_QUERY
	}
	return
}

func updateStatusById(recordId string, status string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update repairRecord [%v]", recordId)

	// validate recordId
	if !bson.IsObjectIdHex(recordId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// get record by recordId
	record, _, err := queryRecordById(recordId)
	if err != nil {
		logrus.Errorf("get record by recordId [%v] failed, error is %v", recordId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(recordId)

	// reset objectId and updated_time
	record.ObjectId = bson.ObjectIdHex(recordId)
	record.Status = status
	record.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&record, dao.QueryStruct{repairRecordCollection, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update record with recordId [%v] status to [%v] failed, error is %v", recordId, status, err)
	}
	return
}

func queryRepairPolicy(serviceGroupId, appContainerId string, x_auth_token string) (total int, polices []entity.RepairPolicy,
	errorCode string, err error) {
	// do authorize first
	authQuery, err := GetAuthService().BuildQueryByAuth("get_repairpolicy", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	var selector = bson.M{}
	selector["service_group_id"] = serviceGroupId
	selector["app_container_id"] = appContainerId
	selector = generateQueryWithAuth(selector, authQuery)

	queryStruct := dao.QueryStruct{repairPolicyCollection, selector, 0, 0, ""}
	total, err = dao.HandleQueryAll(&polices, queryStruct)
	logrus.Debugf("queryRepairPolicy Total is %v", total)
	if err != nil {
		logrus.Errorf("get repairPolicy with serviceGroupId [%v] and appContainerId [%v] error is %v", serviceGroupId, appContainerId, err)
		errorCode = REPAIR_POLICY_ERROR_QUERY
	}
	return
}

func caculateNewAppNumber(sgi entity.ServiceGroupInstance, appId, step string) (numberStr string, err error) {
	app, err := entity.FindAppInSgi(&sgi, appId)
	if err != nil {
		logrus.Errorf("Can not find the App in the service group instance, err is %v", err)
		return
	}

	stepNumber, err := strconv.Atoi(step)
	if err != nil {
		logrus.Errorf("convert step number [%v] failed, error is %v", step, err)
		return
	}
	number := app.Instances + stepNumber
	numberStr = strconv.Itoa(number)

	return
}
