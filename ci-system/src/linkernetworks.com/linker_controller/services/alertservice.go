package services

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"linkernetworks.com/linker_common_lib/persistence/dao"
	"linkernetworks.com/linker_common_lib/persistence/entity"
)

const (
	ALERT_MESSAGES_STATUS_FIRING        = "firing"
	ALERT_MESSAGES_STATUS_REPAIRED      = "repaired"
	ALERT_MESSAGES_STATUS_REPAIR_FAILED = "failed"
	ALERT_MESSAGES_STATUS_IGNORED       = "ignored"
)

var (
	alertService       *AlertService = nil
	onceAlert          sync.Once
	ALERT_ERROR_CREATE string = "E11120"
)

type AlertService struct {
	collectionName string
}

func GetAlertService() *AlertService {
	onceAlert.Do(func() {
		logrus.Debugf("Once called from alertService ......................................")
		alertService = &AlertService{"alert"}
	})
	return alertService
}

func (p *AlertService) Create(alert *entity.AlertMessage) (newAlert *entity.AlertMessage,
	errorCode string, err error) {
	logrus.Infof("start to create alert [%v]", alert)

	// generate ObjectId
	alert.ObjectId = bson.NewObjectId()

	// set created_time and updated_time
	alert.TimeCreate = dao.GetCurrentTime()
	alert.TimeUpdate = alert.TimeCreate

	// insert bson to mongodb
	err = dao.HandleInsert(p.collectionName, alert)
	if err != nil {
		errorCode = ALERT_ERROR_CREATE
		logrus.Errorf("create alert [%v] to bson error is %v", alert, err)
		return
	}

	newAlert = alert
	
	return
}

func (p *AlertService) CheckRelatedRepairs(message *entity.AlertMessage) bool {
	if len(message.Alert) >0 {
		selector := make(bson.M)
		querymatch := make(bson.M)
		queryvalue := make(bson.M)
		queryvalue["labels.service_order_id"] = message.Alert[0].Labels.ServiceOrderId
		queryvalue["labels.app_container_id"] = message.Alert[0].Labels.AppContainerId
		querymatch["$elemMatch"] = queryvalue
	
		selector["alert"] = querymatch
		selector["status"] = ALERT_MESSAGES_STATUS_FIRING
	
		messages := []entity.AlertMessage{}
		queryStruct := dao.QueryStruct{p.collectionName, selector, 0, 0, "...."}
		total, err := dao.HandleQueryAll(&messages, queryStruct)
	
		logrus.Debugf("Query related alert with order_id: %s, app_container_id: %s, total: $d", message.Alert[0].Labels.ServiceOrderId, message.Alert[0].Labels.AppContainerId, total)
	
		if err != nil {
			logrus.Errorf("Query alert failed, error is %v", err)
			return false
		}
		if total > 0 {
			return false
		} else {
			return true
		}
	} else {
		return true
	}
	
}

func (p *AlertService) NotifyRepairResult(alertId, result string) {
	logrus.Infof("Alert id is %s, result is %s", alertId, result)
	selector := bson.M{}
	selector["_id"] = bson.ObjectIdHex(alertId)
	
	change := bson.M{}
	change["status"] = result
	err := dao.HandlePartialUpdateByQuery(p.collectionName, selector, change)
	if err != nil {
		logrus.Errorf("Update alert by id: %s failed, error is %v", alertId, err)
	}
}
