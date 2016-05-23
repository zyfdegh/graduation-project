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

var acCollection = "app_container_package"

func (p Resource) AppPackageWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/appConfigs")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of the app configuration")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.ACPsListHandler).
		Doc("Returns all app configuration items").
		Operation("ACPsListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")).
		Param(ws.QueryParameter("service_group_id", "The name of service group")).
		Param(ws.QueryParameter("app_container_id", "The name of app container")).
		Param(ws.QueryParameter("sort", "Comma separated list of field names to sort")))

	ws.Route(ws.GET("/" + paramID).To(p.ACPsGetHandler).
		Doc("Return an app configuration by its storage identifier ").
		Operation("ACPGettHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))

	ws.Route(ws.POST("/").To(p.ACPCreateHandler).
		Doc("Store an app configuration").
		Operation("ACPCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "App configuration body in json format,for example {\"name\":\"...\",  \"preconditions\":[{\"condition\":\"...\"}],  \"steps\":[]}").DataType("string")))

	ws.Route(ws.PUT("/" + paramID).To(p.ACPUpdateHandler).
		Doc("Update an app configuration by its storage identifier").
		Operation("ACPUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "App configuration body in json format,for example {\"name\":\"...\",  \"preconditions\":[{\"condition\":\"...\"}],  \"steps\":[]}").DataType("string").DataType("string")))

	ws.Route(ws.DELETE("/" + paramID).To(p.ACPsDeleteOneHandler).
		Doc("Delete an app configuration by its storage identifier").
		Operation("ACPsDeleteOneHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.DELETE("/").To(p.ACPsDeleteByQueryHandler).
		Doc("Delete app configurations if query present, otherwise remove the entire configurations").
		Operation("ACPsDeleteByQueryHandler").
		Param(ws.QueryParameter("service_group_id", "The name of service group")).
		Param(ws.QueryParameter("app_container_id", "The name of app container")))
	return ws
}

// ACPsListHandler parses the http request and return app configuration items.
// Usage :
//		GET /v1/appConfigs
// Params :
// If successful,response code will be set to 201.
func (p *Resource) ACPsListHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)
	var sgId string = req.QueryParameter("service_group_id")
	var appConId string = req.QueryParameter("app_container_id")

	total, acps, code, err := services.GetAcpService().QueryAllByName(sgId, appConId, skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: acps}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

// ACPsGetHandler parses the http request and return app configuration by its objectId.
// Usage :
//		GET /v1/appConfigs/{ParamID}
// Params :
//		ParamID : Storage identifier of the app configuration
// If successful,response code will be set to 201.
func (p *Resource) ACPsGetHandler(req *restful.Request, resp *restful.Response) {
	// logrus.Infof("ACPsGetHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	acp, code, err := services.GetAcpService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("acp is %v", acp)

	res := response.QueryStruct{Success: true, Data: acp}
	resp.WriteEntity(res)
	return
}

// ACPsDeleteOneHandler parses the http request and delete app configurations.
// Usage :
//		DELETE /v1/appConfigs/{ParamID}
// Params :
//		ParamID : Storage identifier of the app configuration
// If successful,response code will be set to 201.
func (p *Resource) ACPsDeleteOneHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	code, err = services.GetAcpService().DeleteById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Write success response
	response.WriteSuccess(resp)
	return
}

// ACPCreateHandler parses the http request and store an app configuration
// Usage :
//		POST /v1/appConfigs
// If successful,response code will be set to 201.
func (p *Resource) ACPCreateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ACPCreateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Stub an app to be populated from the body
	acp := entity.AppContainerPackage{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&acp)
	if err != nil {
		logrus.Errorf("convert body to acp failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newAcp, code, err := services.GetAcpService().Create(acp, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.Response{Success: true, Data: newAcp}
	resp.WriteEntity(res)
	return

}

// ACPUpdateHandler parses the http request and update an app configuration.
// Usage :
//		PUT /v1/appConfigs/{ParamID}
// Params :
//		ParamID : Storage identifier of the app configuration
// If successful,response code will be set to 201.
func (p *Resource) ACPUpdateHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("ACPUpdateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	// Stub an acp to be populated from the body
	acp := entity.AppContainerPackage{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&acp)
	if err != nil {
		logrus.Errorf("convert body to acp failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	created, code, err := services.GetAcpService().UpdateById(objectId, acp, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}

// ACPsDeleteHandler parses the http request and delete app configurations.
// Usage :
//		DELETE /v1/appConfigs
// Params :
// If successful,response code will be set to 201.
func (p *Resource) ACPsDeleteByQueryHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var sgId string = req.QueryParameter("service_group_id")
	var appId string = req.QueryParameter("app_container_id")

	code, err = services.GetAcpService().DeleteBySgOrApp(sgId, appId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Write success response
	response.WriteSuccess(resp)
	return
}
