package services

import (
	"errors"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
)

var (
	acpService       *AcpService = nil
	onceAcp          sync.Once
	ACP_ERROR_CREATE string = "E11011"
	ACP_ERROR_UPDATE string = "E11013"
	ACP_ERROR_DELETE string = "E11012"
	ACP_ERROR_QUERY  string = "E11014"
)

type AcpService struct {
	collectionName string
}

func GetAcpService() *AcpService {
	onceAcp.Do(func() {
		logrus.Debugf("Once called from acpService ......................................")
		acpService = &AcpService{"app_container_package"}
	})
	return acpService
}

func (p *AcpService) Create(acp entity.AppContainerPackage, x_auth_token string) (newApp entity.AppContainerPackage,
	errorCode string, err error) {
	logrus.Infof("start to create acp [%v]", acp)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_apppackage", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create acp [%v] error is %v", acp, err)
		return
	}
	// generate ObjectId
	acp.ObjectId = bson.NewObjectId()

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = ACP_ERROR_CREATE
		logrus.Errorf("get token failed when create acp [%v], error is %v", acp, err)
		return
	}

	// set token_id and user_id from token
	acp.Tenant_id = token.Tenant.Id
	acp.User_id = token.User.Id

	// set created_time and updated_time
	acp.TimeCreate = dao.GetCurrentTime()
	acp.TimeUpdate = acp.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, acp)
	if err != nil {
		errorCode = ACP_ERROR_CREATE
		logrus.Errorf("create acp [%v] to bson error is %v", acp, err)
		return
	}

	newApp = acp
	return
}

func (p *AcpService) UpdateById(objectId string, acp entity.AppContainerPackage, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update acp [%v]", acp)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_apppackage", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update acp with objectId [%v] error is %v", objectId, err)
		return
	}
	// validate acp
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	acp.ObjectId = bson.ObjectIdHex(objectId)
	acp.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&acp, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update acp [%v] error is %v", acp, err)
		errorCode = ACP_ERROR_UPDATE
	}
	return
}

func (p *AcpService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete acp with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_apppackage", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete acp with objectId [%v] error is %v", objectId, err)
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
		logrus.Errorf("delete acp [objectId=%v] error is %v", objectId, err)
		errorCode = ACP_ERROR_DELETE
	}
	return
}

func (p *AcpService) DeleteBySgOrApp(sgId, appId, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete acps with [sgId=%v, appId=%v]", sgId, appId)
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("delete_apppackages", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	var oriQuery = bson.M{}
	if strings.TrimSpace(sgId) != "" {
		oriQuery["service_group_id"] = sgId
	}

	if strings.TrimSpace(appId) != "" {
		oriQuery["app_container_id"] = appId
	}

	selector := generateQueryWithAuth(oriQuery, authQuery)

	err = dao.HandleDelete(p.collectionName, false, selector)
	if err != nil {
		logrus.Errorf("delete acps by query [service_group_id=%v, app_container_id=%v] error is %v",
			sgId, appId, err)
		errorCode = ACP_ERROR_DELETE
	}
	return
}

func (p *AcpService) QueryAll(skip int, limit int, x_auth_token string) (total int, acps []entity.AppContainerPackage,
	errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_apppackages", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	logrus.Debugf("auth query is %v", authQuery)

	selector := generateQueryWithAuth(bson.M{}, authQuery)
	logrus.Debugf("selector is %v", selector)
	sort := ""
	acps = []entity.AppContainerPackage{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, sort}
	total, err = dao.HandleQueryAll(&acps, queryStruct)
	if err != nil {
		logrus.Errorf("list acp [token=%v] failed, error is %v", x_auth_token, err)
		errorCode = ACP_ERROR_QUERY
	}
	return
}

func (p *AcpService) QueryAllByName(service_group_id string, app_container_id string,
	skip int, limit int, x_auth_token string) (total int, acps []entity.AppContainerPackage,
	errorCode string, err error) {
	// get auth query from auth service first
	query := bson.M{}
	if strings.TrimSpace(service_group_id) != "" {
		query["service_group_id"] = service_group_id
	}

	if strings.TrimSpace(app_container_id) != "" {
		query["app_container_id"] = app_container_id
	}

	return p.queryByQuery(query, skip, limit, x_auth_token, false)
}

func (p *AcpService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int,
	acps []entity.AppContainerPackage, errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth service first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_apppackages", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	acps = []entity.AppContainerPackage{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, ""}
	total, err = dao.HandleQueryAll(&acps, queryStruct)
	if err != nil {
		logrus.Errorf("query acp by query [%v] error is %v", query, err)
		errorCode = SG_ERROR_QUERY
	}
	return
}

func (p *AcpService) QueryById(objectId string, x_auth_token string) (acp entity.AppContainerPackage,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// do authorize first
	if authorized := GetAuthService().Authorize("get_apppackage", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get acp with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	acp = entity.AppContainerPackage{}
	err = dao.HandleQueryOne(&acp, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query acp [objectId=%v] error is %v", objectId, err)
		errorCode = ACP_ERROR_QUERY
	}
	return
}

func (p *AcpService) GetOperationById(objectId string, x_auth_token string) (operations map[string]int,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	operationList := []string{"update_acp", "delete_acp"}
	operations, err = GetAuthService().AuthOperation(operationList, x_auth_token, objectId, p.collectionName)
	if err != nil {
		logrus.Errorf("get auth operation of [objectId=%v] error is %v", objectId, err)
		errorCode = COMMON_ERROR_INTERNAL
	}
	return
}

func (p *AcpService) QueryByAppPath(appPath, x_auth_token string) (total int,
	acps []entity.AppContainerPackage, errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_apppackages", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	var oriQuery = bson.M{}
	oriQuery["app_container_id"] = appPath
	selector := generateQueryWithAuth(oriQuery, authQuery)
	logrus.Debugf("selector is %v", selector)
	acps = []entity.AppContainerPackage{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, 0, 1, ""}
	total, err = dao.HandleQueryAll(&acps, queryStruct)
	if err != nil {
		logrus.Errorf("get acp by appPath [%v] failed, error is %v", appPath, err)
		errorCode = ACP_ERROR_QUERY
	}
	return
}
