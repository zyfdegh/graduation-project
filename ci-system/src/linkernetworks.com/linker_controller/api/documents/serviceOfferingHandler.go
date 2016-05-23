package documents

import (
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"linkernetworks.com/linker_common_lib/rest/response"
	"linkernetworks.com/linker_controller/services"
)

func (p Resource) ServiceOfferingWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1/serviceOfferings")
	ws.Consumes("application/json")
	ws.Produces(restful.MIME_JSON)

	id := ws.PathParameter(ParamID, "Storage identifier of service offering")
	paramID := "{" + ParamID + "}"

	ws.Route(ws.GET("/" + paramID).To(p.SOGetHandler).
		Doc("Return a service offering by its storage identifier").
		Operation("SOGetHandler").
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	ws.Route(ws.GET("/" + paramID + "/containerinfo").To(p.SOContainerInfoHandler).
		Doc("Return an app instance container info").
		Operation("SOContainerInfoHandler").
		Param(ws.QueryParameter("serviceGroupInstanceId", "service group instance id")).
		Param(ws.QueryParameter("appId", "App path in the group")).
		Param(ws.HeaderParameter("X-Auth-Token", "A valid authentication token")).
		Param(id))

	return ws
}

// SOGetHandler parses the http request and return a service offering.
// Usage :
//		GET /v1/serviceOfferings/{ParamID}
// Params :
//		ParamID : Storage identifier of service offering
// If successful,response code will be set to 201.
func (p *Resource) SOGetHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SOGetHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	objectId := req.PathParameter(ParamID)

	so, code, err := services.GetSoService().QueryById(objectId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	logrus.Debugf("so is %v", so)

	res := response.QueryStruct{Success: true, Data: so}
	resp.WriteEntity(res)
	return
}

func (p *Resource) SOContainerInfoHandler(req *restful.Request, resp *restful.Response) {
	logrus.Infof("SOContainerInfoHandler is called!")
	x_auth_token := req.HeaderParameter("X-Auth-Token")
	code, err := services.TokenValidation(x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}

	soId := req.PathParameter(ParamID)
	sgiId := req.QueryParameter("serviceGroupInstanceId")
	appId := req.QueryParameter("appId")

	app, code, err := services.GetSoService().GetSOContainerInfo(soId, sgiId, appId, x_auth_token)
	if err != nil {
		response.WriteStatusError(code, err, resp)
		return
	}
	response.WriteResponse(app, resp)
}
