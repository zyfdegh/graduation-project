package services

import (
	"encoding/json"
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/jsonq"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"linkernetworks.com/linker_common_lib/httpclient"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_controller/common"
	"strconv"
	"strings"
	"sync"
)

var (
	aciService *AciService = nil
	onceAci    sync.Once

	ACI_STATUS_CREATED    = "CREATED"
	ACI_STATUS_CONFIGED   = "CONFIGED"
	ACI_STATUS_UNCONFIGED = "UNCONFIGED"
	ACI_STATUS_TERMINATED = "TERMINATED"
	ACI_STATUS_FAILED     = "FAILED"

	ACI_STATUS = []string{ACI_STATUS_CREATED, ACI_STATUS_CONFIGED,
		ACI_STATUS_UNCONFIGED, ACI_STATUS_TERMINATED, ACI_STATUS_FAILED}

	ACI_ERROR_CREATE   string = "E11001"
	ACI_ERROR_UPDATE   string = "E11003"
	ACI_ERROR_DELETE   string = "E11002"
	ACI_ERROR_QUERY    string = "E11004"
	ACI_ERROR_ALLOCATE string = "E11005"
)

type AciService struct {
	collectionName string
}

func GetAciService() *AciService {
	onceAci.Do(func() {
		logrus.Debugf("Once called from aciService ......................................")
		aciService = &AciService{"app_container_instance"}
	})
	return aciService
}

func (p *AciService) AllocateResource(objectId, allocateRegx, x_auth_token string) (allocatedStr,
	errorCode string, err error) {
	logrus.Infof("start to allocate resource [%v] to aci [%v]", allocateRegx, objectId)
	// FIXME: skip authorize here, should consider resource type then do authorize
	// if authorized := GetAuthService().Authorize("create_aci", x_auth_token, "", p.collectionName); !authorized {
	// 	err = errors.New("required opertion is not authorized!")
	// 	errorCode = COMMON_ERROR_UNAUTHORIZED
	// 	logrus.Errorf("create aci [%v] error is %v", aci, err)
	// 	return
	// }
	// validate aci
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide aci ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	allocatedStr, err = _getAllocatedInfo(allocateRegx, objectId)
	if err != nil {
		logrus.Errorf("allocate resource by allocateRegx [%v] failed, error is %v", allocateRegx, err)
		errorCode = ACI_ERROR_ALLOCATE
	}
	return

}

func (p *AciService) Create(aci entity.AppContainerInstance, x_auth_token string) (newApp entity.AppContainerInstance,
	errorCode string, err error) {
	logrus.Infof("start to create aci [%v]", aci)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_aci", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create aci [%v] error is %v", aci, err)
		return
	}
	// generate ObjectId
	aci.ObjectId = bson.NewObjectId()

	// set created_time and updated_time
	aci.TimeCreate = dao.GetCurrentTime()
	aci.TimeUpdate = aci.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, aci)
	if err != nil {
		errorCode = ACI_ERROR_CREATE
		logrus.Errorf("create aci [%v] to bson error is %v", aci, err)
		return
	}

	// refine aci into service group instance
	defer GetLockService().ReleaseInstanceLock(aci.ServiceGroupInstanceId)
	if locked := GetLockService().CreateInstanceLock(aci.ServiceGroupInstanceId); locked {
		// get service group instance
		sgi, _, refineError := GetSgiService().QueryById(aci.ServiceGroupInstanceId, x_auth_token)
		if refineError != nil {
			err = refineError
			logrus.Errorf("refine groupInstance %s refineError is %v",
				aci.ServiceGroupInstanceId, err)
			errorCode = ACI_ERROR_CREATE
			return
		}
		// check current refined instanceids and do replace if needed
		index := strings.LastIndex(aci.AppContainerId, "/")
		refinedGroupId := aci.AppContainerId[:index]
		refineError = p.checkCurrentRefine(&sgi, refinedGroupId, aci.MarathonAppPath, aci.ObjectId.Hex())
		if refineError != nil {
			err = refineError
			logrus.Errorf("check current refine err is %v", err)
			errorCode = ACI_ERROR_CREATE
			return
		}
		// save sgi to db
		_, _, refineError = GetSgiService().UpdateById(aci.ServiceGroupInstanceId, sgi, x_auth_token)
		if refineError != nil {
			err = refineError
			logrus.Errorf("update sgi failed, error is %v", err)
			errorCode = ACI_ERROR_CREATE
			return
		}
	} else {
		err = errors.New("can't refine app instance to service group instance, maybe locked")
		errorCode = ACI_ERROR_CREATE
		logrus.Errorf("refine appInstance err is %v", err)
		return
	}
	newApp = aci
	return
}

