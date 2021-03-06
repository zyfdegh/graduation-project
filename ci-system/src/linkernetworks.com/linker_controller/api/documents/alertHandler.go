package documents

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"

	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"

	"gopkg.in/mgo.v2/bson"
)

func (p Resource) AlertWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/alert")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/").To(p.AlertHandler).
		Doc("Receive an alert").
		Operation("AlertHandler").
		Param(ws.BodyParameter("body", "Alert Body in JSON, normally generated by prometheus.").DataType("string")))

	return ws
}

// AlertHandler parses the http request and store an alert.
// Usage :
//		POST /v1/alert
// If successful,response code will be set to 201.
func (p *Resource) AlertHandler(req *restful.Request, resp *restful.Response) {
	// Read a document from request
	document := bson.M{}
	// Handle JSON parsing manually here, instead of relying on go-restful's
	// req.ReadEntity. This is because ReadEntity currently parses JSON with
	// UseNumber() which turns all numbers into strings. See:
	// https://github.com/emicklei/mora/pull/31
	decoder := json.NewDecoder(req.Request.Body)
	err := decoder.Decode(&document)
	if err != nil {
		logrus.Errorf("decode alert err is %v", err)
		response.WriteError(err, resp)
		return
	}

	var alert *entity.AlertMessage
	alert = new(entity.AlertMessage)
	alertout, err := json.Marshal(document)
	logrus.Infoln("request body is: %s", string(alertout))

	if err != nil {
		logrus.Errorf("marshal alert err is %v", err)
		response.WriteError(err, resp)
		return
	}
	err = json.Unmarshal(alertout, &alert)
	if err != nil {
		logrus.Errorf("unmarshal alert err is %v", err)
		response.WriteError(err, resp)
		return
	}

	// Check if related instance is repairing now. 
	result := services.GetAlertService().CheckRelatedRepairs(alert)
	
	if !result {
		// Set the status to ignored for this alert.
		alert.Status = services.ALERT_MESSAGES_STATUS_IGNORED
	}
//	alertnew, _ := json.Marshal(alert)
//	logrus.Infoln("alert body is: %s", string(alertnew))

	newAlert, code, err := services.GetAlertService().Create(alert)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	if result {
		// start to repair
		serviceGroupId := newAlert.Alert[0].Labels.ServiceGroupId
		appCointainerId := newAlert.Alert[0].Labels.AppContainerId
		serviceGroupInsanceId := newAlert.Alert[0].Labels.ServiceGroupInstanceId
		orderId := newAlert.Alert[0].Labels.ServiceOrderId 
		alertId := newAlert.ObjectId.Hex()
		alertName := newAlert.Alert[0].Labels.AlertName
		alertValue := newAlert.Alert[0].Payload.Value
		errorCode, err := services.GetRepairPolicyService().AnalyzeAlert(serviceGroupId, appCointainerId, serviceGroupInsanceId, orderId, alertId, alertName, alertValue)
		if err != nil {
			logrus.Errorf("Failed to call analyze errorCode is %s, err is %v \n", errorCode, err)
		}
	}

	res := response.Response{Success: true, Data: newAlert}
	resp.WriteEntity(res)
	return
}
