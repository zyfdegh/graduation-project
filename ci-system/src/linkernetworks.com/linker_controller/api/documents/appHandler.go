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

func (p Resource) AppWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/apps")
	ws.Consumes("*/*")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of the app")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/").To(p.AppsListHandler).
		Doc("Returns all app items").
		Operation("AppsListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.QueryParameter("count", "Count total items and return the result in X-Object-Count header").DataType("boolean")).
		Param(ws.QueryParameter("skip", "Number of items to skip in the result set, default=0")).
		Param(ws.QueryParameter("limit", "Maximum number of items in the result set, default=0")))

	ws.Route(ws.GET("/" + paramID).To(p.AppsGetHandler).
		Doc("Return an app by its storage identifier").
		Operation("AppsListHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.QueryParameter("fields", "Comma separated list of field names")))

	ws.Route(ws.GET("/operations/" + paramID).To(p.AppAuthOperationHandler).
		Doc("Return authorized operations of an app by its storage identifier").
		Operation("AppAuthOperationHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.POST("/").To(p.AppsCreateHandler).
		Doc("Store an app").
		Operation("AppsCreateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(ws.BodyParameter("body", "App body in json format,for example {\"cmd\":\"...\",   \"id\":\"...\",  \"cpus\":\"...\",   \"mem\":\"...\",  \"instances\":\"...\",   \"env\":\"...\",  \"excutor\":\"...\",  \"constraints\":[],  \"container\":{\"docker\":{}}}").DataType("string")))

	ws.Route(ws.PUT("/" + paramID).To(p.AppsUpdateHandler).
		Doc("Update an app by storage identifier").
		Operation("AppsUpdateHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id).
		Param(ws.BodyParameter("body", "App body in json format,for example {\"cmd\":\"...\",   \"id\":\"...\",   \"cpus\":\"...\",   \"mem\":\"...\",   \"instances\":\"...\",   \"env\":\"...\",   \"excutor\":\"...\",   \"constraints\":[],  \"container\":{\"docker\":{}}}").DataType("string")))

	ws.Route(ws.DELETE("/" + paramID).To(p.AppsDeleteHandler).
		Doc("Delete an app by its storage identifier").
		Operation("AppsDeleteHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	return ws
}

// AppsListHandler parses the http request and return the app items.
// Usage :
//		GET /v1/apps
// Params :
// If successful,response code will be set to 201.
func (p *Resource) AppsListHandler(req *restful.Request,
	resp *restful.Response) {
	// logrus.Infof("AppsListHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	var skip int = queryIntParam(req, "skip", 0)
	var limit int = queryIntParam(req, "limit", 0)

	total, apps, code, err := services.GetAppService().QueryAll(skip, limit, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	res := response.QueryStruct{Success: true, Data: apps}
	if c, _ := strconv.ParseBool(req.QueryParameter("count")); c {
		res.Count = total
		resp.AddHeader("X-Object-Count", strconv.Itoa(total))
	}
	resp.WriteEntity(res)
	return
}

// AppsGetHandler parses the http request and return the app item by objectid.
// Usage :
//		GET /v1/apps/{ParamID}
// Params :
//		ParamID : Storage identifier of the app
// If successful,response code will be set to 201.
func (p *Resource) AppsGetHandler(req *restful.Request,
	resp *restful.Response) {
	// logrus.Infof("AppsGetHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	app, code, err := services.GetAppService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("app is %v", app)

	res := response.QueryStruct{Success: true, Data: app}
	resp.WriteEntity(res)
	return
}

// AppAuthOperationHandler parses the http request and return authorized
//	operations of an app.
// Usage :
//		GET /v1/apps/operations/{ParamID}
// Params :
//		ParamID : Storage identifier of the app
// If successful,response code will be set to 201.
func (p *Resource) AppAuthOperationHandler(req *restful.Request,
	resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	operations, code, err := services.GetAppService().GetOperationById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.QueryStruct{Success: true, Data: operations}
	resp.WriteEntity(res)
	return
}

// AppsDeleteHandler parses the http request and delete the app items.
// Usage :
//		DELETE /v1/apps/{ParamID}
// Params :
//		ParamID : Storage identifier of the app
// If successful,response code will be set to 201.
func (p *Resource) AppsDeleteHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	code, err = services.GetAppService().DeleteById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	// Write success response
	response.WriteSuccess(resp)
	return
}

// AppsCreateHandler parses the http request and store an app.
// Usage :
//		POST /v1/apps
// If successful,response code will be set to 201.
func (p *Resource) AppsCreateHandler(req *restful.Request, resp *restful.Response) {
	// logrus.Infof("AppsCreateHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	// Stub an app to be populated from the body
	app := entity.App{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&app)
	if err != nil {
		logrus.Errorf("convert body to app failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	newApp, code, err := services.GetAppService().Create(app, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	res := response.Response{Success: true, Data: newApp}
	resp.WriteEntity(res)
	return
}

// AppsUpdateHandler parses the http request and update an app.
// Usage :
//		PUT /v1/apps/{ParamID}
// Params :
//		ParamID : Storage identifier of the app
// If successful,response code will be set to 201.
func (p *Resource) AppsUpdateHandler(req *restful.Request, resp *restful.Response) {
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	// Stub an app to be populated from the body
	app := entity.App{}

	// Populate the user data
	err = json.NewDecoder(req.Request.Body).Decode(&app)
	if err != nil {
		logrus.Errorf("convert body to app failed, error is %v", err)
		response.WriteStatusError(services.COMMON_ERROR_INVALIDATE, err, resp)
		return
	}

	created, code, err := services.GetAppService().UpdateById(objectId, app, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	p.successUpdate(objectId, created, req, resp)
}