func (p *AciService) UpdateById(objectId string, aci entity.AppContainerInstance, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update aci [%v]", aci)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_aci", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update aci with objectId [%v] error is %v", objectId, err)
		return
	}
	// save aci to db
	created, errorCode, err = p.updateById(objectId, &aci)
	if err != nil {
		return
	}

	// release ip if newaci's ip is empty
	if len(aci.DockerContainerIp) <= 0 {
		logrus.Infof("start to release ip from aci [%v]", aci)
		// release ip
		releaseIpErr := GetIPResourceService().releaseIp(aci.ObjectId.Hex())
		if releaseIpErr != nil {
			logrus.Warningf("release ip from aci [%v] failed, error is %v",
				aci.ObjectId.Hex(), releaseIpErr)
		}
	}

	// check aci status, if it's status is not CREATED, do notify
	// otherwise, return
	if aci.LifeCycleStatus == ACI_STATUS_CREATED {
		return
	}
	sgi, _, err := GetSgiService().QueryById(aci.ServiceGroupInstanceId, x_auth_token)
	if err != nil {
		logrus.Errorf("can not get sgi [%v], error is %v", aci.ServiceGroupInstanceId, err)
		errorCode = APP_ERROR_UPDATE
		return
	}
	err = doNotifyByAci(&aci, &sgi, "updateNotify", x_auth_token)
	if err != nil {
		logrus.Errorf("do notify sgi [%v] caused by aci [%v] updated failed, error is %v",
			sgi, aci, err)
		errorCode = ACI_ERROR_UPDATE
		return
	}
	// update sgi and sgo status
	updateSgoStatus(&sgi, x_auth_token)
	return
}

func (p *AciService) updateById(objectId string, aci *entity.AppContainerInstance) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update aci [%v]", aci)
	// validate aci
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	aci.ObjectId = bson.ObjectIdHex(objectId)
	aci.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&aci, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update aci [%v] error is %v", aci, err)
		errorCode = APP_ERROR_UPDATE
	}
	return
}

func (p *AciService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete aci with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_aci", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete aci with objectId [%v] error is %v", objectId, err)
		return
	}

	// validate objectId first
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// anyway, should release ip and refine sgi
	// then check sgi status
	// if sgi is terminating, no need to notify
	// if sgi is not terminate, do notify, and update sgi, sgo status
	aci, _, err := p.queryById(objectId)
	if err != nil {
		logrus.Errorf("query aci by object [%d] failed, error is %v", objectId, err)
		errorCode = ACI_ERROR_DELETE
		return
	}

	// finally, update aci status
	defer func() {
		//TODO: if aci status is not FAILED, update aci status to TERMINATED
		p.updateAciStatus(objectId, ACI_STATUS_TERMINATED)
	}()

	// release ip
	releaseIpErr := GetIPResourceService().releaseIp(aci.ObjectId.Hex())
	if releaseIpErr != nil {
		logrus.Warningf("release ip from aci [%v] failed, error is %v",
			aci.ObjectId.Hex(), releaseIpErr)
	}

	// refine sgi
	sgiId := aci.ServiceGroupInstanceId
	sgi := entity.ServiceGroupInstance{}
	defer GetLockService().ReleaseInstanceLock(sgiId)
	if locked := GetLockService().CreateInstanceLock(sgiId); locked {
		// get sgi
		sgi, _, refineError := GetSgiService().QueryById(sgiId, x_auth_token)
		if refineError != nil {
			err = refineError
			logrus.Errorf("refine sgi with id [%v] failed, error is %v", sgiId, err)
			errorCode = ACI_ERROR_DELETE
			return
		}
		index := strings.LastIndex(aci.AppContainerId, "/")
		refinedGroupId := aci.AppContainerId[:index]
		// refine sgi
		_, refineError = p.removeAciFromSgi(&sgi, refinedGroupId, aci.MarathonAppPath, aci.ObjectId.Hex())
		if refineError != nil {
			err = refineError
			logrus.Errorf("remove aci [%v] from sgi [%v] failed, errror is %v", objectId, sgi, err)
			errorCode = ACI_ERROR_DELETE
			return
		}
		_, _, refineError = GetSgiService().UpdateById(aci.ServiceGroupInstanceId, sgi, x_auth_token)
		if refineError != nil {
			err = refineError
			logrus.Errorf("update sgi failed, error is %v", err)
			errorCode = ACI_ERROR_DELETE
			return
		}
	} else {
		err = errors.New("can't remove app instance from service group instance, maybe locked")
		logrus.Errorf("remove appInstance from servic group err is %v", err)
	}

	// check sgi status, if not terminating and terminated, do notify
	sgi, _, refineError := GetSgiService().QueryById(sgiId, x_auth_token)
	if refineError != nil {
			err = refineError
			logrus.Errorf("refine sgi with id [%v] failed, error is %v", sgiId, err)
			errorCode = ACI_ERROR_DELETE
			return
	}
	
	if !strings.EqualFold(sgi.LifeCycleStatus, SGI_STATUS_TERMINATING) &&
		!strings.EqualFold(sgi.LifeCycleStatus, SGI_STATUS_TERMINATED) {
		logrus.Infof("sgi is not terminating, should do notify, aciId [%v]", objectId)
		err = doNotifyByAci(&aci, &sgi, "deleteNotify", x_auth_token)
		if err != nil {
			logrus.Errorf("do notify sgi [%v] caused by aci [%v] deleted failed, error is %v",
				sgi, aci, err)
			errorCode = ACI_ERROR_DELETE
			return
		}

		// update sgi and sgo status
		updateSgoStatus(&sgi, x_auth_token)
	} // end of terminate check
	return
}

