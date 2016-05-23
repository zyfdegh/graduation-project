package documents

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/persistence/entity"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
	"strconv"
)

func (p Resource) ServiceGroupOrderWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/serviceGroupOrders")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of service order")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.SGOListHandler).
		Doc("Return all service orders").
		Operation("SGOListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.GET("/" + paramID).To(p.SGOGetHandler).
		Doc("Return a service group order by its storage identifier").
		Operation("SGOGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/" + "unterminated").To(p.SGOListUnterminateHandler).
		Doc("Return all unterminate service orders").
		Operation("SGOListUnterminateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.POST("/").To(p.SGOCreateHandler).
		Doc("Creat a new order of service group").
		Operation("SGOCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "Service order in json format,for example {\"service_group_id\":\"...\",   \"parameters\": [{\"appId\":\"...\",  \"paramName\":\"...\",  \"paramValue\":\"...\"}]} ").DataType("string")))

	ws.Route(ws.DELETE("/" + paramID).To(p.SGODeleteHandler).
		Doc("Terminate a service order by its storage identifier").
		Operation("SGODeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/operations/" + paramID).To(p.SGOAuthOperationHandler).
		Doc("Return authorized operations of a service group order by its storage identifier").
		Operation("SGOAuthOperationHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/" + paramID + "/scaleInfo").To(p.SGOAppHandler).
		Doc("Return authorized operations of a service group order by its storage identifier").
		Operation("SGOAppHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("appId", "App path in the group")).
		Param(id))

	ws.Route(ws.PUT("/" + paramID + "/scaleApp").To(p.SGOAppScaleHandler).
		Doc("Update the number of app in service order by its storage identifier").
		Operation("SGOAppScaleHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("appId", "App path in the group")).
		Param(ws.QueryParameter("num", "New number of app")))

	return ws
}

func (p *Resource) SGOGetHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SGOGetHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")

	objectId := req.PathParameter(ParamID)
	sgo, errorCode, err := services.GetSgoService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(errorCode, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: sgo}
	resp.WriteEntity(res)
	return

}

// SGOCreateHandler parses the http request and creat a new order of service group.
// Usage :
//		POST /v1/serviceGroupOrders
// If successful,response code will be set to 201.
func (p *Resource) SGOCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SGOCreateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Stub an order to be populated from the body
	order := entity.ServiceGroupOrder{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&order)
	if err != nil {
		logrus.Errorf("convert body to order failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newOrder, code, err := services.GetSgoService().Create(order, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.Response{Success: true, Data: newOrder}
	resp.WriteEntity(res)
	return
}

// SGOListHandler parses the http request and return all service orders.
// Usage :
//		GET /v1/serviceGroupOrders
//If successful,response code will be set to 201.
func (p *Resource) SGOListHandler(req *restful.Request, resp *restful.Response) {
	// logrus.Infof("AppsListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)

	total, sgos, code, err := services.GetSgoService().QueryAll(skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: sgos}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

func (p *Resource) SGOListUnterminateHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)

	total, sgos, code, err := services.GetSgoService().QueryAllUnterminated(skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: sgos}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

// SGODeleteHandler parses the http request and terminate a service order.
// Usage :
//		DELETE /v1/serviceGroupOrders/{ParamID}
// Params :
//		ParamID : Storage identifier of service group orders
// If successful,response code will be set to 201.
func (p *Resource) SGODeleteHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SGODeleteHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	code, err = services.GetSgoService().DeleteById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Write success response
	response.WriteSuccess(resp)
	return
}

// SGOAppScaleHandler parses the http request and update
// the number of app in service order.
// Usage :
//		PUT /v1/serviceGroupOrders/{ParamID}/scaleApp
// Params :
//		ParamID : Storage identifier of service group orders
// If successful,response code will be set to 201.
func (p *Resource) SGOAppScaleHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SGOAppScaleHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	sgoId := req.PathParameter(ParamID)
	appId := req.QueryParameter("appId")
	numStr := req.QueryParameter("num")
	code, err = services.GetSgoService().ScaleAppByOrderId(sgoId, appId, numStr, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return

	}

	response.WriteSuccess(resp)
	return
}

func (p *Resource) SGOAuthOperationHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	operations, code, err := services.GetSgoService().GetOperationById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: operations}
	resp.WriteEntity(res)
	return
}

// SGOAppHandler parses the http request and return authorized
// operations of a service group order.
//Usage :
//		GET /v1/serviceGroupOrders/{ParamID}/scaleInfo
// Params :
//		ParamID : Storage identifier of service group orders
// If successful,response code will be set to 201.
func (d *Resource) SGOAppHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	appId := req.QueryParameter("appId")
	sgoId := req.PathParameter(ParamID)
	logrus.Debugf("get app %v info from oreder %v ", appId, sgoId)

	app, code, err := services.GetSgoService().GetAppInOrder(appId, sgoId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	response.WriteResponse(app, resp)
	return
}
