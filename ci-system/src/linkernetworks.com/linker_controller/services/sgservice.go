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
	sgService             *SgService = nil
	onceSg                sync.Once
	SG_STATUS_UNPUBLISHED = "unpublished"
	SG_STATUS_VERIFYING   = "verifying"
	SG_STATUS_PUBLISHED   = "published"
	SG_STATUS             = []string{
		SG_STATUS_PUBLISHED,
		SG_STATUS_UNPUBLISHED,
		SG_STATUS_VERIFYING,
	}

	SG_ERROR_CREATE    string = "E11041"
	SG_ERROR_UPDATE    string = "E11043"
	SG_ERROR_DELETE    string = "E11042"
	SG_ERROR_PUBLISH   string = "E11047"
	SG_ERROR_UNPUBLISH string = "E11048"
	SG_ERROR_SUBMIT    string = "E11044"
	SG_ERROR_UNIQUE    string = "E11045"
	SG_ERROR_QUERY     string = "E11046"
)

type SgService struct {
	collectionName string
}

func GetSgService() *SgService {
	onceSg.Do(func() {
		logrus.Debugf("Once called from sgService ......................................")
		sgService = &SgService{"service_group"}
	})
	return sgService
}

func (p *SgService) Create(sg entity.ServiceGroup, x_auth_token string) (newSg entity.ServiceGroup,
	errorCode string, err error) {
	logrus.Infof("start to create sg [%v]", sg)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_sg", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create sg [%v] error is %v", sg, err)
		return
	}
	// validate name, should be unique
	_, sg_exists, _, err := p.QueryAllByName(sg.Id, 0, 1, x_auth_token)
	if err == nil && len(sg_exists) > 0 {
		err = errors.New("the name of service group must be unique!")
		errorCode = SG_ERROR_UNIQUE
		logrus.Errorf("create sg [%v] error is %v", sg, err)
		return
	}
	// generate ObjectId
	sg.ObjectId = bson.NewObjectId()

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		logrus.Errorf("get token failed when create sg [%v], error is %v", sg, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	// set token_id and user_id from token
	sg.Tenant_id = token.Tenant.Id
	sg.User_id = token.User.Id

	// set created_time and updated_time
	sg.State = SG_STATUS_UNPUBLISHED
	sg.TimeCreate = dao.GetCurrentTime()
	sg.TimeUpdate = sg.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, sg)
	if err != nil {
		logrus.Errorf("insert sg [%v] to db error is %v", sg, err)
		errorCode = SG_ERROR_CREATE
		return
	}

	newSg = sg
	return
}

func (p *SgService) UpdateById(objectId string, sg entity.ServiceGroup, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update sg [%v]", sg)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_sg", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update sg with objectId [%v] error is %v", objectId, err)
		return
	}
	// validate sg
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	sg.ObjectId = bson.ObjectIdHex(objectId)
	sg.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&sg, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update sg [%v] error is %v", sg, err)
		errorCode = SG_ERROR_UPDATE
	}
	return
}

func (p *SgService) UpdateStateById(objectId string, newState string, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update sg by objectId [%v] status to %v", objectId, newState)
	// do authorize first
	operation := ""
	switch newState {
	case SG_STATUS_PUBLISHED:
		operation = "publish_sg"
		errorCode = SG_ERROR_PUBLISH
	case SG_STATUS_UNPUBLISHED:
		operation = "publish_sg"
		errorCode = SG_ERROR_UNPUBLISH
	case SG_STATUS_VERIFYING:
		operation = "submit_sg"
		errorCode = SG_ERROR_SUBMIT
	default:
		operation = "update_sg"
	}
	if authorized := GetAuthService().Authorize(operation, x_auth_token, objectId, p.collectionName); !authorized {
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

	// get sg by objectId
	sg, _, err := p.QueryById(objectId, x_auth_token)
	if err != nil {
		logrus.Errorf("get sg by objeceId [%v] failed, error is %v", objectId, err)
		return
	}

	if sg.State == newState {
		logrus.Infof("this sg [%v] is already in state [%v]", sg, newState)
		return false, "", nil
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	sg.ObjectId = bson.ObjectIdHex(objectId)
	sg.State = newState
	sg.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&sg, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update sg with objectId [%v] status to [%v] failed, error is %v", objectId, newState, err)
	}
	return
}

func (p *SgService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete sg with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_sg", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete sg with objectId [%v] error is %v", objectId, err)
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
		logrus.Errorf("delete sg [objectId=%v] error is %v", objectId, err)
		errorCode = SG_ERROR_DELETE
	}
	return
}

func (p *SgService) DeleteBySgId(sgId, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete sgs with [sgId=%v]", sgId)
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("delete_sgs", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	query := bson.M{}
	if strings.TrimSpace(sgId) != "" {
		query["id"] = sgId
	}

	selector := generateQueryWithAuth(query, authQuery)

	err = dao.HandleDelete(p.collectionName, false, selector)
	if err != nil {
		logrus.Errorf("delete sgs by query [id=%v] error is %v", sgId, err)
		errorCode = SG_ERROR_DELETE
	}

	return
}

func (p *SgService) QueryById(objectId string, x_auth_token string) (sg entity.ServiceGroup,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// do authorize first
	if authorized := GetAuthService().Authorize("get_sg", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get sg with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	sg = entity.ServiceGroup{}
	err = dao.HandleQueryOne(&sg, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query sg [objectId=%v] error is %v", objectId, err)
		errorCode = SG_ERROR_QUERY
	}
	return
}

func (p *SgService) GetOperationById(objectId string, x_auth_token string) (operations map[string]int,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	operationList := []string{"update_sg", "delete_sg", "publish_sg", "submit_sg"}
	operations, err = GetAuthService().AuthOperation(operationList, x_auth_token,
		objectId, p.collectionName)
	if err != nil {
		logrus.Errorf("get auth operation of [objectId=%v] error is %v", objectId, err)
		errorCode = COMMON_ERROR_INTERNAL
	}
	return
}

func (p *SgService) QueryAll(skip int, limit int, x_auth_token string) (total int,
	sgs []entity.ServiceGroup, errorCode string, err error) {
	return p.queryByQuery(bson.M{}, skip, limit, x_auth_token, false)
}

func (p *SgService) QueryAllByName(sg_name string, skip int,
	limit int, x_auth_token string) (total int, sgs []entity.ServiceGroup,
	errorCode string, err error) {
	// if sg_name is empty, query all
	if strings.TrimSpace(sg_name) == "" {
		return p.QueryAll(skip, limit, x_auth_token)
	}

	query := bson.M{}
	query["id"] = sg_name
	return p.queryByQuery(query, skip, limit, x_auth_token, false)
}

func (p *SgService) QueryAllPublishedByName(sg_name string, skip int, limit int,
	x_auth_token string) (total int, sgs []entity.ServiceGroup,
	errorCode string, err error) {
	query := bson.M{}
	query["state"] = SG_STATUS_PUBLISHED
	// if sg_name is not empty, add query condition
	if strings.TrimSpace(sg_name) != "" {
		query["id"] = sg_name
	}
	// here, last parameter means skipAuth=true.
	return p.queryByQuery(query, skip, limit, x_auth_token, true)
}

func (p *SgService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int, sgs []entity.ServiceGroup,
	errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth service first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_sgs", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	sgs = []entity.ServiceGroup{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, ""}
	total, err = dao.HandleQueryAll(&sgs, queryStruct)
	if err != nil {
		logrus.Errorf("query sgs by query [%v] error is %v", query, err)
		errorCode = SG_ERROR_QUERY
	}
	return
}