func (p *AciService) QueryAll(skip int, limit int, x_auth_token string) (total int, acis []entity.AppContainerInstance,
	errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_acis", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	logrus.Debugf("auth query is %v", authQuery)

	selector := generateQueryWithAuth(bson.M{}, authQuery)
	logrus.Debugf("selector is %v", selector)
	sort := ""
	acis = []entity.AppContainerInstance{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, sort}
	total, err = dao.HandleQueryAll(&acis, queryStruct)
	if err != nil {
		logrus.Errorf("list aci [token=%v] failed, error is %v", x_auth_token, err)
		errorCode = ACI_ERROR_QUERY
	}
	return
}

func (p *AciService) QueryById(objectId string, x_auth_token string) (aci entity.AppContainerInstance,
	errorCode string, err error) {
	// do authorize first
	if authorized := GetAuthService().Authorize("get_aci", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get aci with objectId [%v] error is %v", objectId, err)
		return
	}

	return p.queryById(objectId)
}

func (p *AciService) queryById(objectId string) (aci entity.AppContainerInstance,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	aci = entity.AppContainerInstance{}
	err = dao.HandleQueryOne(&aci, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query aci [objectId=%v] error is %v", objectId, err)
		errorCode = ACI_ERROR_QUERY
	}
	return
}

// func (p *AciService) GetOperationById(objectId string, x_auth_token string) (operations map[string]int,
// 	errorCode string, err error) {
// 	if !bson.IsObjectIdHex(objectId) {
// 		err = errors.New("invalide ObjectId.")
// 		errorCode = COMMON_ERROR_INVALIDATE
// 		return
// 	}

// 	operationList := []string{"update_aci", "delete_aci"}
// 	operations, err = GetAuthService().AuthOperation(operationList, x_auth_token, objectId, p.collectionName)
// 	if err != nil {
// 		logrus.Errorf("get auth operation of [objectId=%v] error is %v", objectId, err)
// 		errorCode = COMMON_ERROR_INTERNAL
// 	}
// 	return
// }

func (p *AciService) queryAllByQuery(skip int, limit int, query bson.M) (total int, acis []entity.AppContainerInstance,
	errorCode string, err error) {
	sort := ""
	acis = []entity.AppContainerInstance{}
	queryStruct := dao.QueryStruct{p.collectionName, query, skip, limit, sort}
	total, err = dao.HandleQueryAll(&acis, queryStruct)
	if err != nil {
		logrus.Errorf("list aci by query [%v] failed, error is %v", query, err)
		errorCode = ACI_ERROR_QUERY
	}
	return
}

