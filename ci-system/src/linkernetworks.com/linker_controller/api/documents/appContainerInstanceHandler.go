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

func (p Resource) AppInstanceWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/appInstances")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of the app instance")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.ACIsListHandler).
		Doc("Return all app instance items").
		Operation("ACIsListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Counts total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("query", "Query in json format")).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.GET("/" + paramID).To(p.ACIsGetHandler).
		Doc("Return an app instance by its storage identifier").
		Operation("ACIsGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))
	ws.Route(ws.GET("/" + paramID + "/allocate").To(p.ACIsResourceAllocateHandler).
		Doc("Return an app instance by its storage identifier").
		Operation("ACIsResourceAllocateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("resource", "the type of the resource you want to allocate, supported ipAddressResource:[][]")))

	ws.Route(ws.POST("/").To(p.ACICreateHandler).
		Doc("Store an app instance").
		Operation("ACICreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "App instance body in json format")))

	ws.Route(ws.PUT("/" + paramID).To(p.ACIUpdateHandler).
		Doc("Update an app instance by its storage identifier").
		Operation("ACIUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "App instance body in json format")))

	ws.Route(ws.DELETE("/" + paramID).To(p.ACIsDeleteHandler).
		Doc("Delete an app instance by its storage identifier").
		Operation("ACIsDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/" + paramID + "/steps").To(p.ACIStepsQueryHandler).
		Doc("Return an app instance configuration steps by its storage identifier").
		Operation("ACIStepsQueryHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	return ws
}

// ACIsListHandler parses the http request and return app instance items.
// Usage :
//		GET /v1/appInstances
// Params :
// If successful,response code will be set to 201.
func (p *Resource) ACIsListHandler(req *restful.Request, resp *restful.Response) {
	// logrus.Infof("AppsListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)

	total, acis, code, err := services.GetAciService().QueryAll(skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: acis}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

// ACIsGetHandler parses the http request and return app instance items.
// Usage :
//		GET /v1/appInstances/{ParamID}
// Params :
//		ParamID : Storage identifier of the app instance
// If successful,response code will be set to 201.
func (p *Resource) ACIsGetHandler(req *restful.Request, resp *restful.Response) {
	// logrus.Infof("ACIsListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	aci, code, err := services.GetAciService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("aci is %v", aci)

	res := response.QueryStruct{Success: true, Data: aci}
	resp.WriteEntity(res)
	return
}

// ACIsResourceAllocateHandler will assign a resource to the aci by resource type
// Usage :
//		GET /v1/appInstances/{ParamID}?resource=ipAddressResource
// Params:
//		ParamId: Storage identifier of the app instance
//		resource: The type of resource you want to allocate
// If successful, response code will be set to 201.
func (p *Resource) ACIsResourceAllocateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ACIsResourceAllocateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	allocateRegx := req.QueryParameter("resource")

	allocatedValue, code, err := services.GetAciService().AllocateResource(objectId,
		allocateRegx, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("allocated resource value is %v", allocatedValue)

	res := response.QueryStruct{Success: true, Data: allocatedValue}
	resp.WriteEntity(res)
	return
}

// ACICreateHandler parses the http request and store an app instance.
// Usage :
//		POST /v1/appInstances
// If successful,response code will be set to 201.
func (p *Resource) ACICreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ACICreateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Stub an aci to be populated from the body
	aci := entity.AppContainerInstance{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&aci)
	if err != nil {
		logrus.Errorf("convert body to aci failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newAci, code, err := services.GetAciService().Create(aci, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.Response{Success: true, Data: newAci}
	resp.WriteEntity(res)
	return
}

// ACIsDeleteHandler parses the http request and delete an app instance.
// Usage :
//		DELETE /v1/appInstances/{ParamID}
// Params :
//		ParamID : Storage identifier of the app instance
// If successful,response code will be set to 201.
func (p *Resource) ACIsDeleteHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ACIsDeleteHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	code, err = services.GetAciService().DeleteById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Write success response
	response.WriteSuccess(resp)
	return
}

// ACIUpdateHandler parses the http request and update an app instance.
// Usage :
//		PUT /v1/appInstances/{ParamID}
// Params :
//		ParamID : Storage identifier of the app instance
// If successful,response code will be set to 201.
func (p *Resource) ACIUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ACIUpdateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	// Stub an app to be populated from the body
	aci := entity.AppContainerInstance{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&aci)
	if err != nil {
		logrus.Errorf("convert body to aci failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	created, code, err := services.GetAciService().UpdateById(objectId, aci, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}

// ACIStepsQueryHandler parses the http request and return
// an app instance configuration steps.
// Usage :
//		GET /v1/appInstances/{ParamID}/steps
// Params :
//		ParamID : Storage identifier of the app instance
// If successful,response code will be set to 201.
func (p *Resource) ACIStepsQueryHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ACIStepsQueryHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)
	callFrom := "query"
	steps, err := services.GetConfigStepsByAciId(objectId, callFrom, x_auth_token)
	if err != nil {
		logrus.Errorf("get aci [%v] configuration step failed, error is %v", objectId, err)
		response.WriteError(err, resp)
		return
	}
	response.WriteResponse(steps, resp)
}
