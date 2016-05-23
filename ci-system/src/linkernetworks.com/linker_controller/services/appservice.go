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
	appService       *AppService = nil
	onceApp          sync.Once
	APP_ERROR_CREATE string = "E11021"
	APP_ERROR_UPDATE string = "E11023"
	APP_ERROR_DELETE string = "E11022"
	APP_ERROR_QUERY  string = "E11024"
)

type AppService struct {
	collectionName string
}

func GetAppService() *AppService {
	onceApp.Do(func() {
		logrus.Debugf("Once called from appService ......................................")
		appService = &AppService{"app"}
	})
	return appService
}

func (p *AppService) Create(app entity.App, x_auth_token string) (newApp entity.App,
	errorCode string, err error) {
	logrus.Infof("start to create app [%v]", app)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_app", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create app [%v] error is %v", app, err)
		return
	}
	// generate ObjectId
	app.ObjectId = bson.NewObjectId()

	token, err := GetTokenById(x_auth_token)
	if err != nil {
		errorCode = APP_ERROR_CREATE
		logrus.Errorf("get token failed when create app [%v], error is %v", app, err)
		return
	}

	// set token_id and user_id from token
	app.Tenant_id = token.Tenant.Id
	app.User_id = token.User.Id

	// set created_time and updated_time
	app.TimeCreate = dao.GetCurrentTime()
	app.TimeUpdate = app.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, app)
	if err != nil {
		errorCode = APP_ERROR_CREATE
		logrus.Errorf("create app [%v] to bson error is %v", app, err)
		return
	}

	newApp = app
	return
}

func (p *AppService) UpdateById(objectId string, app entity.App, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update app [%v]", app)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_app", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update app with objectId [%v] error is %v", objectId, err)
		return
	}
	// validate app
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	app.ObjectId = bson.ObjectIdHex(objectId)
	app.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&app, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update app [%v] error is %v", app, err)
		errorCode = APP_ERROR_UPDATE
	}
	return
}

func (p *AppService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete app with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_app", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete app with objectId [%v] error is %v", objectId, err)
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
		logrus.Errorf("delete app [objectId=%v] error is %v", objectId, err)
		errorCode = APP_ERROR_DELETE
	}
	return
}

func (p *AppService) QueryAll(skip int, limit int, x_auth_token string) (total int, apps []entity.App,
	errorCode string, err error) {
	// get auth query from auth service first
	authQuery, err := GetAuthService().BuildQueryByAuth("list_apps", x_auth_token)
	if err != nil {
		logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
		errorCode = COMMON_ERROR_INTERNAL
		return
	}

	logrus.Debugf("auth query is %v", authQuery)

	selector := generateQueryWithAuth(bson.M{}, authQuery)
	logrus.Debugf("selector is %v", selector)
	sort := ""
	apps = []entity.App{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, sort}
	total, err = dao.HandleQueryAll(&apps, queryStruct)
	if err != nil {
		logrus.Errorf("list app [token=%v] failed, error is %v", x_auth_token, err)
		errorCode = APP_ERROR_QUERY
	}
	return
}

func (p *AppService) QueryById(objectId string, x_auth_token string) (app entity.App,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	// do authorize first
	if authorized := GetAuthService().Authorize("get_app", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get app with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	app = entity.App{}
	err = dao.HandleQueryOne(&app, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query app [objectId=%v] error is %v", objectId, err)
		errorCode = APP_ERROR_QUERY
	}
	return
}

func (p *AppService) GetOperationById(objectId string, x_auth_token string) (operations map[string]int,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	operationList := []string{"update_app", "delete_app"}
	operations, err = GetAuthService().AuthOperation(operationList, x_auth_token, objectId, p.collectionName)
	if err != nil {
		logrus.Errorf("get auth operation of [objectId=%v] error is %v", objectId, err)
		errorCode = COMMON_ERROR_INTERNAL
	}
	return
}