func (p *AciService) checkCurrentRefine(sgi *entity.ServiceGroupInstance, groupId, appId, newId string) (err error) {
	app, err := entity.FindAppInSgi(sgi, appId)
	if err != nil {
		logrus.Errorf("Can not find the App in the service group instance, err is %v", err)
		return
	}
	// check whether instanceIds num is less than expected instances
	// if yes, append newId to instanceIds
	// if no, check which intance is failed, and replace it
	if len(app.InstanceIds) < app.Instances {
		app.InstanceIds = append(app.InstanceIds, newId)
	} else {
		mesosTaskIds, refineError := p.getTheMarathonAppInfo(appId)
		if refineError != nil {
			err = refineError
			logrus.Errorf("get marathon task info error is %+v", err)
			return
		}
		newInstanceIds := []string{}
		for _, theId := range app.InstanceIds {
			if checkFlag := p.checkInstanceLive(theId, mesosTaskIds); checkFlag {
				newInstanceIds = append(newInstanceIds, theId)
			} else {
				//reset ip to the newid to support stable ip address.
				refineError = GetIPResourceService().resetIp(theId, newId)
				if refineError != nil {
					logrus.Errorf("release IP error is %+v", refineError)
				}
				//update instance to failed
				refineError = p.updateAciStatus(theId, ACI_STATUS_FAILED)
				if refineError != nil {
					logrus.Errorf("update aci [%v] status to [%v] is failed, error is %v",
						theId, ACI_STATUS_FAILED, refineError)
				}
			}
		}
		app.InstanceIds = append(newInstanceIds, newId)
	}
	return nil
}

func (p *AciService) getTheMarathonAppInfo(appId string) (mesosTaskIds []string,
	err error) {
	mesosTaskIds = []string{}
	//Do marathon get
	marathonEndpoint, err := common.UTIL.ZkClient.GetMarathonEndpoint()
	if err != nil {
		logrus.Errorf("get marathon endpoint err is %v", err)
		return
	}
	url := strings.Join([]string{"http://", marathonEndpoint, "/v2/apps", appId}, "")
	logrus.Debugf("marathonEndpoint, url=%v", marathonEndpoint)
	resp, err := httpclient.Http_get(url, "", httpclient.Header{"Content-Type", "application/json"})
	defer resp.Body.Close()
	if err != nil {
		logrus.Errorf("http get marathon tasks failed, error is %v", err)
		return
	}
	// check response status, and load body
	data, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err = errors.New("http get marathon tasks return 400 error")
		logrus.Errorf("http get marathon tasks return 400 error, message is %v", string(data))
		return
	}

	// parse response data,
	// get the strings slides of existing mesos task id and return
	jsondata := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.Decode(&jsondata)
	jq := jsonq.NewQuery(jsondata)
	instances, err := jq.Int("app", "instances")
	if err != nil {
		logrus.Errorf("get app instances from marathon response failed, error is %v", err)
		return
	}

	for i := 0; i < instances; i++ {
		id, _ := jq.String("app", "tasks", strconv.Itoa(i), "id")
		mesosTaskIds = append(mesosTaskIds, id)
	}

	return
}

func (p *AciService) checkInstanceLive(instanceId string, mesosTaskIds []string) bool {
	// first, get aci by instanceId
	aci, _, err := p.queryById(instanceId)
	if err != nil {
		logrus.Errorf("check instance live failed caused by query aci by id [%v], error is %v",
			instanceId, err)
		// FIXME: here return true means this instance is alived, need to fix it.
		return true
	}

	for _, id := range mesosTaskIds {
		if id == aci.MesosTaskId {
			return true
		}
	}

	return false
}

func (p *AciService) updateAciStatus(objectId, status string) (err error) {
	statusValidate := false
	status = strings.ToUpper(status)
	for _, stat := range ACI_STATUS {
		if stat == status {
			statusValidate = true
			break
		}
	}
	if statusValidate {
		// get appInstance by objectId
		aci, _, err := p.queryById(objectId)
		if err != nil {
			return err
		}
		aci.LifeCycleStatus = status

		// save to db
		logrus.Infof("update app instance %s status to %s", objectId, status)
		_, _, err = p.updateById(objectId, &aci)
		if err != nil {
			return err
		}
		return nil
	} else {
		err = errors.New("Invalided status.")
		return err
	}
	return nil
}

func (p *AciService) removeAciFromSgi(sgi *entity.ServiceGroupInstance, groupId, appPath, aciId string) (flag bool, err error) {
	app, err := entity.FindAppInSgi(sgi, appPath)
	if err != nil {
		logrus.Errorf("can not find the app by path [%v] in the sgi [%v], err is %v", appPath, sgi.ObjectId, err)
		flag = false
		return
	}

	for k, instanceId := range app.InstanceIds {
		if instanceId == aciId {
			app.InstanceIds = append(app.InstanceIds[:k], app.InstanceIds[k+1:]...)
		}
	}
	flag = true
	return

}
