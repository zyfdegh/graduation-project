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
	notificationService        *NotificationService = nil
	onceNotification           sync.Once
	NOTIFICATICON_ERROR_CREATE string = "E11111"
	NOTIFICATION_ERROR_UPDATE  string = "E11112"
	NOTIFICATION_ERROR_DELETE  string = "E11113"
	NOTIFICATION_ERROR_QUERY   string = "E11114"
	NOTIFICATION_ERROR_UNIQUE  string = "E11115"
)

type NotificationService struct {
	collectionName string
}

func GetNotificationService() *NotificationService {
	onceNotification.Do(func() {
		logrus.Debugf("Once called from notificationService ......................................")
		notificationService = &NotificationService{"notification"}
	})
	return notificationService
}

func (p *NotificationService) Create(notification entity.Notification, x_auth_token string) (newNotification entity.Notification,
	errorCode string, err error) {
	logrus.Infof("start to create notification [%v]", notification)
	// do authorize first
	if authorized := GetAuthService().Authorize("create_notification", x_auth_token, "", p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("create notification [%v] error is %v", notification, err)
		return
	}

	// validate name, should be unique
	_, notification_exists, _, err := p.QueryAllByName(notification.ServiceGroupId, 0, 1, x_auth_token)
	if err == nil && len(notification_exists) > 0 {
		err = errors.New("the name of service group must be unique!")
		errorCode = NOTIFICATION_ERROR_UNIQUE
		logrus.Errorf("create notification [%v] error is %v", notification, err)
		return
	}

	// generate ObjectId
	notification.ObjectId = bson.NewObjectId()
	notification.TimeCreate = dao.GetCurrentTime()
	notification.TimeUpdate = notification.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, notification)
	if err != nil {
		logrus.Errorf("insert notification [%v] to db error is %v", notification, err)
		errorCode = NOTIFICATICON_ERROR_CREATE
		return
	}

	newNotification = notification
	return

}

func (p *NotificationService) UpdateById(objectId string, notification entity.Notification, x_auth_token string) (created bool,
	errorCode string, err error) {
	logrus.Infof("start to update notification [%v]", notification)
	// do authorize first
	if authorized := GetAuthService().Authorize("update_notification", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("update notification with objectId [%v] error is %v", objectId, err)
		return
	}
	// validate notification
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)

	// reset objectId and updated_time
	notification.ObjectId = bson.ObjectIdHex(objectId)
	notification.TimeUpdate = dao.GetCurrentTime()

	// insert bson to mongodb
	created, err = dao.HandleUpdateOne(&notification, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("update notification [%v] error is %v", notification, err)
		errorCode = NOTIFICATION_ERROR_UPDATE
	}
	return

}

func (p *NotificationService) QueryById(objectId string, x_auth_token string) (notification entity.Notification,
	errorCode string, err error) {
	if !bson.IsObjectIdHex(objectId) {
		err = errors.New("invalide ObjectId.")
		errorCode = COMMON_ERROR_INVALIDATE
		return
	}

	// do authorize first
	if authorized := GetAuthService().Authorize("get_notification", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("get notification with objectId [%v] error is %v", objectId, err)
		return
	}

	var selector = bson.M{}
	selector["_id"] = bson.ObjectIdHex(objectId)
	//	notification := entity.Notification{}
	err = dao.HandleQueryOne(&notification, dao.QueryStruct{p.collectionName, selector, 0, 0, ""})
	if err != nil {
		logrus.Errorf("query notification [objectId=%v] error is %v", objectId, err)
		errorCode = NOTIFICATION_ERROR_QUERY
	}
	return
}

func (p *NotificationService) DeleteById(objectId string, x_auth_token string) (errorCode string, err error) {
	logrus.Infof("start to delete notification with objectId [%v]", objectId)
	// do authorize first
	if authorized := GetAuthService().Authorize("delete_notification", x_auth_token, objectId, p.collectionName); !authorized {
		err = errors.New("required opertion is not authorized!")
		errorCode = COMMON_ERROR_UNAUTHORIZED
		logrus.Errorf("delete notification with objectId [%v] error is %v", objectId, err)
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
		logrus.Errorf("delete notification [objectId=%v] error is %v", objectId, err)
		errorCode = NOTIFICATION_ERROR_DELETE
	}

	return
}

func (p *NotificationService) QueryAllByName(notification_name string, skip int,
	limit int, x_auth_token string) (total int, notifications []entity.Notification,
	errorCode string, err error) {
	// if notification_name is empty, query all
	if strings.TrimSpace(notification_name) == "" {
		return p.QueryAll(skip, limit, x_auth_token)
	}

	query := bson.M{}
	query["service_group_id"] = notification_name
	return p.queryByQuery(query, skip, limit, x_auth_token, false)

}

func (p *NotificationService) QueryAll(skip int, limit int, x_auth_token string) (total int,
	notifications []entity.Notification, errorCode string, err error) {
	return p.queryByQuery(bson.M{}, skip, limit, x_auth_token, false)
}

func (p *NotificationService) queryByQuery(query bson.M, skip int, limit int,
	x_auth_token string, skipAuth bool) (total int, notifications []entity.Notification,
	errorCode string, err error) {
	authQuery := bson.M{}
	if !skipAuth {
		// get auth query from auth service first
		authQuery, err = GetAuthService().BuildQueryByAuth("list_notification", x_auth_token)
		if err != nil {
			logrus.Errorf("get auth query by token [%v] error is %v", x_auth_token, err)
			errorCode = COMMON_ERROR_INTERNAL
			return
		}
	}

	selector := generateQueryWithAuth(query, authQuery)
	notifications = []entity.Notification{}
	queryStruct := dao.QueryStruct{p.collectionName, selector, skip, limit, ""}
	total, err = dao.HandleQueryAll(&notifications, queryStruct)
	if err != nil {
		logrus.Errorf("query notifications by query [%v] error is %v", query, err)
		errorCode = NOTIFICATION_ERROR_QUERY
	}
	return
}
