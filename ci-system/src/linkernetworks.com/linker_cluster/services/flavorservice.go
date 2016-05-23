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
	flavorService *FlavorService = nil
	onceFlavor    sync.Once

	FLAVOR_ERROR_QUERY  string = "E40201"
	FLAVOR_ERROR_CREATE string = "E40202"
)

type FlavorService struct {
	collectionName string
}

func GetFlavorService() *FlavorService {
	onceFlavor.Do(func() {
		logrus.Debugf("Once called from flavorService ......................................")
		flavorService = &FlavorService{"flavor"}
	})
	return flavorService

}

func (p *FlavorService) Create(flavor entity.Flavor, x_auth_token string) (newFlavor entity.Flavor,
	errorCode string, err error) {
	logrus.Infof("start to create flavor [%v]", flavor)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_flavor", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create flavor [%v] error is %v", flavor, err)
		return
	}

	// generate ObjectId
	if !bson.IsObjectIdHex(flavor.ObjectId.Hex()) {
		flavor.ObjectId = bson.NewObjectId()
	}

	flavor.TimeCreate = dao.GetCurrentTime()
	flavor.TimeUpdate = flavor.TimeCreate
	err = dao.HandleInsert(p.collectionName, flavor)
	if err != nil {
		errorCode = FLAVOR_ERROR_CREATE
		logrus.Errorf("insert flavor [%v] to db error is %v", flavor, err)
		return
	}
	newFlavor = flavor

	return

}

func (p *FlavorService) QueryAllByName(supplier string, skip int,
	limit int, x_auth_token string) (total int, flavors []entity.Flavor,
	errorCode string, err error) {
	logrus.Infof("start to get flavor by supplier [%v]", supplier)
	if strings.TrimSpace(supplier) == "" {
		return p.QueryAll(skip, limit, x_auth_token)
	}

	query := bson.M{}
	query["provider_type"] = supplier
	return p.queryByQuery(query, skip, limit, x_auth_token, false)
}

func (p *FlavorService) QueryAll(skip int, limit int, x_auth_token string) (total int,
	flavors []entity.Flavor, errorCode string, err error) {
	return p.queryByQuery(bson.M{}, skip, limit, x_auth_token, false)
}

func (p *FlavorService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int, flavors []entity.Flavor,
	errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_flavor", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	flavors = []entity.Flavor{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, "...."}
	total, err = dao.HandleQueryAll(&flavors, queryStruct)
	if err != nil {
		logrus.Errorf("query hosts by query [%v] error is %v", query, err)
		errorCode = FLAVOR_ERROR_QUERY

	}
	return
}

func (p *FlavorService) UpdateById(objectId string, flavor entity.Flavor, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update flavor [%v]", flavor)

	// do authorize first
	if authorized := GetAuthService().Authorize("update_flavor", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update flavor with objectId [%v] error is %v", objectId, err)
		return
	}

	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}
	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	flavor.ObjectId = bson.ObjectIdHex(objectId)
	flavor.TimeUpdate = dao.GetCurrentTime()

	created, err = dao.HandleUpdateOne(&flavor, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update flavor [%v] error is %v", flavor, err)
		errorCode = HOST_ERROR_UPDATE
	}
	return

}
