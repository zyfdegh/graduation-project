package services

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/compose/mejson"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_controller/common"
	"strconv"
	"sync"
)

var (
	quotaService *QuotaService = nil
	onceQuota    sync.Once

	QUOTA_ERROR_ORDER    string = "E11091"
	QUOTA_ERROR_INSTANCE string = "E11092"
)

type QuotaService struct {
	// collectionName string
}

func GetQuotaService() *QuotaService {
	onceQuota.Do(func() {
		logrus.Debugf("Once called from quotaService ......................................")
		quotaService = &QuotaService{}
	})
	return quotaService
}

func (p *QuotaService) CheckOrder(sgo *entity.ServiceGroupOrder) (string, error) {
	logrus.Infoln("check order and app quota")

	orderlimit := common.UTIL.Props.GetInt("order_limit", 1)
	applimit := common.UTIL.Props.GetInt("app_limit", 20)

	//1. check order limitation
	//query sgo status
	statusList := []string{SGO_STATUS_DEPLOYED, SGO_STATUS_MODIFYING,
		SGO_STATUS_TERMINATING, SGO_STATUS_DEPLOYING}
	inMap := make(map[string]interface{})
	inMap["$in"] = statusList

	queryMap := make(map[string]interface{})
	queryMap["life_cycle_status"] = inMap
	queryMap["user_id"] = sgo.User_id

	// field := bson.M{}
	// field["order_id"] = 1
	// field["service_group_instance_id"] = 1

	querybson, err := mejson.Unmarshal(queryMap)
	if err != nil {
		logrus.Errorln("format querymap error %v", err)
		return COMMON_ERROR_INTERNAL, err
	}

	queryStruct := dao.QueryStruct{
		CollectionName: GetSgoService().collectionName,
		Selector:       querybson,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	documents := []entity.ServiceGroupOrder{}
	count, err := dao.HandleQueryAll(&documents, queryStruct)

	// count, _, jsondoc, err := p.Dao.HandleQuery(sgoCollection, queryMap,
	// 	false, field, 0, 0, "", "true")

	if err != nil {
		logrus.Errorln("query sgo by status and user error %v", err)
		return "", nil
	}

	if count+1 > orderlimit {
		errMsg := "Required order have exceed maximum order limitation:" + strconv.Itoa(orderlimit)
		logrus.Warnln(errMsg)
		return QUOTA_ERROR_ORDER, errors.New(errMsg)
	}

	//2. check app limitation
	sgInstanceIdList := []string{}

	// arrays := jsondoc.([]interface{})
	for i := 0; i < len(documents); i++ {
		// doc := arrays[i].(map[string]interface{})
		doc := documents[i]
		sgInstanceIdList = append(sgInstanceIdList, doc.ServiceGroupInstanceId)
	}
	if len(sgInstanceIdList) > 0 {
		instanceInMap := make(map[string]interface{})
		instanceInMap["$in"] = sgInstanceIdList

		appStatusList := []string{ACI_STATUS_CREATED, ACI_STATUS_CONFIGED,
			ACI_STATUS_UNCONFIGED}
		statusInMap := make(map[string]interface{})
		statusInMap["$in"] = appStatusList

		queryMap = make(map[string]interface{})
		queryMap["lifecycle_status"] = statusInMap
		queryMap["service_group_instance_id"] = instanceInMap

		querybson, err := mejson.Unmarshal(queryMap)
		if err != nil {
			logrus.Errorln("format querymap for app instance error %v", err)
			return "", nil
		}

		queryStruct := dao.QueryStruct{
			CollectionName: GetAciService().collectionName,
			Selector:       querybson,
			Skip:           0,
			Limit:          0,
			Sort:           ""}

		acis := []entity.AppContainerInstance{}
		count, err = dao.HandleQueryAll(&acis, queryStruct)
		// count, _, _, err = p.Dao.HandleQuery(aciCollection, queryMap,
		// 	false, bson.M{}, 0, 0, "", "false")
		if err != nil {
			logrus.Errorln("query app instances by id and status error %v ", err)
			return "", nil
		}
	} else {
		count = 0
	}

	sg, err := p.getSgBySgo(sgo)
	if err != nil {
		logrus.Errorln("get sg by sgo error %v ", err)
		return "", nil
	}

	num := getAppNumInSg(sg)
	if num+count > applimit {
		errMsg := "Required apps have execced maximum app limitation:" + strconv.Itoa(applimit)
		logrus.Warnln(errMsg)
		return QUOTA_ERROR_INSTANCE, errors.New(errMsg)
	}

	return "", nil

}

// CheckScale parses the ServiceGroupOrder and check the scale
// If successful, the error is nil.
func (p *QuotaService) CheckScale(sgo *entity.ServiceGroupOrder, num int) (string, error) {
	logrus.Infoln("check scale operation quota")
	applimit := common.UTIL.Props.GetInt("app_limit", 20)
	sgi, err := getSgiBysgiId(sgo.ServiceGroupInstanceId)
	if err != nil {
		logrus.Warnln("get sgi by id error %v ", err)
		return "", nil
	}

	currentNum := getAppNumInSgi(sgi)
	if currentNum == num {
		logrus.Errorln("required app number can not equal with current number!")
		return COMMON_ERROR_INVALIDATE, errors.New("Invalid parameter! Required app instances can not equals with current instances ")
	}

	appCount := getAppInsCountByUser(sgo.User_id)
	desireNum := appCount - currentNum + num

	if desireNum > applimit {
		errMsg := "Required apps have execced maximum app limitation:" + strconv.Itoa(applimit)
		logrus.Errorln(errMsg)
		return QUOTA_ERROR_INSTANCE, errors.New(errMsg)
	}

	return "", nil

}

func getAppInsCountByUser(userid string) int {
	statusList := []string{SGO_STATUS_DEPLOYED, SGO_STATUS_MODIFYING,
		SGO_STATUS_TERMINATING, SGO_STATUS_DEPLOYING}
	inMap := make(map[string]interface{})
	inMap["$in"] = statusList

	queryMap := make(map[string]interface{})
	queryMap["life_cycle_status"] = inMap
	queryMap["user_id"] = userid

	// field := bson.M{}
	// field["order_id"] = 1
	// field["service_group_instance_id"] = 1
	querybson, err := mejson.Unmarshal(queryMap)
	if err != nil {
		logrus.Errorln("format querymap for service group order error %v", err)
		return 0
	}

	queryStruct := dao.QueryStruct{
		CollectionName: GetSgoService().collectionName,
		Selector:       querybson,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	sgos := []entity.ServiceGroupOrder{}
	_, err = dao.HandleQueryAll(&sgos, queryStruct)
	// _, _, jsondoc, err := p.Dao.HandleQuery(sgoCollection, queryMap,
	// 	false, field, 0, 0, "", "true")
	if err != nil {
		logrus.Errorln("query sgo by status and user error %v", err)
		return 0
	}

	//get app instance by sgi id and status
	sgInstanceIdList := []string{}
	// arrays := jsondoc.([]interface{})
	for i := 0; i < len(sgos); i++ {
		// doc := arrays[i].(map[string]interface{})
		doc := sgos[i]
		sgInstanceIdList = append(sgInstanceIdList, doc.ServiceGroupInstanceId)
	}
	if len(sgInstanceIdList) > 0 {
		instanceInMap := make(map[string]interface{})
		instanceInMap["$in"] = sgInstanceIdList

		appStatusList := []string{ACI_STATUS_CREATED, ACI_STATUS_CONFIGED,
			ACI_STATUS_UNCONFIGED}
		statusInMap := make(map[string]interface{})
		statusInMap["$in"] = appStatusList

		queryMap = make(map[string]interface{})
		queryMap["lifecycle_status"] = statusInMap
		queryMap["service_group_instance_id"] = instanceInMap

		querybson, err := mejson.Unmarshal(queryMap)
		if err != nil {
			logrus.Errorln("format querymap for app instance error %v", err)
			return 0
		}

		queryStruct := dao.QueryStruct{
			CollectionName: GetAciService().collectionName,
			Selector:       querybson,
			Skip:           0,
			Limit:          0,
			Sort:           ""}
		acis := []entity.AppContainerInstance{}
		count, err := dao.HandleQueryAll(&acis, queryStruct)
		// count, _, _, err := p.Dao.HandleQuery(aciCollection, queryMap,
		// 	false, bson.M{}, 0, 0, "", "false")

		if err != nil {
			logrus.Errorln("query app instances by id and status error %v ", err)
			return 0
		}
		return count
	} else {
		return 0
	}

}

func (p *QuotaService) getSgBySgo(sgo *entity.ServiceGroupOrder) (sg *entity.ServiceGroup,
	err error) {

	if sgo.ServiceGroupObjId == "" {
		sg, err = getSgByGroupId(sgo.ServiceGroupId)
		if err != nil {
			logrus.Warnln("get service group by group id error %v ", err)
			return
		}

	} else {
		sg, err = getSgById(sgo.ServiceGroupObjId)
		if err != nil {
			logrus.Warnln("get service group by obj id error %v ", err)
			return
		}
	}

	return
}

func getSgiBysgiId(serviceGroupInstanceId string) (serviceGroupInstance *entity.ServiceGroupInstance,
	err error) {

	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(serviceGroupInstanceId)

	queryStruct := dao.QueryStruct{
		CollectionName: GetSgiService().collectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	serviceGroupInstance = new(entity.ServiceGroupInstance)
	err = dao.HandleQueryOne(serviceGroupInstance, queryStruct)
	// _, _, sgiDocument, err := p.Dao.HandleQuery(sgiCollection, selector,
	// 	true, bson.M{}, 0, 1, "", "true")

	if err != nil {
		return
	}
	return
}

func getSgByGroupId(serviceGroupId string) (serviceGroup *entity.ServiceGroup,
	err error) {
	selector := bson.M{}
	selector["id"] = serviceGroupId

	queryStruct := dao.QueryStruct{
		CollectionName: GetSgService().collectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	serviceGroup = new(entity.ServiceGroup)
	err = dao.HandleQueryOne(serviceGroup, queryStruct)
	// _, _, sgDocument, err := p.Dao.HandleQuery(sgCollection, selector,
	// 	true, bson.M{}, 0, 1, "", "true")

	if err != nil {
		return
	}
	return
}

func getSgById(id string) (serviceGroup *entity.ServiceGroup, err error) {
	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(id)

	queryStruct := dao.QueryStruct{
		CollectionName: GetSgService().collectionName,
		Selector:       selector,
		Skip:           0,
		Limit:          0,
		Sort:           ""}

	serviceGroup = new(entity.ServiceGroup)
	err = dao.HandleQueryOne(serviceGroup, queryStruct)
	// _, _, sgDocument, err := p.Dao.HandleQuery(sgCollection, selector,
	// 	true, bson.M{}, 0, 1, "", "true")

	if err != nil {
		return
	}

	return
}
