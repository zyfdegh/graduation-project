package documents

import (
	"github.com/Sirupsen/logrus"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"encoding/json"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
)

var notificationCollection = "notification"

func (p Resource) NotificationWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/notification")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of notification")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.POST("/").To(p.NotificationCreateHandler).
		Doc("Store a notification").
		Operation("NotificationCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "Notification body in json format").DataType("string")))

	ws.Route(ws.PUT("/" + paramID).To(p.NotificationUpdateHandler).
		Doc("Update a notification by its storage identifier").
		Operation("NotificationUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "Notification body in json format").DataType("string")))

	ws.Route(ws.GET("/" + paramID).To(p.NotificationListHandler).
		Doc("Return a notification by its storage identifier").
		Operation("NotificationListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))

	ws.Route(ws.DELETE("/" + paramID).To(p.NotificationDeleteHandler).
		Doc("Detele a notification by its storage identifier").
		Operation("NotificationDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	return ws

}

func (p *Resource) NotificationCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("NotificationCreateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	notification := entity.Notification{}
	err = json.NewDecoder(req.Request.Body).Decode(&notification)
	if err != nil {
		logrus.Errorf("convert body to notification failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}
	
	newNotification, code, err := services.GetNotificationService().Create(notification, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	res := response.QueryStruct{Success: true, Data: newNotification}
	resp.WriteEntity(res)
	return
	
}

func (p *Resource) NotificationUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("NotificationUpdateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	objectId := req.PathParameter(ParamID)
	// Stub an notification to be populated from the body
	notification := entity.Notification{}
	
	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&notification)
	if err != nil {
		logrus.Errorf("convert body to notification failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}
	
	created, code, err := services.GetNotificationService().UpdateById(objectId, notification, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	p.successUpdate(objectId, created, req, resp)
	
	
}

func (p *Resource) NotificationListHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("NotificationListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	objectId := req.PathParameter(ParamID)
	
	notification, code, err := services.GetNotificationService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("notification is %v", notification)
	
	res := response.QueryStruct{Success: true, Data: notification}
	resp.WriteEntity(res)
	return
}

func (p *Resource) NotificationDeleteHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("NotificationDeleteHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	
	objectId := req.PathParameter(ParamID)
	code, err = services.GetNotificationService().DeleteById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Write success response
	response.WriteSuccess(resp)
	return

}
