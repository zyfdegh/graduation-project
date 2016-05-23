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

func (p Resource) ServiceGroupWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/serviceGroups")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of service group")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.SGsListHandler).
		Doc("Return all service group items belongs to the user.").
		Operation("SGsListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		Param(ws.QueryParameter("name", "The name of services wanted to query")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))

	ws.Route(ws.GET("/published").To(p.SGsListAllHandler).
		Doc("Return all published service group items").
		Operation("SGsListAllHandler").
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-count header").DataType("boolean")).
		Param(ws.QueryParameter("name", "The name of services wanted to query")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))

	ws.Route(ws.GET("/" + paramID).To(p.SGsGetHandler).
		Doc("Return a service group by its storage identifier").
		Operation("SGsGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))

	ws.Route(ws.GET("/operations/" + paramID).To(p.SGAuthOperationHandler).
		Doc("Return authorized operations of a service group by its storage identifier").
		Operation("SGAuthOperationHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.POST("/").To(p.SGCreateHandler).
		Doc("Store a service group").
		Operation("SGCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "Service group body in json format,for example {\"id\":\"...\", \"apps\":[], \"dependenceies\":[], \"group\": [{\"id\":\"...\",  \"dependenceies\":[],  \"apps\":[]}]}").DataType("string")))

	ws.Route(ws.PUT("/publish/" + paramID).To(p.SGPublishHandler).
		Doc("Publish a service group by its storage identifier").
		Operation("SGPublishHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.PUT("/unpublish/" + paramID).To(p.SGUnpublishHandler).
		Doc("Unpublish a service group by its storage identifier").
		Operation("SGUnpublishHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.PUT("/submit/" + paramID).To(p.SGSubmitHandler).
		Doc("submit a service group by its storage identifier").
		Operation("SGSubmitHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.PUT("/" + paramID).To(p.SGUpdateHandler).
		Doc("Update a service group by its storage identifier").
		Operation("SGUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "Service group body in json format,for example {\"id\":\"...\", \"apps\":[], \"dependenceies\":[], \"group\": [{\"id\":\"...\", \"dependenceies\":[], \"apps\":[]}]}").DataType("string")))

	ws.Route(ws.DELETE("/" + paramID).To(p.SGsDeleteHandler).
		Doc("Detele a service group by its storage identifier").
		Operation("SGsDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	return ws
}

// SGsListAllHandler parses the http request and return all published
// service group items.
// Usage :
//		GET /v1/serviceGroups/published
// If successful,response code will be set to 201.
func (p *Resource) SGsListAllHandler(req *restful.Request, resp *restful.Response) {
	// no need to check token
	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var name string = req.QueryParameter("name")

	total, sgs, code, err := services.GetSgService().QueryAllPublishedByName(name, skip, limit, "")
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: sgs}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

// SGsListHandler parses the http request and return service group items.
// Usage :
//		GET /v1/serviceGroups
// Params :
// If successful,response code will be set to 201.
func (p *Resource) SGsListHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var name string = req.QueryParameter("name")

	total, sgs, code, err := services.GetSgService().QueryAllByName(name, skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: sgs}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

// SGsGetHandler parses the http request and return service group by it's objectId.
// Usage :
//		GET /v1/serviceGroups/{ParamID}
// Params :
//		ParamID : Storage identifier of the service group
// If successful,response code will be set to 201.
func (p *Resource) SGsGetHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	sg, code, err := services.GetSgService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("sg is %v", sg)

	res := response.QueryStruct{Success: true, Data: sg}
	resp.WriteEntity(res)
	return
}

// SGAuthOperationHandler parses the http request and return authorized
// operations of a service group.
// Usage :
//		GET /v1/serviceGroups/operations/{ParamID}
// Params :
//		ParamID : Storage identifier of the service group
// If successful,response code will be set to 201.
func (p *Resource) SGAuthOperationHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	operations, code, err := services.GetSgService().GetOperationById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: operations}
	resp.WriteEntity(res)
	return
}

// SGsDeleteHandler parses the http request and detele a service group.
//	Usage :
//		DELETE /v1/serviceGroups/{ParamID}
// Params :
//		ParamID : Storage identifier of the service group
// If successful,response code will be set to 201.
func (p *Resource) SGsDeleteHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	code, err = services.GetSgService().DeleteById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Write success response
	response.WriteSuccess(resp)
	return
}

// SGCreateHandler parses the http request and store a service group.
// Usage :
//		POST /v1/serviceGroups
// If successful,response code will be set to 201.
func (p *Resource) SGCreateHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Stub an app to be populated from the body
	sg := entity.ServiceGroup{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&sg)
	if err != nil {
		logrus.Errorf("convert body to sg failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newSg, code, err := services.GetSgService().Create(sg, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: newSg}
	resp.WriteEntity(res)
	return
}

// SGUpdateHandler parses the http request and update the service group.
// Usage :
//		PUT /v1/serviceGroups/{ParamID}
// Params :
//		ParamID : Storage identifier of the service group
// If successful,response code will be set to 201.
func (p *Resource) SGUpdateHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	// Stub an sg to be populated from the body
	sg := entity.ServiceGroup{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&sg)
	if err != nil {
		logrus.Errorf("convert body to sg failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	created, code, err := services.GetSgService().UpdateById(objectId, sg, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}

// SGPublishHandler parses the http request and publish a service group.
// Usage :
//		POST /v1/serviceGroups/publish/{ParamID}
// Params :
//		ParamID : Storage identifier of the service group
// If successful,response code will be set to 201.
func (p *Resource) SGPublishHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SGPublishHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	created, code, err := services.GetSgService().UpdateStateById(objectId,
		services.SG_STATUS_PUBLISHED, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}

// SGUnpublishHandler parses the http request and unpublish a service group.
// Usage :
//		PUT /v1/serviceGroups/publish/{ParamID}
// Params :
//		ParamID : Storage identifier of the service group
// If successful,response code will be set to 201.
func (p *Resource) SGUnpublishHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SGUnpublishHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	created, code, err := services.GetSgService().UpdateStateById(objectId,
		services.SG_STATUS_UNPUBLISHED, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}

// SGSubmitHandler parses the http request and submit a service group.
// Usage :
//		PUT /v1/serviceGroups/submit/{ParamID}
// Params :
//		ParamID : Storage identifier of the service group
// If successful,response code will be set to 201.
func (p *Resource) SGSubmitHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SGSubmitHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	created, code, err := services.GetSgService().UpdateStateById(objectId,
		services.SG_STATUS_VERIFYING, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}
